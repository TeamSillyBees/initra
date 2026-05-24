package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGeneratesAPIProjectWithFrameworkRequireAndNoPkgCopy(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")

	var stdout bytes.Buffer
	err := run([]string{
		"new", target,
		"--type", "api",
		"--module", "example.com/demo",
		"--framework-version", "v1.2.3",
		"--replace", repoRoot(t),
	}, &stdout, "dev")

	require.NoError(t, err)
	goMod := readFile(t, filepath.Join(target, "go.mod"))
	require.Contains(t, goMod, "module example.com/demo")
	require.Contains(t, goMod, "github.com/teamsillybees/initra v1.2.3")
	require.FileExists(t, filepath.Join(target, ".gitignore"))
	require.FileExists(t, filepath.Join(target, "AGENTS.md"))
	require.NoDirExists(t, filepath.Join(target, "pkg"))
	agentInstructions := readFile(t, filepath.Join(target, "AGENTS.md"))
	require.Contains(t, agentInstructions, "业务代码不得 import `github.com/teamsillybees/initra/internal/*`")
	require.Contains(t, agentInstructions, "业务代码按业务模块组织为单一 flat package")
	require.Contains(t, readFile(t, filepath.Join(target, "cmd", "server", "main.go")), "example.com/demo/internal/boot")
	require.FileExists(t, filepath.Join(target, "internal", "boot", "app.go"))
	require.FileExists(t, filepath.Join(target, "internal", "boot", "providers.go"))
	require.FileExists(t, filepath.Join(target, "internal", "boot", "lifecycle.go"))
	require.DirExists(t, filepath.Join(target, "internal", "modules"))
	for _, moduleName := range []string{"auth", "user"} {
		moduleDir := filepath.Join(target, "internal", "modules", moduleName)
		require.DirExists(t, moduleDir)
		for _, suffix := range []string{"handler", "service", "repo", "model", "dto", "routes"} {
			require.FileExists(t, filepath.Join(moduleDir, moduleName+"."+suffix+".go"))
		}
		require.FileExists(t, filepath.Join(moduleDir, "providers.go"))
		handlerContent := readFile(t, filepath.Join(moduleDir, moduleName+".handler.go"))
		require.NotContains(t, handlerContent, "Input struct")
		require.NotContains(t, handlerContent, "Output struct")
		require.NotContains(t, handlerContent, "Response struct")
	}
	require.FileExists(t, filepath.Join(target, "db", "schema", "01_sys_user.sql"))
	require.FileExists(t, filepath.Join(target, "db", "seeds", "001_seed_admin.sql"))
	require.DirExists(t, filepath.Join(target, "internal", "data", "schema"))
	require.FileExists(t, filepath.Join(target, "internal", "data", "ent", "client.go"))
	require.FileExists(t, filepath.Join(target, "internal", "data", "ent", "migrate", "schema.go"))
	require.FileExists(t, filepath.Join(target, "internal", "data", "ent_client.go"))
	require.FileExists(t, filepath.Join(target, "internal", "data", "tx.go"))
	require.FileExists(t, filepath.Join(target, "configs", "config.yaml"))
	require.FileExists(t, filepath.Join(target, "configs", "config.dev.yaml"))
	require.NotContains(t, readFile(t, filepath.Join(target, "configs", "config.yaml")), "env:")
	require.NotContains(t, readFile(t, filepath.Join(target, "configs", "config.dev.yaml")), "env:")
	require.NoDirExists(t, filepath.Join(target, "scripts"))
	require.NoDirExists(t, filepath.Join(target, "internal", "gen", "jet"))
	require.NoDirExists(t, filepath.Join(target, "tools", "jetgen"))
	require.NoDirExists(t, filepath.Join(target, "internal", "app"))
	require.DirExists(t, filepath.Join(target, ".git"))
	require.Contains(t, stdout.String(), "created")
}

func TestNewUsesReleaseCLIVersionWhenFrameworkVersionOmitted(t *testing.T) {
	version, err := resolveFrameworkVersion("", "v1.2.3", "")

	require.NoError(t, err)
	require.Equal(t, "v1.2.3", version)
}

func TestNewGeneratesAPIMigrationsWithoutPhysicalForeignKeys(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")

	err := run([]string{
		"new", target,
		"--type", "api",
		"--module", "example.com/demo",
		"--replace", repoRoot(t),
	}, ioDiscard{}, "dev")

	require.NoError(t, err)
	migrations, err := filepath.Glob(filepath.Join(target, "db", "migrations", "*_init.sql"))
	require.NoError(t, err)
	require.Len(t, migrations, 1)
	migrationContent := readFile(t, migrations[0])
	require.NotContains(t, migrationContent, "FOREIGN KEY")
	require.NotContains(t, migrationContent, "REFERENCES")

	require.NoFileExists(t, filepath.Join(target, "internal", "data", "ent", "migrate", "main.go"))
	diffGenerator := readFile(t, filepath.Join(target, "internal", "data", "migratediff", "main.go"))
	require.Contains(t, diffGenerator, `_ "github.com/lib/pq"`)
	require.Contains(t, diffGenerator, "boot.LoadConfig")
	require.Contains(t, diffGenerator, "databaseURL")
	require.Contains(t, diffGenerator, "migrate.WithForeignKeys(false)")
	require.Contains(t, diffGenerator, "schema.WithMigrationMode(schema.ModeReplay)")

	entSchema := readFile(t, filepath.Join(target, "internal", "data", "schema", "sys_user.go"))
	require.Contains(t, entSchema, "entsql.WithComments(true)")
	require.Contains(t, entSchema, `schema.Comment("系统后台用户表，用于后台登录、审计和权限归属。")`)

	migrateSchema := readFile(t, filepath.Join(target, "internal", "data", "ent", "migrate", "schema.go"))
	require.Contains(t, migrateSchema, `"系统后台用户表，用于后台登录、审计和权限归属。"`)
	require.Contains(t, migrateSchema, `Comment: "登录用户名，全局唯一。"`)

	require.NoDirExists(t, filepath.Join(target, "scripts"))
}

func TestAPITemplateExcludesEntGeneratedCode(t *testing.T) {
	root := repoRoot(t)
	dataTemplateDir := filepath.Join(root, "templates", "api", "internal", "data")
	entTemplateDir := filepath.Join(dataTemplateDir, "ent")

	require.FileExists(t, filepath.Join(dataTemplateDir, "generate.go.tmpl"))
	require.FileExists(t, filepath.Join(dataTemplateDir, "schema", "sys_user.go.tmpl"))
	require.FileExists(t, filepath.Join(dataTemplateDir, "migratediff", "main.go.tmpl"))
	require.NoFileExists(t, filepath.Join(entTemplateDir, "client.go.tmpl"))
	require.NoFileExists(t, filepath.Join(entTemplateDir, "migrate", "schema.go.tmpl"))
	require.NoFileExists(t, filepath.Join(entTemplateDir, "sysuser.go.tmpl"))
	require.NoFileExists(t, filepath.Join(entTemplateDir, "sysuser", "where.go.tmpl"))
}

func TestRootHelpListsCobraCommands(t *testing.T) {
	var stdout bytes.Buffer

	err := run([]string{"--help"}, &stdout, "v1.2.3")

	require.NoError(t, err)
	require.Contains(t, stdout.String(), "用法:")
	require.Contains(t, stdout.String(), "可用命令:")
	require.Contains(t, stdout.String(), "completion")
	require.Contains(t, stdout.String(), "生成 shell 自动补全脚本")
	require.Contains(t, stdout.String(), "new")
	require.Contains(t, stdout.String(), "help")
	require.Contains(t, stdout.String(), "doctor")
}

func TestHelpCommandShowsSubcommandHelp(t *testing.T) {
	var stdout bytes.Buffer

	err := run([]string{"help", "new"}, &stdout, "v1.2.3")

	require.NoError(t, err)
	require.Contains(t, stdout.String(), "initra new <dir>")
	require.Contains(t, stdout.String(), "示例:")
	require.Contains(t, stdout.String(), "--module")
}

func TestSkillCodexCopiesInitraFrameworkSkill(t *testing.T) {
	target := t.TempDir()
	t.Chdir(target)

	var stdout bytes.Buffer
	err := run([]string{"skill", "codex"}, &stdout, "dev")

	require.NoError(t, err)
	skillDir := filepath.Join(target, ".agents", "skills", "initra-framework")
	require.FileExists(t, filepath.Join(skillDir, "SKILL.md"))
	require.FileExists(t, filepath.Join(skillDir, "assets", "capabilities.yaml"))
	require.FileExists(t, filepath.Join(skillDir, "references", "redisx.md"))
	require.FileExists(t, filepath.Join(skillDir, "scripts", "check_initra_usage.go"))
	require.Contains(t, stdout.String(), "created skill")
}

func TestSkillClaudeCodeCopiesInitraFrameworkSkill(t *testing.T) {
	target := t.TempDir()
	t.Chdir(target)

	var stdout bytes.Buffer
	err := run([]string{"skill", "cc"}, &stdout, "dev")

	require.NoError(t, err)
	skillDir := filepath.Join(target, ".claude", "skills", "initra-framework")
	require.FileExists(t, filepath.Join(skillDir, "SKILL.md"))
	require.FileExists(t, filepath.Join(skillDir, "assets", "capabilities.yaml"))
	require.FileExists(t, filepath.Join(skillDir, "references", "redisx.md"))
	require.FileExists(t, filepath.Join(skillDir, "scripts", "check_initra_usage.go"))
	require.Contains(t, stdout.String(), "created skill")
}

func TestSkillInitCommandRemoved(t *testing.T) {
	err := run([]string{"skill", "init"}, ioDiscard{}, "dev")

	require.Error(t, err)
	require.Contains(t, err.Error(), "不接受位置参数")
}

func TestRootWithoutArgsShowsHelp(t *testing.T) {
	var stdout bytes.Buffer

	err := run(nil, &stdout, "dev")

	require.NoError(t, err)
	require.Contains(t, stdout.String(), "用法:")
	require.Contains(t, stdout.String(), "initra [command]")
}

func TestGroupCommandWithoutSubcommandShowsHelp(t *testing.T) {
	var stdout bytes.Buffer

	err := run([]string{"module"}, &stdout, "dev")

	require.NoError(t, err)
	require.Contains(t, stdout.String(), "initra module [flags]")
	require.Contains(t, stdout.String(), "add")
}

func TestNewRejectsMissingTargetWithHelpHint(t *testing.T) {
	err := run([]string{"new"}, ioDiscard{}, "dev")

	require.Error(t, err)
	require.Contains(t, err.Error(), "需要 1 个目标目录参数")
	require.Contains(t, err.Error(), "initra new --help")
}

func TestNewAcceptsFlagsBeforeTarget(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")

	err := run([]string{
		"new",
		"--type", "api",
		"--module", "example.com/demo",
		"--replace", repoRoot(t),
		target,
	}, ioDiscard{}, "dev")

	require.NoError(t, err)
	require.Contains(t, readFile(t, filepath.Join(target, "go.mod")), "module example.com/demo")
	require.FileExists(t, filepath.Join(target, "cmd", "server", "main.go"))
	require.DirExists(t, filepath.Join(target, ".git"))
}

func TestNewWritesLocalReplaceWhenRequested(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")
	frameworkPath := repoRoot(t)

	err := run([]string{
		"new", target,
		"--type", "api",
		"--module", "example.com/demo",
		"--replace", frameworkPath,
	}, ioDiscard{}, "dev")

	require.NoError(t, err)
	goMod := readFile(t, filepath.Join(target, "go.mod"))
	require.Contains(t, goMod, "github.com/teamsillybees/initra v0.0.0")
	require.Contains(t, goMod, "replace github.com/teamsillybees/initra => "+filepath.ToSlash(frameworkPath))
}

func TestNewRejectsDevVersionWithoutFrameworkVersionOrReplace(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")

	err := run([]string{"new", target, "--type", "api", "--module", "example.com/demo"}, ioDiscard{}, "dev")

	require.Error(t, err)
	require.Contains(t, err.Error(), "--framework-version")
}

func TestResolveCLIVersionPrefersInjectedReleaseVersion(t *testing.T) {
	actual := resolveCLIVersion("v1.2.3", "v9.9.9")

	require.Equal(t, "v1.2.3", actual)
}

func TestResolveCLIVersionUsesBuildInfoVersionForOnlineInstall(t *testing.T) {
	actual := resolveCLIVersion("dev", "v1.2.3")

	require.Equal(t, "v1.2.3", actual)
}

func TestResolveCLIVersionKeepsDevForLocalBuild(t *testing.T) {
	actual := resolveCLIVersion("dev", "(devel)")

	require.Equal(t, "dev", actual)
}

func TestNewRejectsUnknownProjectType(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")

	err := run([]string{"new", target, "--type", "batch", "--module", "example.com/demo", "--framework-version", "v1.2.3"}, ioDiscard{}, "dev")

	require.Error(t, err)
	require.Contains(t, err.Error(), "暂不支持项目类型")
}

func TestNewRejectsRemovedWorkerProjectType(t *testing.T) {
	target := filepath.Join(t.TempDir(), "worker")

	err := run([]string{"new", target, "--type", "worker", "--module", "example.com/worker", "--framework-version", "v1.2.3"}, ioDiscard{}, "dev")

	require.Error(t, err)
	require.Contains(t, err.Error(), `暂不支持项目类型 "worker"`)
	require.Contains(t, err.Error(), "仅支持 api")
}

func TestGeneratedProjectBuildsWithLocalReplace(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")

	err := run([]string{
		"new", target,
		"--type", "api",
		"--module", "example.com/demo",
		"--replace", repoRoot(t),
	}, ioDiscard{}, "dev")
	require.NoError(t, err)

	testCmd := exec.Command("go", "test", "./...")
	testCmd.Dir = target
	testOutput, err := testCmd.CombinedOutput()
	require.NoError(t, err, string(testOutput))

	buildCmd := exec.Command("go", "build", "./cmd/server")
	buildCmd.Dir = target
	buildOutput, err := buildCmd.CombinedOutput()
	require.NoError(t, err, string(buildOutput))
}

func TestModuleAddGeneratesVerticalSliceModule(t *testing.T) {
	target := t.TempDir()
	t.Chdir(target)

	var stdout bytes.Buffer
	err := run([]string{"module", "add", "order"}, &stdout, "dev")

	require.NoError(t, err)
	moduleDir := filepath.Join(target, "internal", "modules", "order")
	for _, suffix := range []string{"handler", "service", "repo", "model", "dto", "routes"} {
		require.FileExists(t, filepath.Join(moduleDir, "order."+suffix+".go"))
	}
	require.FileExists(t, filepath.Join(moduleDir, "providers.go"))
	require.FileExists(t, filepath.Join(moduleDir, "order_test.go"))
	handlerContent := readFile(t, filepath.Join(moduleDir, "order.handler.go"))
	require.NotContains(t, handlerContent, "type GetOrderInput")
	require.NotContains(t, handlerContent, "type OrderResponse")
	routesContent := readFile(t, filepath.Join(moduleDir, "order.routes.go"))
	require.Contains(t, routesContent, "AccessMode: platformauth.AccessModePermission")
	dtoContent := readFile(t, filepath.Join(moduleDir, "order.dto.go"))
	require.Contains(t, dtoContent, "type getOrderRequest")
	require.Contains(t, dtoContent, "type getOrderResponse")
	require.Contains(t, dtoContent, "type OrderVO")
	require.NotContains(t, dtoContent, "Input")
	require.NotContains(t, dtoContent, "Output")
	require.NotContains(t, dtoContent, "type OrderResponse")
	require.Contains(t, stdout.String(), "created module order")
}

func TestCRUDAddGeneratesSampleForModule(t *testing.T) {
	target := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(target, "internal", "modules", "order"), 0o755))
	t.Chdir(target)

	err := run([]string{"crud", "add", "order", "--table", "sys_order"}, ioDiscard{}, "dev")

	require.NoError(t, err)
	content := readFile(t, filepath.Join(target, "internal", "modules", "order", "order.crud.go"))
	require.Contains(t, content, "sys_order")
	require.Contains(t, content, "type OrderCRUD")
}

func TestConfigAddGeneratesConfigSnippet(t *testing.T) {
	target := t.TempDir()
	t.Chdir(target)

	err := run([]string{"config", "add", "redis"}, ioDiscard{}, "dev")

	require.NoError(t, err)
	require.FileExists(t, filepath.Join(target, "internal", "boot", "redis.config.go"))
	require.FileExists(t, filepath.Join(target, "configs", "redis.yaml"))
}

func TestMigrateCommandsGenerateFiles(t *testing.T) {
	target := t.TempDir()
	t.Chdir(target)

	require.NoError(t, run([]string{"migrate", "new", "create_order"}, ioDiscard{}, "dev"))
	migrations, err := filepath.Glob(filepath.Join(target, "db", "migrations", "*_create_order.sql"))
	require.NoError(t, err)
	require.Len(t, migrations, 1)
}

func TestMigrateHelpListsApplyCommand(t *testing.T) {
	var stdout bytes.Buffer

	err := run([]string{"migrate", "--help"}, &stdout, "dev")

	require.NoError(t, err)
	require.Contains(t, stdout.String(), "apply")
	require.Contains(t, stdout.String(), "应用 Atlas 迁移")
	require.Contains(t, stdout.String(), "hash")
	require.Contains(t, stdout.String(), "重算 Atlas 迁移校验和")
}

func TestBuildMigrateDiffArgs(t *testing.T) {
	actual := buildMigrateDiffArgs("add_order", migrateDiffOptions{
		env:       "local",
		configDir: "configs",
		devURL:    "postgres://dev",
	})

	require.Equal(t, []string{
		"run",
		"./internal/data/migratediff/main.go",
		"add_order",
		"-config-dir",
		"configs",
		"-env",
		"local",
		"-dev-url",
		"postgres://dev",
	}, actual)
}

func TestBuildMigrateApplyArgs(t *testing.T) {
	actual := buildMigrateApplyArgs(migrateApplyOptions{env: "local"})

	require.Equal(t, []string{
		"-c",
		"file://db/atlas.hcl",
		"migrate",
		"apply",
		"--env",
		"local",
	}, actual)
}

func TestMigrateApplyRequiresEnv(t *testing.T) {
	err := run([]string{"migrate", "apply"}, ioDiscard{}, "dev")

	require.Error(t, err)
	require.Contains(t, err.Error(), "必须提供 --env")
	require.Contains(t, err.Error(), "initra migrate apply --help")
}

func TestBuildMigrateHashArgs(t *testing.T) {
	actual := buildMigrateHashArgs(migrateHashOptions{env: "local"})

	require.Equal(t, []string{
		"-c",
		"file://db/atlas.hcl",
		"migrate",
		"hash",
		"--env",
		"local",
	}, actual)
}

func TestMigrateHashDefaultsToLocalEnv(t *testing.T) {
	var stdout bytes.Buffer

	err := run([]string{"migrate", "hash", "--help"}, &stdout, "dev")

	require.NoError(t, err)
	require.Contains(t, stdout.String(), `--env string`)
	require.Contains(t, stdout.String(), `(default "local")`)
}

func TestDoctorReportsEnvironmentChecks(t *testing.T) {
	var stdout bytes.Buffer

	err := run([]string{"doctor"}, &stdout, "dev")

	require.NoError(t, err)
	require.Contains(t, stdout.String(), "Go:")
	require.Contains(t, stdout.String(), "Atlas:")
	require.Contains(t, stdout.String(), "Ent:")
	require.Contains(t, stdout.String(), "golangci-lint:")
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			content := readFile(t, filepath.Join(dir, "go.mod"))
			if strings.Contains(content, "module github.com/teamsillybees/initra") {
				return filepath.ToSlash(dir)
			}
		}
		parent := filepath.Dir(dir)
		require.NotEqual(t, dir, parent, "repo root not found")
		dir = parent
	}
}
