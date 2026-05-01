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
