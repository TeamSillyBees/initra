package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/teamsillybees/initra/templates"
)

const frameworkModule = "github.com/teamsillybees/initra"

var (
	goPackageNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	safeNamePattern      = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

var version = "dev"

type newOptions struct {
	modulePath       string
	appName          string
	projectType      string
	templateName     string
	frameworkVersion string
	localReplacePath string
}

type crudAddOptions struct {
	tableName string
}

type templateData struct {
	ModulePath       string
	AppName          string
	FrameworkModule  string
	FrameworkVersion string
	LocalReplacePath string
}

func main() {
	if err := run(os.Args[1:], os.Stdout, currentCLIVersion()); err != nil {
		log.Fatalf("initra: %v", err)
	}
}

func currentCLIVersion() string {
	return resolveCLIVersion(version, buildInfoVersion())
}

func resolveCLIVersion(injectedVersion string, buildVersion string) string {
	if version := strings.TrimSpace(injectedVersion); version != "" && version != "dev" {
		return version
	}
	if version := strings.TrimSpace(buildVersion); version != "" && version != "(devel)" {
		return version
	}
	if version := strings.TrimSpace(injectedVersion); version != "" {
		return version
	}
	return "dev"
}

func buildInfoVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	return info.Main.Version
}

func run(args []string, stdout io.Writer, cliVersion string) error {
	cmd := newRootCommand(stdout, cliVersion)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func newRootCommand(stdout io.Writer, cliVersion string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "initra",
		Short:         "企业内部 Go 服务快速开发脚手架",
		Version:       cliVersion,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("用法: initra <command>")
		},
	}
	configureCommand(cmd, stdout)
	cmd.AddCommand(
		newNewCommand(stdout, cliVersion),
		newModuleCommand(stdout),
		newCRUDCommand(stdout),
		newConfigCommand(stdout),
		newMigrateCommand(stdout),
		newDoctorCommand(stdout),
	)
	return cmd
}

func configureCommand(cmd *cobra.Command, stdout io.Writer) {
	if stdout == nil {
		stdout = io.Discard
	}
	cmd.SetOut(stdout)
	cmd.SetErr(io.Discard)
	cmd.DisableSuggestions = true
}

func newNewCommand(stdout io.Writer, cliVersion string) *cobra.Command {
	opts := newOptions{projectType: "api"}
	cmd := &cobra.Command{
		Use:           "new <dir>",
		Short:         "生成 API 或 worker 项目",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("用法: initra new <dir>")
			}
			if len(args) > 1 {
				return fmt.Errorf("new 命令只接受一个目标目录")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return createProject(args[0], cmd.OutOrStdout(), cliVersion, opts)
		},
	}
	configureCommand(cmd, stdout)
	flags := cmd.Flags()
	flags.StringVar(&opts.modulePath, "module", "", "生成项目的 Go module path")
	flags.StringVar(&opts.appName, "app-name", "", "应用名称")
	flags.StringVar(&opts.projectType, "type", "api", "项目类型：api 或 worker")
	flags.StringVar(&opts.templateName, "template", "", "兼容旧参数，等同于 --type")
	flags.StringVar(&opts.frameworkVersion, "framework-version", "", "initra 框架版本")
	flags.StringVar(&opts.localReplacePath, "replace", "", "本地 initra 仓库路径，用于 go.mod replace")
	return cmd
}

func newModuleCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "module",
		Short:         "管理业务模块骨架",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("用法: initra module add <name>")
		},
	}
	configureCommand(cmd, stdout)
	cmd.AddCommand(newModuleAddCommand(stdout))
	return cmd
}

func newModuleAddCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "add <name>",
		Short:         "生成 flat package 业务模块骨架",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("用法: initra module add <name>")
			}
			if len(args) > 1 {
				return fmt.Errorf("module add 只接受一个模块名")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return addModule(args[0], cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	return cmd
}

func newCRUDCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "crud",
		Short:         "管理 CRUD 示例代码",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("用法: initra crud add <module> --table <table>")
		},
	}
	configureCommand(cmd, stdout)
	cmd.AddCommand(newCRUDAddCommand(stdout))
	return cmd
}

func newCRUDAddCommand(stdout io.Writer) *cobra.Command {
	opts := crudAddOptions{}
	cmd := &cobra.Command{
		Use:           "add <module>",
		Short:         "为现有模块生成 CRUD 示例",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("crud add 缺少模块名")
			}
			if len(args) > 1 {
				return fmt.Errorf("crud add 只接受一个模块名")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return addCRUDSample(args[0], opts, cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	cmd.Flags().StringVar(&opts.tableName, "table", "", "数据表名")
	return cmd
}

func newConfigCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "config",
		Short:         "管理配置片段",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("用法: initra config add <capability>")
		},
	}
	configureCommand(cmd, stdout)
	cmd.AddCommand(newConfigAddCommand(stdout))
	return cmd
}

func newConfigAddCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "add <capability>",
		Short:         "生成配置结构和 YAML 示例",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("用法: initra config add <capability>")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return addConfigSnippet(args[0], cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	return cmd
}

func newMigrateCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "migrate",
		Short:         "管理迁移辅助文件",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("用法: initra migrate <new|diff> <name>")
		},
	}
	configureCommand(cmd, stdout)
	cmd.AddCommand(newMigrateNewCommand(stdout), newMigrateDiffCommand(stdout))
	return cmd
}

func newMigrateNewCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "new <name>",
		Short:         "创建空迁移文件",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("用法: initra migrate new <name>")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return createMigrationArtifact("new", args[0], cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	return cmd
}

func newMigrateDiffCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "diff <name>",
		Short:         "创建 Atlas diff 脚本",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("用法: initra migrate diff <name>")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return createMigrationArtifact("diff", args[0], cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	return cmd
}

func newDoctorCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "doctor",
		Short:         "检查本地开发环境",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("用法: initra doctor")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctorChecks(cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	return cmd
}

func runNew(args []string, stdout io.Writer, cliVersion string) error {
	cmd := newNewCommand(stdout, cliVersion)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func createProject(targetDir string, stdout io.Writer, cliVersion string, opts newOptions) error {
	resolvedType := strings.TrimSpace(opts.projectType)
	if resolvedType == "" {
		resolvedType = "api"
	}
	if template := strings.TrimSpace(opts.templateName); template != "" {
		if resolvedType != "api" && resolvedType != template {
			return fmt.Errorf("--type 与 --template 不能指定不同项目类型")
		}
		resolvedType = template
	}
	if resolvedType == "basic" {
		resolvedType = "api"
	}
	if resolvedType != "api" && resolvedType != "worker" {
		return fmt.Errorf("暂不支持项目类型 %q", resolvedType)
	}

	normalizedAppName := strings.TrimSpace(opts.appName)
	if normalizedAppName == "" {
		normalizedAppName = filepath.Base(filepath.Clean(targetDir))
	}
	normalizedModulePath := strings.TrimSpace(opts.modulePath)
	if normalizedModulePath == "" {
		normalizedModulePath = normalizedAppName
	}

	resolvedReplace, err := normalizeReplacePath(opts.localReplacePath)
	if err != nil {
		return err
	}
	resolvedVersion, err := resolveFrameworkVersion(opts.frameworkVersion, cliVersion, resolvedReplace)
	if err != nil {
		return err
	}

	if err := ensureWritableTarget(targetDir); err != nil {
		return err
	}

	data := templateData{
		ModulePath:       normalizedModulePath,
		AppName:          normalizedAppName,
		FrameworkModule:  frameworkModule,
		FrameworkVersion: resolvedVersion,
		LocalReplacePath: resolvedReplace,
	}
	if err := renderTemplate(resolvedType, targetDir, data); err != nil {
		return err
	}
	if resolvedType == "api" {
		if err := generateEntCode(targetDir); err != nil {
			return err
		}
	}
	if err := initGitRepository(targetDir); err != nil {
		return err
	}

	if stdout != nil {
		_, _ = fmt.Fprintf(stdout, "created %s\n", targetDir)
	}
	return nil
}

func generateEntCode(targetDir string) error {
	command := exec.Command("go", "generate", "./internal/ent")
	command.Dir = targetDir
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return fmt.Errorf("生成 Ent 代码失败: %w", err)
		}
		return fmt.Errorf("生成 Ent 代码失败: %w: %s", err, message)
	}
	return nil
}

func initGitRepository(targetDir string) error {
	output, err := exec.Command("git", "-C", targetDir, "init").CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return fmt.Errorf("初始化 git 仓库失败: %w", err)
		}
		return fmt.Errorf("初始化 git 仓库失败: %w: %s", err, message)
	}
	return nil
}

func runModule(args []string, stdout io.Writer) error {
	cmd := newModuleCommand(stdout)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func addModule(name string, stdout io.Writer) error {
	name, err := normalizeGoPackageName(name)
	if err != nil {
		return err
	}

	moduleDir := filepath.Join("internal", "module", name)
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		return err
	}

	files := map[string]string{
		name + ".model.go":   moduleModelTemplate(name),
		name + ".service.go": moduleServiceTemplate(name),
		name + ".repo.go":    moduleRepoTemplate(name),
		name + ".dto.go":     moduleDTOTemplate(name),
		name + ".handler.go": moduleHandlerTemplate(name),
		name + ".routes.go":  moduleRoutesTemplate(name),
		"providers.go":       moduleProvidersTemplate(name),
		name + "_test.go":    moduleTestTemplate(name),
	}
	for filename, content := range files {
		if err := writeNewFile(filepath.Join(moduleDir, filename), content); err != nil {
			return err
		}
	}

	if stdout != nil {
		_, _ = fmt.Fprintf(stdout, "created module %s\n", name)
	}
	return nil
}

func runCRUD(args []string, stdout io.Writer) error {
	cmd := newCRUDCommand(stdout)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func addCRUDSample(moduleName string, opts crudAddOptions, stdout io.Writer) error {
	moduleName, err := normalizeGoPackageName(moduleName)
	if err != nil {
		return err
	}

	table := strings.TrimSpace(opts.tableName)
	if table == "" {
		return fmt.Errorf("crud add 必须提供 --table")
	}

	moduleDir := filepath.Join("internal", "module", moduleName)
	if _, err := os.Stat(moduleDir); err != nil {
		return fmt.Errorf("模块 %s 不存在，请先执行 initra module add %s", moduleName, moduleName)
	}
	if err := writeNewFile(filepath.Join(moduleDir, moduleName+".crud.go"), crudTemplate(moduleName, table)); err != nil {
		return err
	}
	if stdout != nil {
		_, _ = fmt.Fprintf(stdout, "created crud sample for %s\n", moduleName)
	}
	return nil
}

func runConfig(args []string, stdout io.Writer) error {
	cmd := newConfigCommand(stdout)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func addConfigSnippet(capability string, stdout io.Writer) error {
	capability, err := normalizeGoPackageName(capability)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join("internal", "boot"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll("configs", 0o755); err != nil {
		return err
	}
	if err := writeNewFile(filepath.Join("internal", "boot", capability+".config.go"), configGoTemplate(capability)); err != nil {
		return err
	}
	if err := writeNewFile(filepath.Join("configs", capability+".yaml"), configYAMLTemplate(capability)); err != nil {
		return err
	}
	if stdout != nil {
		_, _ = fmt.Fprintf(stdout, "created config %s\n", capability)
	}
	return nil
}

func runMigrate(args []string, stdout io.Writer) error {
	cmd := newMigrateCommand(stdout)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func createMigrationArtifact(kind string, name string, stdout io.Writer) error {
	name, err := normalizeSafeName(name)
	if err != nil {
		return err
	}

	switch kind {
	case "new":
		if err := os.MkdirAll(filepath.Join("db", "migrations"), 0o755); err != nil {
			return err
		}
		filename := time.Now().UTC().Format("20060102150405") + "_" + name + ".sql"
		path := filepath.Join("db", "migrations", filename)
		if err := writeNewFile(path, "-- "+name+"\n\n"); err != nil {
			return err
		}
		if stdout != nil {
			_, _ = fmt.Fprintf(stdout, "created migration %s\n", path)
		}
		return nil
	case "diff":
		if err := os.MkdirAll("scripts", 0o755); err != nil {
			return err
		}
		path := filepath.Join("scripts", "migrate-diff-"+name+".ps1")
		content := migrateDiffScript(name)
		if err := writeNewFile(path, content); err != nil {
			return err
		}
		if stdout != nil {
			_, _ = fmt.Fprintf(stdout, "created migrate diff script %s\n", path)
		}
		return nil
	default:
		return fmt.Errorf("未知 migrate 子命令 %q", kind)
	}
}

func migrateDiffScript(name string) string {
	return fmt.Sprintf(`param(
    [string]$Env = "",
    [string]$ConfigDir = "configs",
    [string]$DevURL = ""
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
try {
    $optionalArgs = @()
    if (![string]::IsNullOrWhiteSpace($Env)) {
        $optionalArgs += @("-env", $Env)
    }
    if (![string]::IsNullOrWhiteSpace($DevURL)) {
        $optionalArgs += @("-dev-url", $DevURL)
    }
    go run ./internal/ent/migratediff/main.go %s -config-dir $ConfigDir @optionalArgs
}
finally {
    Pop-Location
}
`, name)
}

func runDoctor(args []string, stdout io.Writer) error {
	cmd := newDoctorCommand(stdout)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func runDoctorChecks(stdout io.Writer) error {
	if stdout == nil {
		stdout = io.Discard
	}

	reportTool(stdout, "Go", "go", "version")
	reportTool(stdout, "Atlas", "atlas", "version")
	reportTool(stdout, "Ent", "go", "run", "entgo.io/ent/cmd/ent", "--help")
	reportTool(stdout, "golangci-lint", "golangci-lint", "version")
	reportFile(stdout, "config.yaml", filepath.Join("configs", "config.yaml"))
	reportFile(stdout, "config.dev.yaml", filepath.Join("configs", "config.dev.yaml"))
	reportFile(stdout, "Atlas config", filepath.Join("db", "atlas.hcl"))
	return nil
}

func resolveFrameworkVersion(flagVersion string, cliVersion string, replacePath string) (string, error) {
	if version := strings.TrimSpace(flagVersion); version != "" {
		return version, nil
	}
	if replacePath != "" {
		return "v0.0.0", nil
	}
	if version := strings.TrimSpace(cliVersion); version != "" && version != "dev" {
		return version, nil
	}
	return "", fmt.Errorf("开发版 CLI 必须提供 --framework-version 或 --replace")
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
	return filepath.ToSlash(absolute), nil
}

func ensureWritableTarget(targetDir string) error {
	entries, err := os.ReadDir(targetDir)
	switch {
	case err == nil && len(entries) > 0:
		return fmt.Errorf("目标目录 %s 已存在且非空", targetDir)
	case err == nil:
		return nil
	case errors.Is(err, os.ErrNotExist):
		return os.MkdirAll(targetDir, 0o755)
	default:
		return err
	}
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

func normalizeGoPackageName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if !goPackageNamePattern.MatchString(name) {
		return "", fmt.Errorf("名称 %q 必须匹配 %s", name, goPackageNamePattern.String())
	}
	return name, nil
}

func normalizeSafeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if !safeNamePattern.MatchString(name) {
		return "", fmt.Errorf("名称 %q 只能包含字母、数字、下划线和中划线", name)
	}
	return name, nil
}

func writeNewFile(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("文件 %s 已存在", path)
		}
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

func reportTool(stdout io.Writer, label string, command string, args ...string) {
	path, err := exec.LookPath(command)
	if err != nil {
		_, _ = fmt.Fprintf(stdout, "%s: MISSING (%s)\n", label, command)
		return
	}
	output, err := exec.Command(command, args...).CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintf(stdout, "%s: FOUND %s\n", label, path)
		return
	}
	_, _ = fmt.Fprintf(stdout, "%s: OK %s\n", label, firstLine(string(output)))
}

func reportFile(stdout io.Writer, label string, path string) {
	if _, err := os.Stat(path); err != nil {
		_, _ = fmt.Fprintf(stdout, "%s: MISSING %s\n", label, path)
		return
	}
	_, _ = fmt.Fprintf(stdout, "%s: OK %s\n", label, path)
}

func firstLine(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if index := strings.IndexAny(value, "\r\n"); index >= 0 {
		return value[:index]
	}
	return value
}

func exportedName(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || r == '-'
	})
	var builder strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			builder.WriteString(part[1:])
		}
	}
	return builder.String()
}

func pluralName(name string) string {
	if strings.HasSuffix(name, "s") {
		return name
	}
	return name + "s"
}

func moduleModelTemplate(name string) string {
	typeName := exportedName(name)
	return fmt.Sprintf(`package %s

// %s 是 %s 模块的领域占位模型。
type %s struct {
	ID int64
}
`, name, typeName, name, typeName)
}

func moduleServiceTemplate(name string) string {
	typeName := exportedName(name)
	return fmt.Sprintf(`package %s

import "context"

// Service 是 %s 模块的应用服务。
type Service struct{}

// NewService 创建 %s 模块应用服务。
func NewService() *Service {
	return &Service{}
}

// Get 返回 %s 详情占位数据。
func (s *Service) Get(ctx context.Context, id int64) (*%s, error) {
	_ = s
	_ = ctx
	return &%s{ID: id}, nil
}
`, name, name, name, name, typeName, typeName)
}

func moduleRepoTemplate(name string) string {
	return fmt.Sprintf(`package %s

// Repository 是 %s 模块的数据访问占位实现。
type Repository struct{}

// NewRepository 创建 %s 模块仓储。
func NewRepository() *Repository {
	return &Repository{}
}
`, name, name, name)
}

func moduleHandlerTemplate(name string) string {
	typeName := exportedName(name)
	return fmt.Sprintf(`package %s

import (
	"context"

	"github.com/teamsillybees/initra/pkg/requestctx"
	"github.com/teamsillybees/initra/pkg/response"
)

// Handler 封装 %s 模块 HTTP 适配逻辑。
type Handler struct {
	service *Service
}

// NewHandler 创建 %s 模块 Handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) get(ctx context.Context, input *get%sRequest) (*get%sResponse, error) {
	item, err := h.service.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	return &get%sResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), %sVO{ID: item.ID}),
	}, nil
}
`, name, name, name, typeName, typeName, typeName, typeName)
}

func moduleDTOTemplate(name string) string {
	typeName := exportedName(name)
	return fmt.Sprintf(`package %s

import "github.com/teamsillybees/initra/pkg/response"

type get%sRequest struct {
	ID int64 `+"`path:\"id\" doc:\"ID\"`"+`
}

// %sVO 描述 %s 对外 JSON DTO。
type %sVO struct {
	ID int64 `+"`json:\"id\"`"+`
}

type get%sResponse struct {
	Body response.SuccessVO[%sVO]
}
`, name, typeName, typeName, name, typeName, typeName, typeName)
}

func moduleRoutesTemplate(name string) string {
	typeName := exportedName(name)
	path := "/api/v1/" + pluralName(name) + "/{id}"
	return fmt.Sprintf(`package %s

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/server"
)

// Module 负责 %s 模块路由注册。
type Module struct {
	handler *Handler
}

// NewModule 创建 %s 模块实例。
func NewModule(handler *Handler) *Module {
	return &Module{handler: handler}
}

// Register 将 %s 模块注册到应用。
func (m *Module) Register(api huma.API, registry *server.RouteRegistry) {
	huma.Register(api, huma.Operation{
		OperationID: "get-%s",
		Method:      http.MethodGet,
		Path:        "%s",
		Summary:     "查询%s详情",
		Tags:        []string{"%s"},
	}, m.handler.get)
	registry.Register(http.MethodGet, "%s", platformauth.RouteSecurity{Resource: "%s", Action: "read"})
}
`, name, name, name, name, name, path, typeName, typeName, path, name)
}

func moduleProvidersTemplate(name string) string {
	return fmt.Sprintf(`package %s

import "github.com/samber/do"

const (
	%sServiceName = "%s.service"
	%sHandlerName = "%s.handler"
)

// Provide 使用 do 注册 %s 模块依赖。
func Provide(injector *do.Injector) {
	do.ProvideNamed(injector, %sServiceName, func(i *do.Injector) (*Service, error) {
		return NewService(), nil
	})
	do.ProvideNamed(injector, %sHandlerName, func(i *do.Injector) (*Handler, error) {
		service := do.MustInvokeNamed[*Service](i, %sServiceName)
		return NewHandler(service), nil
	})
	do.Provide(injector, func(i *do.Injector) (*Module, error) {
		handler := do.MustInvokeNamed[*Handler](i, %sHandlerName)
		return NewModule(handler), nil
	})
}
`, name, name, name, name, name, name, name, name, name, name)
}

func moduleTestTemplate(name string) string {
	typeName := exportedName(name)
	return fmt.Sprintf(`package %s

import (
	"context"
	"testing"
)

// TestServiceGet 验证 %s 模块服务占位逻辑可调用。
func TestServiceGet(t *testing.T) {
	item, err := NewService().Get(context.Background(), 1001)
	if err != nil {
		t.Fatalf("Get() error = %%v", err)
	}
	if item.ID != 1001 {
		t.Fatalf("Get() ID = %%d, want 1001", item.ID)
	}
}

var _ = (*%s)(nil)
`, name, name, typeName)
}

func crudTemplate(moduleName string, tableName string) string {
	typeName := exportedName(moduleName)
	return fmt.Sprintf(`package %s

const %sCRUDTable = %q

// %sCRUD 是基于数据表 %s 的 CRUD 样例占位。
type %sCRUD struct{}

// New%sCRUD 创建 CRUD 样例。
func New%sCRUD() *%sCRUD {
	return &%sCRUD{}
}
`, moduleName, moduleName, tableName, typeName, tableName, typeName, typeName, typeName, typeName, typeName)
}

func configGoTemplate(capability string) string {
	typeName := exportedName(capability) + "Config"
	return fmt.Sprintf(`package boot

// %s 描述 %s 能力的配置占位。
type %s struct {
	Enabled bool `+"`mapstructure:\"enabled\"`"+`
}
`, typeName, capability, typeName)
}

func configYAMLTemplate(capability string) string {
	return fmt.Sprintf(`%s:
  enabled: false
`, capability)
}
