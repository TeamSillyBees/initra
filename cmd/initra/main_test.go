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
	}, &stdout, "dev")

	require.NoError(t, err)
	goMod := readFile(t, filepath.Join(target, "go.mod"))
	require.Contains(t, goMod, "module example.com/demo")
	require.Contains(t, goMod, "github.com/teamsillybees/initra v1.2.3")
	require.NoDirExists(t, filepath.Join(target, "pkg"))
	require.Contains(t, readFile(t, filepath.Join(target, "cmd", "server", "main.go")), "example.com/demo/internal/boot")
	require.FileExists(t, filepath.Join(target, "internal", "boot", "app.go"))
	require.FileExists(t, filepath.Join(target, "internal", "boot", "providers.go"))
	require.FileExists(t, filepath.Join(target, "internal", "boot", "lifecycle.go"))
	require.DirExists(t, filepath.Join(target, "internal", "module"))
	require.NoDirExists(t, filepath.Join(target, "internal", "module", "auth"))
	require.NoDirExists(t, filepath.Join(target, "internal", "module", "user"))
	require.NoDirExists(t, filepath.Join(target, "internal", "app"))
	require.Contains(t, stdout.String(), "created")
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
	for _, suffix := range []string{"handler", "service", "repo", "model", "routes"} {
		require.FileExists(t, filepath.Join(moduleDir, "order."+suffix+".go"))
	}
	require.FileExists(t, filepath.Join(moduleDir, "providers.go"))
	require.FileExists(t, filepath.Join(moduleDir, "order_test.go"))
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
	require.Contains(t, diffScript, "atlas migrate diff add_order")
}

func TestDoctorReportsEnvironmentChecks(t *testing.T) {
	var stdout bytes.Buffer

	err := run([]string{"doctor"}, &stdout, "dev")

	require.NoError(t, err)
	require.Contains(t, stdout.String(), "Go:")
	require.Contains(t, stdout.String(), "Atlas:")
	require.Contains(t, stdout.String(), "go-jet:")
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
