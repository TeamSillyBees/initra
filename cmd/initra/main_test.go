package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
		for _, suffix := range []string{"handler", "service", "dto", "routes"} {
			require.FileExists(t, filepath.Join(moduleDir, moduleName+"."+suffix+".go"))
		}
		require.FileExists(t, filepath.Join(moduleDir, "providers.go"))
		handlerContent := readFile(t, filepath.Join(moduleDir, moduleName+".handler.go"))
		require.NotContains(t, handlerContent, "Input struct")
		require.NotContains(t, handlerContent, "Output struct")
		require.NotContains(t, handlerContent, "Response struct")
	}
	require.NoDirExists(t, filepath.Join(target, "db", "schema"))
	migrations, err := filepath.Glob(filepath.Join(target, "db", "migrations", "*_init.sql"))
	require.NoError(t, err)
	require.Len(t, migrations, 1)
	require.FileExists(t, filepath.Join(target, "db", "seeds", "001_seed_admin.sql"))
	require.DirExists(t, filepath.Join(target, "internal", "data", "schema"))
	require.FileExists(t, filepath.Join(target, "internal", "data", "ent", "client.go"))
	require.FileExists(t, filepath.Join(target, "internal", "data", "ent", "migrate", "schema.go"))
	require.FileExists(t, filepath.Join(target, "go.sum"))
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

// TestCreateProjectPreparesDependenciesBeforeEntAndCommitsOnceComplete 验证项目仅在全部生成步骤成功后落入目标目录。
func TestCreateProjectPreparesDependenciesBeforeEntAndCommitsOnceComplete(t *testing.T) {
	parent := t.TempDir()
	target := filepath.Join(parent, "demo")
	var calls []string
	var workDir string
	runner := func(dir string, name string, args ...string) ([]byte, error) {
		if workDir == "" {
			workDir = dir
		}
		require.Equal(t, workDir, dir)
		calls = append(calls, name+" "+strings.Join(args, " "))
		if name == "go" && strings.Join(args, " ") == "mod download all" {
			require.NoError(t, os.WriteFile(filepath.Join(dir, "go.sum"), []byte("prepared\n"), 0o644))
		}
		if name == "git" {
			require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))
		}
		return nil, nil
	}

	err := createProjectWithRunner(target, ioDiscard{}, "v1.2.3", newOptions{
		projectType: "api",
		modulePath:  "example.com/demo",
	}, runner)

	require.NoError(t, err)
	require.Equal(t, []string{
		"go mod download all",
		"go run ./internal/data/entgenerate",
		"go test ./... -count=1",
		"git init",
	}, calls)
	require.NotEqual(t, target, workDir)
	require.FileExists(t, filepath.Join(target, "go.sum"))
	require.DirExists(t, filepath.Join(target, ".git"))
	goMod := readFile(t, filepath.Join(target, "go.mod"))
	require.Contains(t, goMod, "github.com/teamsillybees/initra v1.2.3")
	require.NotContains(t, goMod, "replace github.com/teamsillybees/initra")
	if runtime.GOOS != "windows" {
		info, statErr := os.Stat(target)
		require.NoError(t, statErr)
		require.Equal(t, os.FileMode(0o755), info.Mode().Perm())
	}
	temporaryDirs, globErr := filepath.Glob(filepath.Join(parent, ".initra-new-*"))
	require.NoError(t, globErr)
	require.Empty(t, temporaryDirs)
}

// TestCreateProjectSupportsCurrentEmptyDirectory 验证 Windows 等平台可以把当前空目录安全替换为完整项目并恢复 cwd。
func TestCreateProjectSupportsCurrentEmptyDirectory(t *testing.T) {
	parent := t.TempDir()
	target := filepath.Join(parent, "current-project")
	require.NoError(t, os.Mkdir(target, 0o750))
	require.NoError(t, os.Chmod(target, 0o750))
	t.Chdir(target)
	runner := func(dir string, name string, args ...string) ([]byte, error) {
		if name == "go" && strings.Join(args, " ") == "mod download all" {
			require.NoError(t, os.WriteFile(filepath.Join(dir, "go.sum"), []byte("prepared\n"), 0o644))
		}
		if name == "git" {
			require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))
		}
		return nil, nil
	}

	err := createProjectWithRunner(".", ioDiscard{}, "v1.2.3", newOptions{
		projectType: "api",
		modulePath:  "example.com/current-project",
	}, runner)

	require.NoError(t, err)
	cwdInfo, statErr := os.Stat(".")
	require.NoError(t, statErr)
	targetInfo, statErr := os.Stat(target)
	require.NoError(t, statErr)
	require.True(t, os.SameFile(cwdInfo, targetInfo), "生成完成后必须恢复到新的目标目录")
	require.Contains(t, readFile(t, filepath.Join(target, "README.md")), "# current-project")
	require.FileExists(t, filepath.Join(target, "go.sum"))
	if runtime.GOOS != "windows" {
		require.Equal(t, os.FileMode(0o750), targetInfo.Mode().Perm())
	}
}

// TestCreateProjectFailureDoesNotLeavePartialTarget 验证生成失败会清理临时目录并保留目标目录的原始状态。
func TestCreateProjectFailureDoesNotLeavePartialTarget(t *testing.T) {
	for _, existingTarget := range []bool{false, true} {
		t.Run(fmt.Sprintf("existing_target_%t", existingTarget), func(t *testing.T) {
			parent := t.TempDir()
			target := filepath.Join(parent, "demo")
			if existingTarget {
				require.NoError(t, os.Mkdir(target, 0o755))
			}
			runner := func(_ string, name string, args ...string) ([]byte, error) {
				if name == "go" && strings.Join(args, " ") == "test ./... -count=1" {
					return []byte("project tests failed"), errors.New("exit status 1")
				}
				return nil, nil
			}

			err := createProjectWithRunner(target, ioDiscard{}, "v1.2.3", newOptions{
				projectType: "api",
				modulePath:  "example.com/demo",
			}, runner)

			require.Error(t, err)
			require.Contains(t, err.Error(), "验证生成项目失败")
			if existingTarget {
				require.DirExists(t, target)
				entries, readErr := os.ReadDir(target)
				require.NoError(t, readErr)
				require.Empty(t, entries)
			} else {
				require.NoDirExists(t, target)
			}
			temporaryDirs, globErr := filepath.Glob(filepath.Join(parent, ".initra-new-*"))
			require.NoError(t, globErr)
			require.Empty(t, temporaryDirs)
		})
	}
}

// TestExecuteProjectCommandDisablesWorkspace 验证生成项目的 Go 命令不受调用方 go.work 影响。
func TestExecuteProjectCommandDisablesWorkspace(t *testing.T) {
	output, err := executeProjectCommand(t.TempDir(), "go", "env", "GOWORK")

	require.NoError(t, err)
	require.Equal(t, "off", strings.TrimSpace(string(output)))
}

// TestNewGeneratesPublishedProjectWithoutReplace 验证已发布框架版本在无本地 replace 时能完成端到端生成。
func TestNewGeneratesPublishedProjectWithoutReplace(t *testing.T) {
	frameworkVersion := strings.TrimSpace(os.Getenv("INITRA_RELEASE_INTEGRATION_VERSION"))
	if frameworkVersion == "" {
		t.Skip("设置 INITRA_RELEASE_INTEGRATION_VERSION=vX.Y.Z 后执行需要网络的发布版集成测试")
	}
	t.Setenv("GOWORK", "off")
	target := filepath.Join(t.TempDir(), "demo")

	err := run([]string{
		"new", target,
		"--type", "api",
		"--module", "example.com/demo",
		"--framework-version", frameworkVersion,
	}, ioDiscard{}, frameworkVersion)
	require.NoError(t, err)

	goMod := readFile(t, filepath.Join(target, "go.mod"))
	require.NotContains(t, goMod, "replace github.com/teamsillybees/initra")
	require.FileExists(t, filepath.Join(target, "go.sum"))
	command := exec.Command("go", "test", "./...")
	command.Dir = target
	command.Env = append(os.Environ(), "GOWORK=off")
	output, testErr := command.CombinedOutput()
	require.NoError(t, testErr, string(output))
}

func TestNewGeneratesAPIMigrationsWithPhysicalForeignKeys(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")

	err := run([]string{
		"new", target,
		"--type", "api",
		"--module", "example.com/demo",
		"--replace", repoRoot(t),
	}, ioDiscard{}, "dev")

	require.NoError(t, err)
	migrations, err := filepath.Glob(filepath.Join(target, "db", "migrations", "*_add_relationship_foreign_keys.sql"))
	require.NoError(t, err)
	require.Len(t, migrations, 1)
	migrationContent := readFile(t, migrations[0])
	require.Contains(t, migrationContent, `CONSTRAINT "fk_sys_user_role_user" FOREIGN KEY ("user_id") REFERENCES "sys_user" ("id")`)
	require.Contains(t, migrationContent, `CONSTRAINT "fk_sys_role_menu_role" FOREIGN KEY ("role_id") REFERENCES "sys_role" ("id")`)

	require.NoFileExists(t, filepath.Join(target, "internal", "data", "ent", "migrate", "main.go"))
	diffGenerator := readFile(t, filepath.Join(target, "internal", "data", "migratediff", "main.go"))
	require.Contains(t, diffGenerator, `_ "github.com/lib/pq"`)
	require.Contains(t, diffGenerator, "boot.LoadConfig")
	require.Contains(t, diffGenerator, "data.SQLDBConfig")
	require.Contains(t, diffGenerator, "migrate.WithForeignKeys(true)")
	require.Contains(t, diffGenerator, "schema.WithMigrationMode(schema.ModeReplay)")
	require.Contains(t, readFile(t, filepath.Join(target, "db", "atlas.hcl")), "docker://postgres/16/dev")

	entSchema := readFile(t, filepath.Join(target, "internal", "data", "schema", "sys_user.go"))
	require.Contains(t, entSchema, "entsql.WithComments(true)")
	require.Contains(t, entSchema, `schema.Comment("系统后台用户表，用于后台登录、审计和权限归属。")`)

	migrateSchema := readFile(t, filepath.Join(target, "internal", "data", "ent", "migrate", "schema.go"))
	require.Contains(t, migrateSchema, `"系统后台用户表，用于后台登录、审计和权限归属。"`)
	require.Contains(t, migrateSchema, `Comment: "登录用户名，全局唯一。"`)

	require.NoDirExists(t, filepath.Join(target, "scripts"))

	smokeCommand := exec.Command(
		"go", "run", "./internal/data/migratediff/main.go",
		"smoke", "--unknown-option",
	)
	smokeCommand.Dir = target
	smokeCommand.Env = append(os.Environ(), "GOWORK=off")
	smokeOutput, smokeErr := smokeCommand.CombinedOutput()
	require.Error(t, smokeErr)
	require.Contains(t, string(smokeOutput), `unknown option "--unknown-option"`)
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
	require.FileExists(t, filepath.Join(skillDir, "references", "modules.md"))
	require.FileExists(t, filepath.Join(skillDir, "references", "redisx.md"))
	require.FileExists(t, filepath.Join(skillDir, "scripts", "check_initra_usage.go"))
	require.NoDirExists(t, filepath.Join(skillDir, "agents"))
	require.NoDirExists(t, filepath.Join(skillDir, "assets"))
	require.NoDirExists(t, filepath.Join(skillDir, "examples"))
	require.Contains(t, stdout.String(), "created skill")
}

func TestSkillDefaultCopiesCodexInitraFrameworkSkill(t *testing.T) {
	target := t.TempDir()
	t.Chdir(target)

	var stdout bytes.Buffer
	err := run([]string{"skill"}, &stdout, "dev")

	require.NoError(t, err)
	skillDir := filepath.Join(target, ".agents", "skills", "initra-framework")
	require.FileExists(t, filepath.Join(skillDir, "SKILL.md"))
	require.FileExists(t, filepath.Join(skillDir, "references", "modules.md"))
	require.FileExists(t, filepath.Join(skillDir, "references", "redisx.md"))
	require.FileExists(t, filepath.Join(skillDir, "scripts", "check_initra_usage.go"))
	require.NoDirExists(t, filepath.Join(skillDir, "agents"))
	require.NoDirExists(t, filepath.Join(skillDir, "assets"))
	require.NoDirExists(t, filepath.Join(skillDir, "examples"))
	require.Contains(t, stdout.String(), "created skill")
}

func TestSkillClaudeCodeTargetsRemoved(t *testing.T) {
	for _, target := range []string{"cc", "claude", "claude-code"} {
		t.Run(target, func(t *testing.T) {
			workspace := t.TempDir()
			t.Chdir(workspace)

			err := run([]string{"skill", target}, ioDiscard{}, "dev")

			require.Error(t, err)
			require.NoDirExists(t, filepath.Join(workspace, ".claude"))
		})
	}
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
	for _, suffix := range []string{"handler", "service", "dto", "routes"} {
		require.FileExists(t, filepath.Join(moduleDir, "order."+suffix+".go"))
	}
	require.FileExists(t, filepath.Join(moduleDir, "providers.go"))
	require.FileExists(t, filepath.Join(moduleDir, "order_test.go"))
	require.NoFileExists(t, filepath.Join(moduleDir, "order.model.go"))
	require.NoFileExists(t, filepath.Join(moduleDir, "order.repo.go"))
	handlerContent := readFile(t, filepath.Join(moduleDir, "order.handler.go"))
	require.NotContains(t, handlerContent, "type GetOrderInput")
	require.NotContains(t, handlerContent, "type OrderResponse")
	require.Contains(t, handlerContent, "response.OK(ctx, item)")
	require.NotContains(t, handlerContent, "requestctx")
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

// TestModuleAddGeneratedCodePassesProjectChecks 验证模块骨架通过格式、架构、编译和静态检查。
func TestModuleAddGeneratedCodePassesProjectChecks(t *testing.T) {
	root := repoRoot(t)
	target := t.TempDir()
	goMod := fmt.Sprintf(`module example.com/demo

go 1.26.0

require github.com/teamsillybees/initra v0.0.0

replace github.com/teamsillybees/initra => %s
`, filepath.ToSlash(root))
	require.NoError(t, os.WriteFile(filepath.Join(target, "go.mod"), []byte(goMod), 0o644))
	t.Chdir(target)
	require.NoError(t, run([]string{"module", "add", "order"}, ioDiscard{}, "dev"))

	moduleDir := filepath.Join(target, "internal", "modules", "order")
	goFiles, err := filepath.Glob(filepath.Join(moduleDir, "*.go"))
	require.NoError(t, err)
	formatCommand := exec.Command("gofmt", append([]string{"-l"}, goFiles...)...)
	formatOutput, err := formatCommand.CombinedOutput()
	require.NoError(t, err, string(formatOutput))
	require.Empty(t, strings.TrimSpace(string(formatOutput)), "生成的 Go 文件必须已经过 gofmt")

	checkerCommand := exec.Command("go", "run", "./tools/skills/initra-framework/scripts/check_initra_usage.go", "--root", target)
	checkerCommand.Dir = root
	checkerOutput, err := checkerCommand.CombinedOutput()
	require.NoError(t, err, string(checkerOutput))

	for _, args := range [][]string{
		{"test", "-mod=mod", "./internal/modules/order"},
		{"vet", "-mod=mod", "./internal/modules/order"},
	} {
		command := exec.Command("go", args...)
		command.Dir = target
		command.Env = append(os.Environ(), "GOWORK=off")
		output, commandErr := command.CombinedOutput()
		require.NoError(t, commandErr, "%s\n%s", strings.Join(args, " "), string(output))
	}
}

func TestSnippetAddGeneratesExplicitTableSnippet(t *testing.T) {
	target := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(target, "internal", "modules", "order"), 0o755))
	t.Chdir(target)

	err := run([]string{"snippet", "add", "order", "--table", "sys_order"}, ioDiscard{}, "dev")

	require.NoError(t, err)
	content := readFile(t, filepath.Join(target, "internal", "modules", "order", "order.table.go"))
	require.Contains(t, content, "sys_order")
	require.Contains(t, content, "const OrderTableName")
	require.NotContains(t, content, "type OrderCRUD")
}

func TestConfigAddConnectsAggregateAndBaseYAML(t *testing.T) {
	target := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(target, "internal", "boot"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(target, "configs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(target, "internal", "boot", "config.go"), []byte("package boot\n\ntype Config struct{}\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(target, "configs", "config.yaml"), []byte("app:\n  name: demo\n"), 0o644))
	t.Chdir(target)

	err := run([]string{"config", "add", "feature_flag"}, ioDiscard{}, "dev")

	require.NoError(t, err)
	require.FileExists(t, filepath.Join(target, "internal", "boot", "feature_flag.config.go"))
	require.Contains(t, readFile(t, filepath.Join(target, "internal", "boot", "config.go")), "FeatureFlag FeatureFlagConfig")
	require.Contains(t, readFile(t, filepath.Join(target, "internal", "boot", "config.go")), `mapstructure:"feature_flag"`)
	require.Contains(t, readFile(t, filepath.Join(target, "configs", "config.yaml")), "feature_flag:\n  enabled: false")
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
	t.Chdir(t.TempDir())

	err := run([]string{"doctor"}, &stdout, "dev")

	require.Error(t, err)
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
