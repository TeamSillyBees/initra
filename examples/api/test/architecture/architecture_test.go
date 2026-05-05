package architecture_test

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// goPackage 承接 go list 输出的包路径和直接依赖列表。
type goPackage struct {
	ImportPath string
	Imports    []string
}

// TestExampleDoesNotImportFrameworkInternalPackages 验证示例项目只依赖框架公开 pkg API。
func TestExampleDoesNotImportFrameworkInternalPackages(t *testing.T) {
	packages := goList(t, "./...")

	for _, pkg := range packages {
		for _, imported := range pkg.Imports {
			require.Falsef(
				t,
				strings.HasPrefix(imported, "github.com/teamsillybees/initra/internal/"),
				"%s imports framework internal package %s; examples must use pkg APIs",
				pkg.ImportPath,
				imported,
			)
		}
	}
}

// TestExampleUsesAPIFoundationLayout 固定 API 模板包含基础业务和数据文件。
func TestExampleUsesAPIFoundationLayout(t *testing.T) {
	root := repoRoot(t)

	require.DirExists(t, filepath.Join(root, "internal", "module"))
	for _, moduleName := range []string{"auth", "user"} {
		moduleDir := filepath.Join(root, "internal", "module", moduleName)
		require.DirExists(t, moduleDir)
		for _, suffix := range []string{"handler", "service", "repo", "model", "dto", "routes"} {
			require.FileExists(t, filepath.Join(moduleDir, moduleName+"."+suffix+".go"))
		}
		require.FileExists(t, filepath.Join(moduleDir, "providers.go"))
		requireNoTransportTypes(t, filepath.Join(moduleDir, moduleName+".handler.go"))
	}
	require.FileExists(t, filepath.Join(root, "db", "schema", "01_sys_user.sql"))
	require.FileExists(t, filepath.Join(root, "db", "seeds", "001_seed_admin.sql"))
	require.FileExists(t, filepath.Join(root, "internal", "gen", "jet", "table", "sys_user.go"))
	require.FileExists(t, filepath.Join(root, "tools", "jetgen", "main.go"))
	_, err := os.Stat(filepath.Join(root, "internal", "app"))
	require.True(t, errors.Is(err, os.ErrNotExist), "示例项目不应继续保留 internal/app 业务目录")
}

// requireNoTransportTypes 确认 handler 文件只承载 HTTP 适配逻辑，不再混放 DTO 类型。
func requireNoTransportTypes(t *testing.T, path string) {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	transportType := regexp.MustCompile(`(?m)^type\s+\w+(Input|Request|Response|Output)\s+`)
	require.Falsef(t, transportType.Match(content), "%s should keep transport DTOs in *.dto.go", path)
}

// goList 调用 go list 并解析指定包模式的依赖元信息。
func goList(t *testing.T, pattern string) []goPackage {
	t.Helper()

	cmd := exec.Command("go", "list", "-json", pattern)
	cmd.Dir = repoRoot(t)
	output, err := cmd.Output()
	require.NoError(t, err)

	decoder := json.NewDecoder(strings.NewReader(string(output)))
	packages := make([]goPackage, 0)
	// go list -json 会连续输出多个 JSON 对象，而不是 JSON 数组；
	// 这里按流式对象逐个解码，避免目录数量变化时测试不稳定。
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

// repoRoot 从当前测试文件位置推导仓库根目录。
func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
