package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"ariga.io/atlas/sql/migrate"
)

const (
	// defaultSourceDir 是 API 示例项目的默认来源目录。
	defaultSourceDir = "examples"
	// defaultTargetDir 是 API 模板文件的默认目标目录。
	defaultTargetDir = "templates/api"
	// exampleModule 是 examples 中需要替换的固定模块路径。
	exampleModule = "github.com/teamsillybees/initra/examples"
	// frameworkModule 是生成项目依赖的 initra 根模块路径。
	frameworkModule = "github.com/teamsillybees/initra"
)

// frameworkRequirePattern 匹配 examples go.mod 中的 initra 依赖版本行。
var frameworkRequirePattern = regexp.MustCompile(`(?m)^\tgithub\.com/teamsillybees/initra v[^\r\n]+$`)

var adminPasswordHashPattern = regexp.MustCompile(`(?m)^(\s*1000000000001,\n\s*'admin',\n\s*)'\$2[aby]\$[^']+'(,?)$`)

// syncOptions 描述本次 API 模板同步的输入和行为开关。
type syncOptions struct {
	repoRoot string
	source   string
	target   string
	dryRun   bool
	check    bool
	delete   bool
}

// action 记录一次同步动作，用于 dry-run 和执行结果输出。
type action struct {
	kind string
	path string
}

// main 执行 examples 到 templates/api 的同步。
func main() {
	opts, err := parseOptions()
	if err != nil {
		fatal(err)
	}

	actions, err := syncTemplates(opts)
	if err != nil {
		fatal(err)
	}

	if err := reportSyncActions(os.Stdout, opts, actions); err != nil {
		fatal(err)
	}
}

// parseOptions 解析命令行参数并定位仓库根目录。
func parseOptions() (syncOptions, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return syncOptions{}, err
	}
	repoRoot, err := findRepoRoot(cwd)
	if err != nil {
		return syncOptions{}, err
	}

	opts := syncOptions{repoRoot: repoRoot, delete: true}
	flag.StringVar(&opts.source, "source", defaultSourceDir, "example project path, relative to repo root")
	flag.StringVar(&opts.target, "target", defaultTargetDir, "template path, relative to repo root")
	flag.BoolVar(&opts.dryRun, "dry-run", false, "print pending changes without writing files")
	flag.BoolVar(&opts.check, "check", false, "check template drift without writing; exit non-zero when changes are pending")
	flag.BoolVar(&opts.delete, "delete", true, "delete stale template files whose source files no longer exist")
	flag.Parse()

	opts.source = filepath.Join(opts.repoRoot, filepath.FromSlash(opts.source))
	opts.target = filepath.Join(opts.repoRoot, filepath.FromSlash(opts.target))
	return opts, nil
}

// findRepoRoot 从当前目录向上查找 initra 仓库根目录。
func findRepoRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		content, readErr := os.ReadFile(filepath.Join(dir, "go.mod"))
		if readErr == nil && bytes.Contains(content, []byte("module github.com/teamsillybees/initra")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("cannot find initra repo root from %s", start)
		}
		dir = parent
	}
}

// syncTemplates 将源文件转换为 .tmpl 文件并写入目标目录。
func syncTemplates(opts syncOptions) ([]action, error) {
	if err := ensureDir(opts.source, "source"); err != nil {
		return nil, err
	}
	if !opts.dryRun && !opts.check {
		if err := os.MkdirAll(opts.target, 0o755); err != nil {
			return nil, err
		}
	}

	wanted := make(map[string]struct{})
	var actions []action

	err := filepath.WalkDir(opts.source, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(opts.source, filePath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if shouldSkipSource(rel, entry) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}

		templateRel := rel + ".tmpl"
		wanted[templateRel] = struct{}{}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		var rendered string
		if rel == "db/migrations/atlas.sum" {
			rendered, err = templateMigrationHash(opts.source)
		} else {
			rendered, err = transformTemplateContent(rel, normalizeLineEndingsString(string(content)))
		}
		if err != nil {
			return err
		}
		if err := validateTemplate(templateRel, rendered); err != nil {
			return err
		}

		targetPath := filepath.Join(opts.target, filepath.FromSlash(templateRel))
		changed, err := fileContentChanged(targetPath, []byte(rendered))
		if err != nil {
			return err
		}
		if !changed {
			return nil
		}
		kind := changeKind(targetPath)
		if opts.dryRun || opts.check {
			actions = append(actions, action{kind: kind, path: filepath.ToSlash(filepath.Join(defaultTargetDir, templateRel))})
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(targetPath, []byte(rendered), 0o644); err != nil {
			return err
		}
		actions = append(actions, action{kind: kind, path: filepath.ToSlash(filepath.Join(defaultTargetDir, templateRel))})
		return nil
	})
	if err != nil {
		return nil, err
	}

	if opts.delete {
		deleteActions, err := deleteStaleTemplates(opts, wanted)
		if err != nil {
			return nil, err
		}
		actions = append(actions, deleteActions...)
	}
	return actions, nil
}

// ensureDir 确认路径存在且为目录。
func ensureDir(dir string, label string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("%s dir %s: %w", label, dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s path %s is not a directory", label, dir)
	}
	return nil
}

// shouldSkipSource 判断源文件或目录是否不应进入 API 模板。
func shouldSkipSource(rel string, entry fs.DirEntry) bool {
	base := path.Base(rel)
	if base == ".git" || base == ".DS_Store" || base == "var" || base == "tmp" {
		return true
	}
	if rel == "README.md" {
		return true
	}
	if entry.IsDir() {
		return shouldSkipDir(rel)
	}
	return strings.HasPrefix(rel, "internal/data/ent/")
}

// shouldSkipDir 判断目录是否属于需要跳过的 Ent 生成代码目录。
func shouldSkipDir(rel string) bool {
	return rel == "internal/data/ent" || strings.HasPrefix(rel, "internal/data/ent/")
}

// transformTemplateContent 将示例项目中的固定值替换为模板占位符。
func transformTemplateContent(rel string, content string) (string, error) {
	content = strings.ReplaceAll(content, exampleModule, "{{ .ModulePath }}")

	switch rel {
	case "go.mod":
		var err error
		content, err = replacePatternExactlyOnce(
			rel,
			content,
			frameworkRequirePattern,
			"\t{{ .FrameworkModule }} {{ .FrameworkVersion }}",
			"initra require",
		)
		if err != nil {
			return "", err
		}
		content, err = replaceExactlyOnce(
			rel,
			content,
			"replace github.com/teamsillybees/initra => ..\n\n",
			"",
		)
		if err != nil {
			return "", err
		}
	case "configs/config.yaml":
		var err error
		content, err = replaceTemplateValues(rel, content, map[string]string{
			"  name: initra":             `  name: {{ printf "%q" .AppName }}`,
			"  slug: initra":             "  slug: {{ .AppSlug }}",
			"  user: \"initra\"":         "  user: \"{{ .AppSlug }}\"",
			"  dbname: \"initra\"":       "  dbname: \"{{ .AppSlug }}\"",
			"  application_name: initra": "  application_name: {{ .AppSlug }}",
			"    issuer: initra":         "    issuer: {{ .AppSlug }}",
			"    secret: \"local-only-change-me-0123456789abcdef\"": "    secret: \"{{ .LocalJWTSecret }}\"",
		})
		if err != nil {
			return "", err
		}
	case "configs/config.local.yaml":
		var err error
		content, err = replaceTemplateValues(rel, content, map[string]string{
			"  dbname: \"initra\"":                                  "  dbname: \"{{ .AppSlug }}\"",
			"    secret: \"local-only-change-me-0123456789abcdef\"": "    secret: \"{{ .LocalJWTSecret }}\"",
		})
		if err != nil {
			return "", err
		}
	case "configs/config.dev.yaml":
		var err error
		content, err = replaceTemplateValues(rel, content, map[string]string{
			"  dbname: \"initra_dev\"":                            "  dbname: \"{{ .AppSlug }}_dev\"",
			"    secret: \"dev-only-change-me-0123456789abcdef\"": "    secret: \"{{ .DevJWTSecret }}\"",
		})
		if err != nil {
			return "", err
		}
	case "configs/config.test.yaml":
		var err error
		content, err = replaceTemplateValues(rel, content, map[string]string{
			"  dbname: \"initra_test\"":                            "  dbname: \"{{ .AppSlug }}_test\"",
			"    secret: \"test-only-change-me-0123456789abcdef\"": "    secret: \"{{ .TestJWTSecret }}\"",
		})
		if err != nil {
			return "", err
		}
	case "docker-compose.yml":
		var err error
		content, err = replaceTemplateValues(rel, content, map[string]string{
			"name: initra":              "name: {{ .AppSlug }}",
			"      POSTGRES_DB: initra": "      POSTGRES_DB: {{ .AppSlug }}",
		})
		if err != nil {
			return "", err
		}
	case "db/atlas.hcl":
		var err error
		content, err = replaceTemplateValues(rel, content, map[string]string{
			`  default = "postgres://postgres:postgres@127.0.0.1:5432/initra?sslmode=disable"`:                  `  default = "postgres://postgres:postgres@127.0.0.1:5432/{{ .AppSlug }}?sslmode=disable"`,
			`  default = "postgres://initra:QRD5jmc9nex3qxw-kez@192.168.100.2:5432/initra_dev?sslmode=require"`: `  default = "postgres://{{ .AppSlug }}:QRD5jmc9nex3qxw-kez@192.168.100.2:5432/{{ .AppSlug }}_dev?sslmode=require"`,
			`  default = "postgres://postgres:postgres@127.0.0.1:5432/initra_test?sslmode=disable"`:             `  default = "postgres://postgres:postgres@127.0.0.1:5432/{{ .AppSlug }}_test?sslmode=disable"`,
			`  default = "postgres://postgres:postgres@127.0.0.1:5432/initra?sslmode=verify-full"`:              `  default = "postgres://postgres:postgres@127.0.0.1:5432/{{ .AppSlug }}?sslmode=verify-full"`,
		})
		if err != nil {
			return "", err
		}
	case "db/migrations/20260715000000_add_relationship_foreign_keys.sql":
		if count := strings.Count(strings.ToUpper(content), "FOREIGN KEY"); count != 5 {
			return "", fmt.Errorf("transform %s: physical foreign key count = %d, expected 5", rel, count)
		}
		content = "-- 兼容保留历史迁移版本号；新生成项目不建立物理外键，关系由应用事务校验。\n"
	case "db/seeds/001_seed_admin.sql":
		var err error
		content, err = replaceExactlyOnce(
			rel,
			content,
			"-- 默认管理员账号为 admin；示例仓库不提供初始密码明文。",
			"-- 默认管理员账号为 admin；初始密码由 initra new 成功后一次性输出，明文不会写入项目文件。",
		)
		if err != nil {
			return "", err
		}
		content, err = replacePatternExactlyOnce(
			rel,
			content,
			adminPasswordHashPattern,
			`${1}'{{ .AdminPasswordHash }}'${2}`,
			"admin bcrypt hash",
		)
		if err != nil {
			return "", err
		}
	}

	return content, nil
}

// templateMigrationHash 依据模板实际生成的 SQL 内容计算 Atlas 完整性文件。
// examples 保留已发布迁移，而新项目会把旧外键迁移渲染为 no-op，因此不能直接复制 examples 的 atlas.sum。
func templateMigrationHash(sourceDir string) (string, error) {
	migrationDir := filepath.Join(sourceDir, "db", "migrations")
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return "", fmt.Errorf("read migration dir %s: %w", migrationDir, err)
	}
	files := make([]migrate.File, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		filePath := filepath.Join(migrationDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("read migration %s: %w", filePath, err)
		}
		rel := path.Join("db/migrations", entry.Name())
		transformed, err := transformTemplateContent(rel, normalizeLineEndingsString(string(content)))
		if err != nil {
			return "", err
		}
		files = append(files, migrate.NewLocalFile(entry.Name(), []byte(transformed)))
	}
	hashFile, err := migrate.NewHashFile(files)
	if err != nil {
		return "", fmt.Errorf("hash template migrations: %w", err)
	}
	content, err := hashFile.MarshalText()
	if err != nil {
		return "", fmt.Errorf("marshal template migration hash: %w", err)
	}
	return string(content), nil
}

func replaceTemplateValues(rel string, content string, replacements map[string]string) (string, error) {
	for current, templateValue := range replacements {
		var err error
		content, err = replaceExactlyOnce(rel, content, current, templateValue)
		if err != nil {
			return "", err
		}
	}
	return content, nil
}

func replaceExactlyOnce(rel string, content string, current string, templateValue string) (string, error) {
	if count := strings.Count(content, current); count != 1 {
		return "", fmt.Errorf("transform %s: anchor %q matched %d times, expected exactly once", rel, current, count)
	}
	return strings.Replace(content, current, templateValue, 1), nil
}

func replacePatternExactlyOnce(
	rel string,
	content string,
	pattern *regexp.Regexp,
	templateValue string,
	label string,
) (string, error) {
	if count := len(pattern.FindAllStringIndex(content, -1)); count != 1 {
		return "", fmt.Errorf("transform %s: %s matched %d times, expected exactly once", rel, label, count)
	}
	return pattern.ReplaceAllString(content, templateValue), nil
}

// validateTemplate 确认生成后的内容仍是合法的 Go text/template。
func validateTemplate(rel string, content string) error {
	_, err := template.New(rel).Option("missingkey=error").Parse(content)
	if err != nil {
		return fmt.Errorf("parse generated template %s: %w", rel, err)
	}
	return nil
}

// fileContentChanged 判断目标文件内容是否需要更新。
func fileContentChanged(filePath string, content []byte) (bool, error) {
	current, err := os.ReadFile(filePath)
	if err == nil {
		return !bytes.Equal(normalizeLineEndings(current), normalizeLineEndings(content)), nil
	}
	if os.IsNotExist(err) {
		return true, nil
	}
	return false, err
}

func normalizeLineEndings(content []byte) []byte {
	normalized := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	return bytes.ReplaceAll(normalized, []byte("\r"), []byte("\n"))
}

func normalizeLineEndingsString(content string) string {
	return string(normalizeLineEndings([]byte(content)))
}

// changeKind 根据目标文件是否存在返回 add 或 update。
func changeKind(targetPath string) string {
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return "add"
	}
	return "update"
}

// deleteStaleTemplates 删除目标目录中源文件已不存在的模板文件。
func deleteStaleTemplates(opts syncOptions, wanted map[string]struct{}) ([]action, error) {
	preserved := map[string]struct{}{
		"README.md.tmpl": {},
	}
	var actions []action
	if _, err := os.Stat(opts.target); errors.Is(err, os.ErrNotExist) {
		return actions, nil
	} else if err != nil {
		return nil, err
	}

	err := filepath.WalkDir(opts.target, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(opts.target, filePath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !strings.HasSuffix(rel, ".tmpl") {
			return nil
		}
		if _, ok := wanted[rel]; ok {
			return nil
		}
		if _, ok := preserved[rel]; ok {
			return nil
		}
		if opts.dryRun || opts.check {
			actions = append(actions, action{kind: "delete", path: filepath.ToSlash(filepath.Join(defaultTargetDir, rel))})
			return nil
		}
		if err := os.Remove(filePath); err != nil {
			return err
		}
		actions = append(actions, action{kind: "delete", path: filepath.ToSlash(filepath.Join(defaultTargetDir, rel))})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return actions, nil
}

func reportSyncActions(stdout io.Writer, opts syncOptions, actions []action) error {
	sort.Slice(actions, func(i, j int) bool {
		if actions[i].kind == actions[j].kind {
			return actions[i].path < actions[j].path
		}
		return actions[i].kind < actions[j].kind
	})
	if len(actions) == 0 {
		fmt.Fprintln(stdout, "api templates already in sync")
		return nil
	}
	for _, item := range actions {
		fmt.Fprintf(stdout, "%s %s\n", item.kind, item.path)
	}
	if opts.check {
		fmt.Fprintf(stdout, "check: %d pending changes\n", len(actions))
		return fmt.Errorf("api templates are out of sync")
	}
	if opts.dryRun {
		fmt.Fprintf(stdout, "dry-run: %d pending changes\n", len(actions))
		return nil
	}
	fmt.Fprintf(stdout, "synced %d template changes\n", len(actions))
	return nil
}

// fatal 输出错误并以非零状态退出。
func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
