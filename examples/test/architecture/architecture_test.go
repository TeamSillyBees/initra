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

	require.DirExists(t, filepath.Join(root, "internal", "modules"))
	for _, moduleName := range []string{"auth", "user"} {
		moduleDir := filepath.Join(root, "internal", "modules", moduleName)
		require.DirExists(t, moduleDir)
		for _, suffix := range []string{"handler", "service", "repo", "model", "dto", "routes"} {
			require.FileExists(t, filepath.Join(moduleDir, moduleName+"."+suffix+".go"))
		}
		require.FileExists(t, filepath.Join(moduleDir, "providers.go"))
		requireNoTransportTypes(t, filepath.Join(moduleDir, moduleName+".handler.go"))
		requireHTTPNaming(t, filepath.Join(moduleDir, moduleName+".dto.go"))
		requireNoHTTPNamingInModel(t, filepath.Join(moduleDir, moduleName+".model.go"))
	}
	require.FileExists(t, filepath.Join(root, "db", "schema", "01_sys_user.sql"))
	require.FileExists(t, filepath.Join(root, "db", "seeds", "001_seed_admin.sql"))
	require.NoDirExists(t, filepath.Join(root, "internal", "gen", "jet"))
	require.NoDirExists(t, filepath.Join(root, "tools", "jetgen"))
	require.NoDirExists(t, filepath.Join(root, "scripts"))
	require.DirExists(t, filepath.Join(root, "internal", "data", "schema"))
	require.FileExists(t, filepath.Join(root, "internal", "data", "ent", "client.go"))
	require.FileExists(t, filepath.Join(root, "internal", "data", "tx.go"))
	requireNoEntImportOutsideRepositories(t, root)
	_, err := os.Stat(filepath.Join(root, "internal", "app"))
	require.True(t, errors.Is(err, os.ErrNotExist), "示例项目不应继续保留 internal/app 业务目录")
}

// requireNoTransportTypes 确认 handler 文件只承载 HTTP 适配逻辑，不再混放类型定义。
func requireNoTransportTypes(t *testing.T, path string) {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	transportType := regexp.MustCompile(`(?m)^type\s+\w+(Input|Output|Request|Response|Query|Body|VO|DTO|Params)\s+`)
	exportedHandlerMethod := regexp.MustCompile(`(?m)^func\s+\(h\s+\*Handler\)\s+[A-Z]\w*\s*\(`)
	require.Falsef(t, transportType.Match(content), "%s should keep type definitions outside handler files", path)
	require.Falsef(t, exportedHandlerMethod.Match(content), "%s should keep Huma handler methods unexported", path)
}

// requireHTTPNaming 固定 HTTP 边界类型命名：Huma 包装类型仅内部使用，JSON 出参使用 VO。
func requireHTTPNaming(t *testing.T, path string) {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	legacyType := regexp.MustCompile(`(?m)^type\s+\w+(Input|Output|Params)\s+`)
	exportedWrapper := regexp.MustCompile(`(?m)^type\s+[A-Z]\w*(Request|Response)\s+`)
	require.Falsef(t, legacyType.Match(content), "%s should not use Input/Output/Params suffixes", path)
	require.Falsef(t, exportedWrapper.Match(content), "%s should keep Huma Request/Response wrappers unexported", path)
}

// requireNoHTTPNamingInModel 确认领域模型文件不承载 HTTP 边界类型。
func requireNoHTTPNamingInModel(t *testing.T, path string) {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	httpType := regexp.MustCompile(`(?m)^type\s+\w+(Input|Output|Request|Response|Query|Body|VO|Params)\s+`)
	require.Falsef(t, httpType.Match(content), "%s should keep HTTP boundary types in *.dto.go", path)
}

// requireNoEntImportOutsideRepositories 确认 Ent 只出现在持久化边界，不泄漏到 handler/service/dto。
func requireNoEntImportOutsideRepositories(t *testing.T, root string) {
	t.Helper()

	moduleRoot := filepath.Join(root, "internal", "modules")
	require.NoError(t, filepath.WalkDir(moduleRoot, func(path string, entry os.DirEntry, err error) error {
		require.NoError(t, err)
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".repo.go") || filepath.Base(path) == "providers.go" {
			return nil
		}
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		require.NotContainsf(t, string(content), "/internal/data/ent", "%s should not import internal/data/ent", path)
		return nil
	}))
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
