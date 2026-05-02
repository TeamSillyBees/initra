package architecture_test

import (
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type goPackage struct {
	ImportPath string
	Imports    []string
}

// TestPublicPackagesDoNotImportExamples 固定框架公共包不能反向依赖示例项目。
func TestPublicPackagesDoNotImportExamples(t *testing.T) {
	for _, pkg := range goList(t, "./pkg/...") {
		for _, imported := range pkg.Imports {
			require.Falsef(
				t,
				strings.HasPrefix(imported, "github.com/teamsillybees/initra/examples/"),
				"%s imports example package %s",
				pkg.ImportPath,
				imported,
			)
		}
	}
}

// TestLowLevelPackagesDoNotImportWeb 固定底层公共包不能依赖 Web 装配层。
func TestLowLevelPackagesDoNotImportWeb(t *testing.T) {
	for _, pattern := range []string{"./pkg/errors", "./pkg/requestctx"} {
		for _, pkg := range goList(t, pattern) {
			for _, imported := range pkg.Imports {
				require.NotEqualf(t, "github.com/teamsillybees/initra/pkg/server", imported, "%s imports pkg/server", pkg.ImportPath)
			}
		}
	}
}

func goList(t *testing.T, pattern string) []goPackage {
	t.Helper()

	cmd := exec.Command("go", "list", "-json", pattern)
	cmd.Dir = repoRoot(t)
	output, err := cmd.Output()
	require.NoError(t, err)

	decoder := json.NewDecoder(strings.NewReader(string(output)))
	packages := make([]goPackage, 0)
	for {
		var pkg goPackage
		err := decoder.Decode(&pkg)
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		packages = append(packages, pkg)
	}
	return packages
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
