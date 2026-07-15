package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

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
		Long:          "调用当前项目的 migratediff 入口生成迁移；默认按 env 和 config-dir 读取业务数据库配置，也可用 dev-url 显式覆盖。",
		Example:       "  initra migrate diff add_order --env local --config-dir configs\n  initra migrate diff add_order --dev-url postgres://dev",
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
	flags.StringVar(&opts.devURL, "dev-url", "", "可选的数据库 URL 覆盖值，传递给 migratediff")
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
