package architecture_test

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
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
			require.Falsef(
				t,
				strings.Contains(imported, "/internal/data/ent"),
				"%s imports generated Ent package %s",
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

// TestRedisConstructionGoesThroughRedisx 固定 Redis client 和 Lua 脚本必须通过 pkg/redisx 统一封装。
func TestRedisConstructionGoesThroughRedisx(t *testing.T) {
	disallowed := []string{
		"redis.NewClient(",
		"redis.NewFailoverClient(",
		"redis.NewUniversalClient(",
		"redis.NewScript(",
	}
	allowedDirs := []string{
		filepath.Join("pkg", "redisx"),
	}

	root := repoRoot(t)
	for _, dir := range []string{"pkg", filepath.Join("examples", "api")} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, entry fs.DirEntry, err error) error {
			require.NoError(t, err)
			if entry.IsDir() {
				rel, relErr := filepath.Rel(root, path)
				require.NoError(t, relErr)
				if isPathUnderAny(rel, allowedDirs) {
					return filepath.SkipDir
				}
				return nil
			}
			if filepath.Ext(path) != ".go" {
				return nil
			}

			content, readErr := os.ReadFile(path)
			require.NoError(t, readErr)
			for _, token := range disallowed {
				require.NotContainsf(t, string(content), token, "%s must use pkg/redisx instead of %s", path, token)
			}
			return nil
		})
		require.NoError(t, err)
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

func isPathUnderAny(path string, prefixes []string) bool {
	clean := filepath.Clean(path)
	for _, prefix := range prefixes {
		prefix = filepath.Clean(prefix)
		if clean == prefix || strings.HasPrefix(clean, prefix+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
