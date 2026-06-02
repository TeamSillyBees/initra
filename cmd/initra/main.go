package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
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
	skillassets "github.com/teamsillybees/initra/tools/skills"
)

const frameworkModule = "github.com/teamsillybees/initra"

// commandHelpTemplate 是所有命令共用的中文帮助模板。
const commandHelpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{.UsageString}}`

// commandUsageTemplate 是所有命令共用的中文用法模板。
const commandUsageTemplate = `用法:
  {{.UseLine}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if .HasAvailableSubCommands}}

可用命令:
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

选项:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

全局选项:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasExample}}

示例:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

运行 "{{.CommandPath}} [command] --help" 查看子命令帮助。{{end}}
`

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

type migrateDiffOptions struct {
	env       string
	configDir string
	devURL    string
}

type migrateApplyOptions struct {
	env string
}

type migrateHashOptions struct {
	env string
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
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
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
		Long:          "initra 用于生成和维护企业内部 Go 服务脚手架，覆盖 API 项目、业务模块、CRUD 示例、配置片段和迁移辅助文件。",
		Example:       "  initra new ./demo --type api --module example.com/demo\n  initra module add order\n  initra doctor",
		Version:       cliVersion,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCommandHelp(cmd)
		},
	}
	configureCommand(cmd, stdout)
	cmd.SetHelpCommand(newHelpCommand(stdout))
	cmd.AddCommand(
		newNewCommand(stdout, cliVersion),
		newModuleCommand(stdout),
		newCRUDCommand(stdout),
		newConfigCommand(stdout),
		newMigrateCommand(stdout),
		newSkillCommand(stdout),
		newDoctorCommand(stdout),
	)
	localizeCompletionCommand(cmd, stdout)
	return cmd
}

func configureCommand(cmd *cobra.Command, stdout io.Writer) {
	if stdout == nil {
		stdout = io.Discard
	}
	cmd.SetOut(stdout)
	cmd.SetErr(io.Discard)
	cmd.SetHelpTemplate(commandHelpTemplate)
	cmd.SetUsageTemplate(commandUsageTemplate)
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return commandUsageError(cmd, "参数错误：%v", err)
	})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SuggestionsMinimumDistance = 2
}

// newHelpCommand 创建显式 help 命令，并复用 Cobra 的命令查找能力展示目标命令帮助。
func newHelpCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "help [command]",
		Short:         "查看命令帮助",
		Long:          "查看 initra 或指定子命令的帮助信息。",
		Example:       "  initra help\n  initra help new\n  initra help migrate diff",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			if len(args) == 0 {
				return showCommandHelp(root)
			}
			target, _, err := root.Find(args)
			if err != nil || target == nil {
				return commandUsageError(cmd, "未知命令 %q", strings.Join(args, " "))
			}
			return showCommandHelp(target)
		},
	}
	configureCommand(cmd, stdout)
	return cmd
}

// showCommandHelp 输出命令帮助，供根命令和分组命令的默认动作复用。
func showCommandHelp(cmd *cobra.Command) error {
	return cmd.Help()
}

// commandUsageError 构造包含 help 指引的命令参数错误。
func commandUsageError(cmd *cobra.Command, format string, values ...any) error {
	message := fmt.Sprintf(format, values...)
	return fmt.Errorf("%s\n\n运行 %q 查看帮助", message, cmd.CommandPath()+" --help")
}

// requireExactArgs 返回中文化的固定参数数量校验器。
func requireExactArgs(count int, noun string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == count {
			return nil
		}
		return commandUsageError(cmd, "参数错误：需要 %d 个%s参数，收到 %d 个", count, noun, len(args))
	}
}

// requireNoArgs 校验命令不接受位置参数。
func requireNoArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	return commandUsageError(cmd, "参数错误：不接受位置参数，收到 %d 个", len(args))
}

// completeValues 返回固定候选值，并禁用文件名补全。
func completeValues(values ...string) cobra.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return values, cobra.ShellCompDirectiveNoFileComp
	}
}

// localizeCompletionCommand 初始化 Cobra 默认 completion 命令，并改成项目一致的中文帮助。
func localizeCompletionCommand(root *cobra.Command, stdout io.Writer) {
	root.InitDefaultCompletionCmd()
	completion, _, err := root.Find([]string{"completion"})
	if err != nil || completion == nil || completion.Name() != "completion" {
		return
	}
	configureCommand(completion, stdout)
	completion.Short = "生成 shell 自动补全脚本"
	completion.Long = "生成 bash、zsh、fish 或 powershell 自动补全脚本。"
	completion.Example = "  initra completion powershell\n  initra completion bash"
	completion.RunE = func(cmd *cobra.Command, args []string) error {
		return showCommandHelp(cmd)
	}
	for _, child := range completion.Commands() {
		configureCommand(child, stdout)
		child.Short = fmt.Sprintf("生成 %s 自动补全脚本", completionShellName(child.Name()))
		child.Long = child.Short + "。"
	}
}

// completionShellName 返回用于帮助文案展示的 shell 名称。
func completionShellName(name string) string {
	switch name {
	case "powershell":
		return "PowerShell"
	default:
		return name
	}
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

func newModuleCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "module",
		Short:         "管理业务模块骨架",
		Long:          "管理标准项目的 internal/modules/<name> 业务模块骨架。模块按 flat package 组织，包含 handler、service、repo、model、dto、routes、providers 和测试文件。",
		Example:       "  initra module add order",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCommandHelp(cmd)
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
		Long:          "在当前项目的 internal/modules 下生成一个业务模块骨架，模块名必须是合法 Go package 名称。",
		Example:       "  initra module add order",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireExactArgs(1, "模块名"),
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
		Long:          "为已存在的业务模块追加 CRUD 示例文件，用于快速展示标准模块内的简单数据访问写法。",
		Example:       "  initra crud add order --table sys_order",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCommandHelp(cmd)
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
		Long:          "在现有业务模块目录中生成 <module>.crud.go 示例文件。目标模块必须已经通过 initra module add 或手动创建。",
		Example:       "  initra crud add order --table sys_order",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireExactArgs(1, "模块名"),
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
		Long:          "为标准项目追加配置结构和 YAML 示例片段，便于把框架能力显式接入 internal/boot 配置。",
		Example:       "  initra config add redis",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCommandHelp(cmd)
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
		Long:          "生成 internal/boot/<capability>.config.go 和 configs/<capability>.yaml，作为接入框架能力的配置起点。",
		Example:       "  initra config add redis\n  initra config add storage",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireExactArgs(1, "能力名"),
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
		Long:          "管理 Ent/Atlas 迁移辅助文件。new 创建空迁移文件，diff 调用当前项目的 migratediff 入口生成 schema diff，apply 应用已有迁移，hash 重算迁移目录校验和。",
		Example:       "  initra migrate new create_order\n  initra migrate diff add_order --env local --config-dir configs\n  initra migrate apply --env local\n  initra migrate hash",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCommandHelp(cmd)
		},
	}
	configureCommand(cmd, stdout)
	cmd.AddCommand(newMigrateNewCommand(stdout), newMigrateDiffCommand(stdout), newMigrateApplyCommand(stdout), newMigrateHashCommand(stdout))
	return cmd
}

func newMigrateNewCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "new <name>",
		Short:         "创建空迁移文件",
		Long:          "在 db/migrations 下创建一个带时间戳的空 SQL 迁移文件。",
		Example:       "  initra migrate new create_order",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireExactArgs(1, "迁移名"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return createMigrationArtifact("new", args[0], cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	return cmd
}

func newMigrateDiffCommand(stdout io.Writer) *cobra.Command {
	opts := migrateDiffOptions{}
	cmd := &cobra.Command{
		Use:           "diff <name>",
		Short:         "执行 Ent/Atlas 迁移 diff",
		Long:          "调用当前项目的 go run ./internal/data/migratediff/main.go <name>，并把 env、config-dir、dev-url 透传给 migratediff。",
		Example:       "  initra migrate diff add_order\n  initra migrate diff add_order --env local --config-dir configs --dev-url postgres://dev",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireExactArgs(1, "迁移名"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrationDiff(args[0], opts, cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	flags := cmd.Flags()
	flags.StringVar(&opts.env, "env", "", "运行环境，传递给 migratediff")
	flags.StringVar(&opts.configDir, "config-dir", "", "配置目录，传递给 migratediff")
	flags.StringVar(&opts.devURL, "dev-url", "", "Atlas dev database URL，传递给 migratediff")
	_ = cmd.RegisterFlagCompletionFunc("env", completeValues("dev", "test", "local", "prod"))
	return cmd
}

func newMigrateApplyCommand(stdout io.Writer) *cobra.Command {
	opts := migrateApplyOptions{}
	cmd := &cobra.Command{
		Use:           "apply",
		Short:         "应用 Atlas 迁移",
		Long:          "执行 atlas -c file://db/atlas.hcl migrate apply --env <env>，用于把 db/migrations 下的迁移应用到指定环境。",
		Example:       "  initra migrate apply --env local\n  initra migrate apply --env prod",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(opts.env) == "" {
				return commandUsageError(cmd, "参数错误：migrate apply 必须提供 --env")
			}
			return runMigrationApply(opts, cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	cmd.Flags().StringVar(&opts.env, "env", "", "Atlas 迁移环境，例如 local、dev、prod")
	_ = cmd.RegisterFlagCompletionFunc("env", completeValues("dev", "test", "local", "prod"))
	return cmd
}

func newMigrateHashCommand(stdout io.Writer) *cobra.Command {
	opts := migrateHashOptions{env: "local"}
	cmd := &cobra.Command{
		Use:           "hash",
		Short:         "重算 Atlas 迁移校验和",
		Long:          "执行 atlas -c file://db/atlas.hcl migrate hash --env <env>，用于手动修改 db/migrations 后重新计算 atlas.sum。",
		Example:       "  initra migrate hash\n  initra migrate hash --env dev",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrationHash(opts, cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	cmd.Flags().StringVar(&opts.env, "env", "local", "Atlas 配置环境，用于选择 migration.dir")
	_ = cmd.RegisterFlagCompletionFunc("env", completeValues("dev", "test", "local", "prod"))
	return cmd
}

func newSkillCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "skill",
		Short:         "管理 initra 相关 skill 文档",
		Long:          "管理 initra 框架相关的 skill 文档，用于在 Codex 或 Claude Code 中复用 initra 的开发约束和辅助脚本。",
		Example:       "  initra skill codex\n  initra skill cc",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCommandHelp(cmd)
		},
	}
	configureCommand(cmd, stdout)
	cmd.AddCommand(newSkillCodexCommand(stdout), newSkillClaudeCodeCommand(stdout))
	return cmd
}

func newSkillCodexCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "codex",
		Short:         "添加 Codex skill 文档",
		Long:          "在当前项目写入 .agents/skills/initra-framework，供 Codex 理解并检查 initra 项目约束。",
		Example:       "  initra skill codex",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initFrameworkSkill(filepath.Join(".agents", "skills", "initra-framework"), cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	return cmd
}

func newSkillClaudeCodeCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "cc",
		Aliases:       []string{"claude", "claude-code"},
		Short:         "添加 Claude Code skill 文档",
		Long:          "在当前项目写入 .claude/skills/initra-framework，供 Claude Code 理解并检查 initra 项目约束。",
		Example:       "  initra skill cc\n  initra skill claude",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initFrameworkSkill(filepath.Join(".claude", "skills", "initra-framework"), cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	return cmd
}

func newDoctorCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "doctor",
		Short:         "检查本地开发环境",
		Long:          "检查当前开发环境中 Go、Atlas、Ent、golangci-lint 和标准项目配置文件是否可用。",
		Example:       "  initra doctor",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctorChecks(cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	return cmd
}

func createProject(targetDir string, stdout io.Writer, cliVersion string, opts newOptions) error {
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

func generateEntCode(targetDir string) error {
	command := exec.Command("go", "run", "./internal/data/entgenerate")
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

func addModule(name string, stdout io.Writer) error {
	name, err := normalizeGoPackageName(name)
	if err != nil {
		return err
	}

	moduleDir := filepath.Join("internal", "modules", name)
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

func addCRUDSample(moduleName string, opts crudAddOptions, stdout io.Writer) error {
	moduleName, err := normalizeGoPackageName(moduleName)
	if err != nil {
		return err
	}

	table := strings.TrimSpace(opts.tableName)
	if table == "" {
		return fmt.Errorf("crud add 必须提供 --table")
	}

	moduleDir := filepath.Join("internal", "modules", moduleName)
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
	default:
		return fmt.Errorf("未知 migrate 子命令 %q", kind)
	}
}

func runMigrationDiff(name string, opts migrateDiffOptions, stdout io.Writer) error {
	name, err := normalizeSafeName(name)
	if err != nil {
		return err
	}
	command := exec.Command("go", buildMigrateDiffArgs(name, opts)...)
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return fmt.Errorf("生成迁移 diff 失败: %w", err)
		}
		return fmt.Errorf("生成迁移 diff 失败: %w: %s", err, message)
	}
	if stdout != nil {
		_, _ = fmt.Fprintf(stdout, "generated migration diff %s\n", name)
	}
	return nil
}

func runMigrationApply(opts migrateApplyOptions, stdout io.Writer) error {
	env, err := normalizeMigrationEnv(opts.env)
	if err != nil {
		return err
	}
	command := exec.Command("atlas", buildMigrateApplyArgs(migrateApplyOptions{env: env})...)
	output, err := command.CombinedOutput()
	message := strings.TrimSpace(string(output))
	if err != nil {
		if message == "" {
			return fmt.Errorf("应用迁移失败: %w", err)
		}
		return fmt.Errorf("应用迁移失败: %w: %s", err, message)
	}
	if stdout != nil {
		if message != "" {
			_, _ = fmt.Fprintln(stdout, message)
		}
		_, _ = fmt.Fprintf(stdout, "applied migrations for env %s\n", env)
	}
	return nil
}

func runMigrationHash(opts migrateHashOptions, stdout io.Writer) error {
	env, err := normalizeMigrationEnv(opts.env)
	if err != nil {
		return err
	}
	command := exec.Command("atlas", buildMigrateHashArgs(migrateHashOptions{env: env})...)
	output, err := command.CombinedOutput()
	message := strings.TrimSpace(string(output))
	if err != nil {
		if message == "" {
			return fmt.Errorf("重算迁移 hash 失败: %w", err)
		}
		return fmt.Errorf("重算迁移 hash 失败: %w: %s", err, message)
	}
	if stdout != nil {
		if message != "" {
			_, _ = fmt.Fprintln(stdout, message)
		}
		_, _ = fmt.Fprintf(stdout, "recomputed migration hash for env %s\n", env)
	}
	return nil
}

func buildMigrateDiffArgs(name string, opts migrateDiffOptions) []string {
	args := []string{"run", "./internal/data/migratediff/main.go", name}
	if configDir := strings.TrimSpace(opts.configDir); configDir != "" {
		args = append(args, "-config-dir", configDir)
	}
	if env := strings.TrimSpace(opts.env); env != "" {
		args = append(args, "-env", env)
	}
	if devURL := strings.TrimSpace(opts.devURL); devURL != "" {
		args = append(args, "-dev-url", devURL)
	}
	return args
}

func buildMigrateApplyArgs(opts migrateApplyOptions) []string {
	return []string{"-c", "file://db/atlas.hcl", "migrate", "apply", "--env", strings.TrimSpace(opts.env)}
}

func buildMigrateHashArgs(opts migrateHashOptions) []string {
	return []string{"-c", "file://db/atlas.hcl", "migrate", "hash", "--env", strings.TrimSpace(opts.env)}
}

func normalizeMigrationEnv(env string) (string, error) {
	env = strings.TrimSpace(env)
	if env == "" {
		return "", fmt.Errorf("migrate 命令必须提供 --env")
	}
	return normalizeSafeName(env)
}

func initFrameworkSkill(targetRoot string, stdout io.Writer) error {
	const sourceRoot = "initra-framework"
	err := fs.WalkDir(skillassets.FS, sourceRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(sourceRoot, filepath.ToSlash(path))
		if err != nil {
			return err
		}
		if relativePath == "." {
			return os.MkdirAll(targetRoot, 0o755)
		}
		outputPath := filepath.Join(targetRoot, filepath.FromSlash(relativePath))
		if entry.IsDir() {
			return os.MkdirAll(outputPath, 0o755)
		}
		content, err := skillassets.FS.ReadFile(path)
		if err != nil {
			return err
		}
		return writeNewFile(outputPath, string(content))
	})
	if err != nil {
		return err
	}
	if stdout != nil {
		_, _ = fmt.Fprintf(stdout, "created skill %s\n", targetRoot)
	}
	return nil
}

func runDoctorChecks(stdout io.Writer) error {
	if stdout == nil {
		stdout = io.Discard
	}

	reportTool(stdout, "Go", "go", "version")
	reportTool(stdout, "Atlas", "atlas", "version")
	reportTool(stdout, "Ent", "go", "run", "entgo.io/ent/cmd/ent", "--help")
	reportTool(stdout, "golangci-lint", "golangci-lint", "version")
	reportOptionalFile(stdout, "config.yaml", filepath.Join("configs", "config.yaml"))
	reportOptionalFile(stdout, "config.dev.yaml", filepath.Join("configs", "config.dev.yaml"))
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

// reportOptionalFile 输出可选文件状态，缺失时不表达为硬性环境问题。
func reportOptionalFile(stdout io.Writer, label string, path string) {
	if _, err := os.Stat(path); err != nil {
		_, _ = fmt.Fprintf(stdout, "%s: OPTIONAL MISSING %s\n", label, path)
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

import "github.com/teamsillybees/initra/pkg/idgen"

// %s 是 %s 模块的领域占位模型。
type %s struct {
	ID idgen.ID
}
`, name, typeName, name, typeName)
}

func moduleServiceTemplate(name string) string {
	typeName := exportedName(name)
	return fmt.Sprintf(`package %s

import (
	"context"

	"github.com/teamsillybees/initra/pkg/idgen"
)

// Service 是 %s 模块的应用服务。
type Service struct{}

// NewService 创建 %s 模块应用服务。
func NewService() *Service {
	return &Service{}
}

// Get 返回 %s 详情占位数据。
func (s *Service) Get(ctx context.Context, id idgen.ID) (*%s, error) {
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
	traceID, _ := requestctx.TraceIDFromContext(ctx)
	return &get%sResponse{
		Body: response.OK(traceID, %sVO{ID: item.ID}),
	}, nil
}
`, name, name, name, typeName, typeName, typeName, typeName)
}

func moduleDTOTemplate(name string) string {
	typeName := exportedName(name)
	return fmt.Sprintf(`package %s

import (
	"github.com/teamsillybees/initra/pkg/idgen"
	"github.com/teamsillybees/initra/pkg/response"
)

type get%sRequest struct {
	ID idgen.ID `+"`path:\"id\" example:\"1771234567890123456\" doc:\"ID\"`"+`
}

// %sVO 描述 %s 对外 JSON DTO。
type %sVO struct {
	ID idgen.ID `+"`json:\"id\"`"+`
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
	registry.Register(http.MethodGet, "%s", platformauth.RouteSecurity{AccessMode: platformauth.AccessModePermission, Resource: "%s", Action: "read"})
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
