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
	require.NoDirExists(t, filepath.Join(target, "pkg"))
	require.Contains(t, readFile(t, filepath.Join(target, "cmd", "server", "main.go")), "example.com/demo/internal/boot")
	require.FileExists(t, filepath.Join(target, "internal", "boot", "app.go"))
	require.FileExists(t, filepath.Join(target, "internal", "boot", "providers.go"))
	require.FileExists(t, filepath.Join(target, "internal", "boot", "lifecycle.go"))
	require.DirExists(t, filepath.Join(target, "internal", "module"))
	for _, moduleName := range []string{"auth", "user"} {
		moduleDir := filepath.Join(target, "internal", "module", moduleName)
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
	require.DirExists(t, filepath.Join(target, "internal", "ent", "schema"))
	require.FileExists(t, filepath.Join(target, "internal", "ent", "client.go"))
	require.FileExists(t, filepath.Join(target, "internal", "ent", "migrate", "schema.go"))
	require.FileExists(t, filepath.Join(target, "internal", "data", "ent_client.go"))
	require.FileExists(t, filepath.Join(target, "internal", "data", "tx.go"))
	require.FileExists(t, filepath.Join(target, "configs", "config.yaml"))
	require.FileExists(t, filepath.Join(target, "configs", "config.dev.yaml"))
	require.NotContains(t, readFile(t, filepath.Join(target, "configs", "config.yaml")), "env:")
	require.NotContains(t, readFile(t, filepath.Join(target, "configs", "config.dev.yaml")), "env:")
	require.FileExists(t, filepath.Join(target, "scripts", "ent.ps1"))
	require.NoDirExists(t, filepath.Join(target, "internal", "gen", "jet"))
	require.NoDirExists(t, filepath.Join(target, "tools", "jetgen"))
	require.NoFileExists(t, filepath.Join(target, "scripts", "jet"+".ps1"))
	require.NoDirExists(t, filepath.Join(target, "internal", "app"))
	require.DirExists(t, filepath.Join(target, ".git"))
	require.Contains(t, stdout.String(), "created")
}

func TestNewUsesReleaseCLIVersionWhenFrameworkVersionOmitted(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")

	err := run([]string{
		"new", target,
		"--type", "worker",
		"--module", "example.com/demo",
	}, ioDiscard{}, "v1.2.3")

	require.NoError(t, err)
	goMod := readFile(t, filepath.Join(target, "go.mod"))
	require.Contains(t, goMod, "github.com/teamsillybees/initra v1.2.3")
	require.NotContains(t, goMod, "replace github.com/teamsillybees/initra")
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

	require.NoFileExists(t, filepath.Join(target, "internal", "ent", "migrate", "main.go"))
	diffGenerator := readFile(t, filepath.Join(target, "internal", "ent", "migratediff", "main.go"))
	require.Contains(t, diffGenerator, `_ "github.com/lib/pq"`)
	require.Contains(t, diffGenerator, "boot.LoadConfig")
	require.Contains(t, diffGenerator, "databaseURL")
	require.Contains(t, diffGenerator, "migrate.WithForeignKeys(false)")
	require.Contains(t, diffGenerator, "schema.WithMigrationMode(schema.ModeReplay)")

	entSchema := readFile(t, filepath.Join(target, "internal", "ent", "schema", "sys_user.go"))
	require.Contains(t, entSchema, "entsql.WithComments(true)")
	require.Contains(t, entSchema, `schema.Comment("系统后台用户表，用于后台登录、审计和权限归属。")`)

	migrateSchema := readFile(t, filepath.Join(target, "internal", "ent", "migrate", "schema.go"))
	require.Contains(t, migrateSchema, `"系统后台用户表，用于后台登录、审计和权限归属。"`)
	require.Contains(t, migrateSchema, `Comment: "登录用户名，全局唯一。"`)

	atlasScript := readFile(t, filepath.Join(target, "scripts", "atlas.ps1"))
	require.Contains(t, atlasScript, "go run ./internal/ent/migratediff/main.go")
	require.Contains(t, atlasScript, "-config-dir")
	require.NotContains(t, atlasScript, "docker://postgres/16/dev")
}

func TestNewGeneratesWorkerPlaceholderProject(t *testing.T) {
	target := filepath.Join(t.TempDir(), "worker")

	err := run([]string{
		"new", target,
		"--type", "worker",
		"--module", "example.com/worker",
		"--replace", repoRoot(t),
	}, ioDiscard{}, "dev")

	require.NoError(t, err)
	require.Contains(t, readFile(t, filepath.Join(target, "go.mod")), "module example.com/worker")
	require.FileExists(t, filepath.Join(target, "cmd", "worker", "main.go"))
	require.FileExists(t, filepath.Join(target, "internal", "worker", "worker.go"))
	require.NoFileExists(t, filepath.Join(target, "cmd", "server", "main.go"))
}

func TestAPITemplateExcludesEntGeneratedCode(t *testing.T) {
	root := repoRoot(t)
	templateDir := filepath.Join(root, "templates", "api", "internal", "ent")

	require.FileExists(t, filepath.Join(templateDir, "generate.go.tmpl"))
	require.FileExists(t, filepath.Join(templateDir, "schema", "sys_user.go.tmpl"))
	require.FileExists(t, filepath.Join(templateDir, "migratediff", "main.go.tmpl"))
	require.NoFileExists(t, filepath.Join(templateDir, "client.go.tmpl"))
	require.NoFileExists(t, filepath.Join(templateDir, "migrate", "schema.go.tmpl"))
	require.NoFileExists(t, filepath.Join(templateDir, "sysuser.go.tmpl"))
	require.NoFileExists(t, filepath.Join(templateDir, "sysuser", "where.go.tmpl"))
}

func TestRootHelpListsCobraCommands(t *testing.T) {
	var stdout bytes.Buffer

	err := run([]string{"--help"}, &stdout, "v1.2.3")

	require.NoError(t, err)
	require.Contains(t, stdout.String(), "Usage:")
	require.Contains(t, stdout.String(), "new")
	require.Contains(t, stdout.String(), "doctor")
}

func TestRootRejectsMissingCommand(t *testing.T) {
	err := run(nil, ioDiscard{}, "dev")

	require.Error(t, err)
	require.Contains(t, err.Error(), "用法: initra <command>")
}

func TestNewAcceptsFlagsBeforeTarget(t *testing.T) {
	target := filepath.Join(t.TempDir(), "worker")

	err := run([]string{
		"new",
		"--type", "worker",
		"--module", "example.com/worker",
		"--replace", repoRoot(t),
		target,
	}, ioDiscard{}, "dev")

	require.NoError(t, err)
	require.Contains(t, readFile(t, filepath.Join(target, "go.mod")), "module example.com/worker")
	require.FileExists(t, filepath.Join(target, "cmd", "worker", "main.go"))
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
	moduleDir := filepath.Join(target, "internal", "module", "order")
	for _, suffix := range []string{"handler", "service", "repo", "model", "dto", "routes"} {
		require.FileExists(t, filepath.Join(moduleDir, "order."+suffix+".go"))
	}
	require.FileExists(t, filepath.Join(moduleDir, "providers.go"))
	require.FileExists(t, filepath.Join(moduleDir, "order_test.go"))
	handlerContent := readFile(t, filepath.Join(moduleDir, "order.handler.go"))
	require.NotContains(t, handlerContent, "type GetOrderInput")
	require.NotContains(t, handlerContent, "type OrderResponse")
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
	require.NoError(t, os.MkdirAll(filepath.Join(target, "internal", "module", "order"), 0o755))
	t.Chdir(target)

	err := run([]string{"crud", "add", "order", "--table", "sys_order"}, ioDiscard{}, "dev")

	require.NoError(t, err)
	content := readFile(t, filepath.Join(target, "internal", "module", "order", "order.crud.go"))
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

	require.NoError(t, run([]string{"migrate", "diff", "add_order"}, ioDiscard{}, "dev"))
	diffScript := readFile(t, filepath.Join(target, "scripts", "migrate-diff-add_order.ps1"))
	require.Contains(t, diffScript, "go run ./internal/ent/migratediff/main.go add_order")
	require.Contains(t, diffScript, "-config-dir $ConfigDir")
	require.Contains(t, diffScript, `"-dev-url", $DevURL`)
	require.NotContains(t, diffScript, "docker://postgres/16/dev")
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
