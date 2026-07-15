package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/teamsillybees/initra/templates"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/mod/modfile"
	xmodule "golang.org/x/mod/module"
)

const frameworkModule = "github.com/teamsillybees/initra"

const (
	generatedSecretBytes        = 32
	generatedAdminPasswordBytes = 20
	maxAppSlugLength            = 48
)

type templateSecrets struct {
	local             string
	dev               string
	test              string
	adminPassword     string
	adminPasswordHash string
}

type newOptions struct {
	modulePath       string
	appName          string
	projectType      string
	templateName     string
	frameworkVersion string
	localReplacePath string
}

type templateData struct {
	ModulePath        string
	AppName           string
	AppSlug           string
	FrameworkModule   string
	FrameworkVersion  string
	LocalJWTSecret    string
	DevJWTSecret      string
	TestJWTSecret     string
	AdminPasswordHash string
}

// projectCommandRunner 在指定目录执行项目生成阶段的外部命令。
type projectCommandRunner func(dir string, name string, args ...string) ([]byte, error)

// projectTarget 记录最终目标目录、生成前状态和需要保留的目录权限。
type projectTarget struct {
	path    string
	existed bool
	mode    fs.FileMode
}

func newNewCommand(stdout io.Writer, cliVersion string) *cobra.Command {
	opts := newOptions{projectType: "api"}
	cmd := &cobra.Command{
		Use:           "new <dir>",
		Short:         "生成 API 项目",
		Long:          "根据标准 API 模板生成独立 Go module，自动执行 Ent 代码生成并初始化 Git 仓库。",
		Example:       "  initra new ./demo-api --type api --module example.com/demo-api\n  initra new $env:TEMP\\demo-api --type api --module example.com/demo-api --replace C:\\Project\\teamsillybees\\initra",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireExactArgs(1, "目标目录"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return createProject(args[0], cmd.OutOrStdout(), cliVersion, opts)
		},
	}
	configureCommand(cmd, stdout)
	flags := cmd.Flags()
	flags.StringVar(&opts.modulePath, "module", "", "生成项目的 Go module path，默认使用目录名")
	flags.StringVar(&opts.appName, "app-name", "", "应用名称，默认使用目录名")
	flags.StringVar(&opts.projectType, "type", "api", "项目类型，仅支持 api")
	flags.StringVar(&opts.templateName, "template", "", "旧版兼容参数，等同于 --type；basic 会映射为 api")
	flags.StringVar(&opts.frameworkVersion, "framework-version", "", "写入 go.mod 的 initra 框架版本")
	flags.StringVar(&opts.localReplacePath, "replace", "", "本地 initra 仓库路径，用于 go.mod replace")
	_ = cmd.RegisterFlagCompletionFunc("type", completeValues("api"))
	_ = cmd.RegisterFlagCompletionFunc("template", completeValues("api", "basic"))
	return cmd
}

func normalizeProjectModulePath(value string, appSlug string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		// 保留原有的目录名默认值；它是合法 import path，但用户显式指定时要求完整 module path。
		if err := xmodule.CheckImportPath(appSlug); err != nil {
			return "", fmt.Errorf("默认 Go module path %q 无效: %w", appSlug, err)
		}
		return appSlug, nil
	}
	if err := xmodule.CheckPath(value); err != nil {
		return "", fmt.Errorf("Go module path %q 无效: %w", value, err)
	}
	return value, nil
}

func validateFrameworkVersion(version string) (string, error) {
	if err := xmodule.Check(frameworkModule, version); err != nil {
		return "", fmt.Errorf("initra 框架版本 %q 无效: %w", version, err)
	}
	return version, nil
}

// normalizeAppSlug 将展示名称转换为可用于容器、数据库和缓存命名空间的稳定 slug。
func normalizeAppSlug(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	separatorPending := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			if separatorPending && builder.Len() > 0 {
				builder.WriteByte('-')
			}
			builder.WriteRune(r)
			separatorPending = false
		case unicode.IsSpace(r), r == '-', r == '_', r == '.', r == '/', r == '\\':
			separatorPending = builder.Len() > 0
		default:
			separatorPending = builder.Len() > 0
		}
	}
	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		return "", fmt.Errorf("应用名称 %q 无法转换为非空 AppSlug，请使用至少一个 ASCII 字母或数字", value)
	}
	if slug[0] >= '0' && slug[0] <= '9' {
		slug = "app-" + slug
	}
	if len(slug) > maxAppSlugLength {
		slug = strings.TrimRight(slug[:maxAppSlugLength], "-")
	}
	if slug == "" {
		return "", fmt.Errorf("应用名称 %q 无法转换为非空 AppSlug", value)
	}
	return slug, nil
}

func generateTemplateSecrets() (templateSecrets, error) {
	return generateTemplateSecretsFrom(rand.Reader)
}

func generateTemplateSecretsFrom(reader io.Reader) (templateSecrets, error) {
	local, err := generateURLSafeSecret(reader)
	if err != nil {
		return templateSecrets{}, fmt.Errorf("生成 local JWT secret 失败: %w", err)
	}
	dev, err := generateURLSafeSecret(reader)
	if err != nil {
		return templateSecrets{}, fmt.Errorf("生成 dev JWT secret 失败: %w", err)
	}
	test, err := generateURLSafeSecret(reader)
	if err != nil {
		return templateSecrets{}, fmt.Errorf("生成 test JWT secret 失败: %w", err)
	}
	adminPassword, err := generateURLSafeValue(reader, generatedAdminPasswordBytes)
	if err != nil {
		return templateSecrets{}, fmt.Errorf("生成管理员初始密码失败: %w", err)
	}
	adminPasswordHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return templateSecrets{}, fmt.Errorf("生成管理员初始密码哈希失败: %w", err)
	}
	return templateSecrets{
		local:             local,
		dev:               dev,
		test:              test,
		adminPassword:     adminPassword,
		adminPasswordHash: string(adminPasswordHash),
	}, nil
}

func generateURLSafeSecret(reader io.Reader) (string, error) {
	return generateURLSafeValue(reader, generatedSecretBytes)
}

func generateURLSafeValue(reader io.Reader, size int) (string, error) {
	value := make([]byte, size)
	if _, err := io.ReadFull(reader, value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func normalizeReplacePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("解析 replace 路径失败: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return "", fmt.Errorf("检查 replace 路径失败: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("replace 路径 %s 不是目录", absolute)
	}
	goModPath := filepath.Join(absolute, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("读取 replace 目标 go.mod 失败: %w", err)
	}
	parsed, err := modfile.Parse(goModPath, content, nil)
	if err != nil {
		return "", fmt.Errorf("解析 replace 目标 go.mod 失败: %w", err)
	}
	if parsed.Module == nil || parsed.Module.Mod.Path != frameworkModule {
		actual := ""
		if parsed.Module != nil {
			actual = parsed.Module.Mod.Path
		}
		return "", fmt.Errorf("replace 目标模块必须是 %s，实际为 %q", frameworkModule, actual)
	}
	return filepath.ToSlash(absolute), nil
}

// applyFrameworkReplace 使用 modfile API 写入 replace，确保含空格路径得到正确引用。
func applyFrameworkReplace(goModPath string, replacePath string) error {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("读取生成项目 go.mod 失败: %w", err)
	}
	parsed, err := modfile.Parse(goModPath, content, nil)
	if err != nil {
		return fmt.Errorf("解析生成项目 go.mod 失败: %w", err)
	}
	if err := parsed.AddReplace(frameworkModule, "", replacePath, ""); err != nil {
		return fmt.Errorf("写入本地 replace 失败: %w", err)
	}
	formatted, err := parsed.Format()
	if err != nil {
		return fmt.Errorf("格式化生成项目 go.mod 失败: %w", err)
	}
	if err := os.WriteFile(goModPath, formatted, 0o644); err != nil {
		return fmt.Errorf("保存生成项目 go.mod 失败: %w", err)
	}
	return nil
}

func createProject(targetDir string, stdout io.Writer, cliVersion string, opts newOptions) error {
	return createProjectWithRunner(targetDir, stdout, cliVersion, opts, executeProjectCommand)
}

// createProjectWithRunner 在临时目录完成项目生成，成功后再把完整项目移动到目标目录。
func createProjectWithRunner(targetDir string, stdout io.Writer, cliVersion string, opts newOptions, runner projectCommandRunner) error {
	resolvedType, err := normalizeProjectType(opts.projectType)
	if err != nil {
		return err
	}
	if t := strings.TrimSpace(opts.templateName); t != "" {
		templateType, err := normalizeProjectType(t)
		if err != nil {
			return err
		}
		if resolvedType != templateType {
			return fmt.Errorf("--type 与 --template 不能指定不同项目类型")
		}
		resolvedType = templateType
	}
	target, err := inspectProjectTarget(targetDir)
	if err != nil {
		return err
	}

	normalizedAppName := strings.TrimSpace(opts.appName)
	if normalizedAppName == "" {
		normalizedAppName = filepath.Base(target.path)
	}
	appSlug, err := normalizeAppSlug(normalizedAppName)
	if err != nil {
		return err
	}
	normalizedModulePath, err := normalizeProjectModulePath(opts.modulePath, appSlug)
	if err != nil {
		return err
	}

	resolvedReplace, err := normalizeReplacePath(opts.localReplacePath)
	if err != nil {
		return err
	}
	resolvedVersion, err := resolveFrameworkVersion(opts.frameworkVersion, cliVersion, resolvedReplace)
	if err != nil {
		return err
	}
	secrets, err := generateTemplateSecrets()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(target.path), 0o755); err != nil {
		return fmt.Errorf("创建目标父目录失败: %w", err)
	}
	workDir, err := os.MkdirTemp(filepath.Dir(target.path), ".initra-new-*")
	if err != nil {
		return fmt.Errorf("创建项目临时目录失败: %w", err)
	}
	defer os.RemoveAll(workDir)

	data := templateData{
		ModulePath:        normalizedModulePath,
		AppName:           normalizedAppName,
		AppSlug:           appSlug,
		FrameworkModule:   frameworkModule,
		FrameworkVersion:  resolvedVersion,
		LocalJWTSecret:    secrets.local,
		DevJWTSecret:      secrets.dev,
		TestJWTSecret:     secrets.test,
		AdminPasswordHash: secrets.adminPasswordHash,
	}
	if err := renderTemplate(resolvedType, workDir, data); err != nil {
		return err
	}
	if resolvedReplace != "" {
		if err := applyFrameworkReplace(filepath.Join(workDir, "go.mod"), resolvedReplace); err != nil {
			return err
		}
	}
	if resolvedType == "api" {
		if err := prepareProjectDependencies(workDir, runner); err != nil {
			return err
		}
		if err := generateEntCode(workDir, runner); err != nil {
			return err
		}
		if err := validateGeneratedProject(workDir, runner); err != nil {
			return err
		}
	}
	if err := initGitRepository(workDir, runner); err != nil {
		return err
	}
	if err := commitProject(workDir, target); err != nil {
		return err
	}

	if stdout != nil {
		if _, err := fmt.Fprintf(
			stdout,
			"created %s\ninitial admin password (shown once): %s\n",
			targetDir,
			secrets.adminPassword,
		); err != nil {
			return fmt.Errorf("项目已创建，但输出一次性管理员密码失败；请删除项目后重新生成，或重置 seed 中的密码哈希: %w", err)
		}
	}
	return nil
}

// normalizeProjectType 将历史项目类型归一到当前唯一的 API 模板。
func normalizeProjectType(projectType string) (string, error) {
	normalized := strings.TrimSpace(projectType)
	switch normalized {
	case "", "api", "basic":
		return "api", nil
	default:
		return "", fmt.Errorf("暂不支持项目类型 %q，仅支持 api", normalized)
	}
}

func prepareProjectDependencies(targetDir string, runner projectCommandRunner) error {
	output, err := runner(targetDir, "go", "mod", "download", "all")
	if err != nil {
		return projectCommandError("下载项目依赖失败", err, output)
	}
	return nil
}

func generateEntCode(targetDir string, runner projectCommandRunner) error {
	output, err := runner(targetDir, "go", "run", "./internal/data/entgenerate")
	if err != nil {
		return projectCommandError("生成 Ent 代码失败", err, output)
	}
	return nil
}

func validateGeneratedProject(targetDir string, runner projectCommandRunner) error {
	output, err := runner(targetDir, "go", "test", "./...", "-count=1")
	if err != nil {
		return projectCommandError("验证生成项目失败", err, output)
	}
	return nil
}

func initGitRepository(targetDir string, runner projectCommandRunner) error {
	output, err := runner(targetDir, "git", "init")
	if err != nil {
		return projectCommandError("初始化 git 仓库失败", err, output)
	}
	return nil
}

func executeProjectCommand(dir string, name string, args ...string) ([]byte, error) {
	command := exec.Command(name, args...)
	command.Dir = dir
	command.Env = append(os.Environ(), "GOWORK=off")
	return command.CombinedOutput()
}

func projectCommandError(action string, err error, output []byte) error {
	message := strings.TrimSpace(string(output))
	if message == "" {
		return fmt.Errorf("%s: %w", action, err)
	}
	return fmt.Errorf("%s: %w: %s", action, err, message)
}

func resolveFrameworkVersion(flagVersion string, cliVersion string, replacePath string) (string, error) {
	if version := strings.TrimSpace(flagVersion); version != "" {
		return validateFrameworkVersion(version)
	}
	if replacePath != "" {
		return validateFrameworkVersion("v0.0.0")
	}
	if version := strings.TrimSpace(cliVersion); version != "" && version != "dev" {
		return validateFrameworkVersion(version)
	}
	return "", fmt.Errorf("开发版 CLI 必须提供 --framework-version 或 --replace")
}

func inspectProjectTarget(targetDir string) (projectTarget, error) {
	if strings.TrimSpace(targetDir) == "" {
		return projectTarget{}, fmt.Errorf("目标目录不能为空")
	}
	absolute, err := filepath.Abs(targetDir)
	if err != nil {
		return projectTarget{}, fmt.Errorf("解析目标目录失败: %w", err)
	}
	target := projectTarget{path: filepath.Clean(absolute), mode: 0o755}
	entries, err := os.ReadDir(target.path)
	switch {
	case err == nil && len(entries) > 0:
		return projectTarget{}, fmt.Errorf("目标目录 %s 已存在且非空", targetDir)
	case err == nil:
		info, statErr := os.Stat(target.path)
		if statErr != nil {
			return projectTarget{}, statErr
		}
		target.existed = true
		target.mode = info.Mode().Perm()
		return target, nil
	case errors.Is(err, os.ErrNotExist):
		return target, nil
	default:
		return projectTarget{}, err
	}
}

func commitProject(workDir string, target projectTarget) (returnErr error) {
	if err := os.Chmod(workDir, target.mode); err != nil {
		return fmt.Errorf("设置项目目录权限失败: %w", err)
	}
	restoreWorkingDirectory, err := leaveTargetWorkingDirectory(target)
	if err != nil {
		return err
	}
	if restoreWorkingDirectory != nil {
		defer func() {
			if restoreErr := restoreWorkingDirectory(); restoreErr != nil {
				returnErr = errors.Join(returnErr, fmt.Errorf("恢复项目工作目录失败: %w", restoreErr))
			}
		}()
	}

	var backupRoot string
	var backupPath string
	if target.existed {
		entries, err := os.ReadDir(target.path)
		if err != nil {
			return fmt.Errorf("重新检查目标目录失败: %w", err)
		}
		if len(entries) > 0 {
			return fmt.Errorf("目标目录 %s 在生成期间变为非空", target.path)
		}
		backupRoot, err = os.MkdirTemp(filepath.Dir(target.path), ".initra-target-backup-*")
		if err != nil {
			return fmt.Errorf("创建目标目录备份位置失败: %w", err)
		}
		backupPath = filepath.Join(backupRoot, "target")
		if err := os.Rename(target.path, backupPath); err != nil {
			_ = os.Remove(backupRoot)
			return fmt.Errorf("备份现有空目标目录失败: %w", err)
		}
	} else if _, err := os.Stat(target.path); err == nil {
		return fmt.Errorf("目标目录 %s 在生成期间已被创建", target.path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("重新检查目标目录失败: %w", err)
	}

	if err := os.Rename(workDir, target.path); err != nil {
		commitErr := fmt.Errorf("提交生成项目失败: %w", err)
		if backupPath == "" {
			return commitErr
		}
		if restoreErr := os.Rename(backupPath, target.path); restoreErr != nil {
			return errors.Join(
				commitErr,
				fmt.Errorf("恢复原目标目录失败，原目录保留在 %s: %w", backupPath, restoreErr),
			)
		}
		if cleanupErr := os.Remove(backupRoot); cleanupErr != nil {
			return errors.Join(commitErr, fmt.Errorf("清理目标目录备份位置失败: %w", cleanupErr))
		}
		return commitErr
	}
	if backupRoot != "" {
		// 备份中只有原来的空目录；新项目已原子提交，清理失败不应让调用方丢失一次性凭据。
		_ = os.RemoveAll(backupRoot)
	}
	return nil
}

// leaveTargetWorkingDirectory 在 Windows 提交当前工作目录时临时切换到其父目录。
func leaveTargetWorkingDirectory(target projectTarget) (func() error, error) {
	if !target.existed {
		return nil, nil
	}
	currentInfo, err := os.Stat(".")
	if err != nil {
		return nil, fmt.Errorf("检查当前工作目录失败: %w", err)
	}
	targetInfo, err := os.Stat(target.path)
	if err != nil {
		return nil, fmt.Errorf("检查目标工作目录失败: %w", err)
	}
	if !os.SameFile(currentInfo, targetInfo) {
		return nil, nil
	}
	if err := os.Chdir(filepath.Dir(target.path)); err != nil {
		return nil, fmt.Errorf("临时切换到目标父目录失败: %w", err)
	}
	return func() error {
		return os.Chdir(target.path)
	}, nil
}

func renderTemplate(templateName string, targetDir string, data templateData) error {
	root := templateName
	return fs.WalkDir(templates.FS, root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}

		relativePath, err := filepath.Rel(root, filepath.ToSlash(path))
		if err != nil {
			return err
		}
		outputPath := filepath.Join(targetDir, filepath.FromSlash(strings.TrimSuffix(relativePath, ".tmpl")))

		if entry.IsDir() {
			return os.MkdirAll(outputPath, 0o755)
		}
		if !strings.HasSuffix(relativePath, ".tmpl") {
			return nil
		}

		content, err := templates.FS.ReadFile(path)
		if err != nil {
			return err
		}
		parsed, err := template.New(relativePath).Option("missingkey=error").Parse(string(content))
		if err != nil {
			return fmt.Errorf("解析模板 %s 失败: %w", relativePath, err)
		}

		var rendered bytes.Buffer
		if err := parsed.Execute(&rendered, data); err != nil {
			return fmt.Errorf("渲染模板 %s 失败: %w", relativePath, err)
		}

		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(outputPath, rendered.Bytes(), 0o644)
	})
}
