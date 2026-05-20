//go:build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

const (
	// defaultSourceDir 是 API 示例项目的默认来源目录。
	defaultSourceDir = "examples/api"
	// defaultTargetDir 是 API 模板文件的默认目标目录。
	defaultTargetDir = "templates/api"
	// exampleModule 是 examples/api 中需要替换的固定模块路径。
	exampleModule = "github.com/teamsillybees/initra/examples/api"
	// frameworkModule 是生成项目依赖的 initra 根模块路径。
	frameworkModule = "github.com/teamsillybees/initra"
)

// frameworkRequirePattern 匹配 examples/api go.mod 中的 initra 依赖版本行。
var frameworkRequirePattern = regexp.MustCompile(`(?m)^\tgithub\.com/teamsillybees/initra v[^\r\n]+$`)

// syncOptions 描述本次 API 模板同步的输入和行为开关。
type syncOptions struct {
	repoRoot string
	source   string
	target   string
	dryRun   bool
	delete   bool
}

// action 记录一次同步动作，用于 dry-run 和执行结果输出。
type action struct {
	kind string
	path string
}

// main 执行 examples/api 到 templates/api 的同步。
func main() {
	opts, err := parseOptions()
	if err != nil {
		fatal(err)
	}

	actions, err := syncTemplates(opts)
	if err != nil {
		fatal(err)
	}

	sort.Slice(actions, func(i, j int) bool {
		if actions[i].kind == actions[j].kind {
			return actions[i].path < actions[j].path
		}
		return actions[i].kind < actions[j].kind
	})

	if len(actions) == 0 {
		fmt.Println("api templates already in sync")
		return
	}

	for _, item := range actions {
		fmt.Printf("%s %s\n", item.kind, item.path)
	}
	if opts.dryRun {
		fmt.Printf("dry-run: %d pending changes\n", len(actions))
		return
	}
	fmt.Printf("synced %d template changes\n", len(actions))
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
	if !opts.dryRun {
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
		rendered := transformTemplateContent(rel, string(content))
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
		if opts.dryRun {
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
	if !strings.HasPrefix(rel, "internal/ent/") {
		return false
	}
	return rel != "internal/ent/generate.go" &&
		!strings.HasPrefix(rel, "internal/ent/schema/") &&
		!strings.HasPrefix(rel, "internal/ent/migratediff/")
}

// shouldSkipDir 判断目录是否属于需要跳过的 Ent 生成代码目录。
func shouldSkipDir(rel string) bool {
	if !strings.HasPrefix(rel, "internal/ent") {
		return false
	}
	if rel == "internal/ent" || rel == "internal/ent/schema" || rel == "internal/ent/migratediff" {
		return false
	}
	return !strings.HasPrefix(rel, "internal/ent/schema/")
}

// transformTemplateContent 将示例项目中的固定值替换为模板占位符。
func transformTemplateContent(rel string, content string) string {
	content = strings.ReplaceAll(content, exampleModule, "{{ .ModulePath }}")

	if rel == "go.mod" {
		content = frameworkRequirePattern.ReplaceAllString(content, "\t{{ .FrameworkModule }} {{ .FrameworkVersion }}")
		content = strings.Replace(
			content,
			"replace github.com/teamsillybees/initra => ../..",
			"{{- if .LocalReplacePath }}\nreplace {{ .FrameworkModule }} => {{ .LocalReplacePath }}\n{{- end }}",
			1,
		)
	}

	return content
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
		return !bytes.Equal(current, content), nil
	}
	if os.IsNotExist(err) {
		return true, nil
	}
	return false, err
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
		if opts.dryRun {
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

// fatal 输出错误并以非零状态退出。
func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
