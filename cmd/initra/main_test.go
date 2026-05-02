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

func TestNewGeneratesProjectWithFrameworkRequireAndNoPkgCopy(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")

	var stdout bytes.Buffer
	err := run([]string{
		"new", target,
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
	require.Contains(t, stdout.String(), "created")
}

func TestNewWritesLocalReplaceWhenRequested(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")
	frameworkPath := repoRoot(t)

	err := run([]string{
		"new", target,
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

	err := run([]string{"new", target, "--module", "example.com/demo"}, ioDiscard{}, "dev")

	require.Error(t, err)
	require.Contains(t, err.Error(), "--framework-version")
}

func TestGeneratedProjectBuildsWithLocalReplace(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")

	err := run([]string{
		"new", target,
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
