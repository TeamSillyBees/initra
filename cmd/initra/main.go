package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

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

var version = "dev"

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
		Long:          "initra 用于生成和维护企业内部 Go 服务脚手架，覆盖 API 项目、业务模块、代码片段、配置能力和迁移辅助文件。",
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
		newSnippetCommand(stdout),
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
