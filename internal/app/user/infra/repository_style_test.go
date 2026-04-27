package infra

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRepositorySourceUsesJetDotImports 验证 user 仓储保持 go-jet dot import 风格且不回退到硬编码 SQL。
func TestRepositorySourceUsesJetDotImports(t *testing.T) {
	content := readLocalGoFile(t, "repository.go")

	require.Contains(t, content, `. "github.com/go-jet/jet/v2/postgres"`)
	require.Contains(t, content, `. "github.com/teamsillybees/initra/internal/gen/jet/table"`)
	require.NotContains(t, content, "postgres.")
	require.NotContains(t, content, "table.")
	require.NotContains(t, content, "DELETE FROM public.sys_user_role")
	require.NotContains(t, content, "INSERT INTO public.sys_user_role")
	require.NotContains(t, content, "SELECT sr.")
}

// readLocalGoFile 读取当前测试文件同目录下的 Go 源码，并统一换行符便于字符串断言。
func readLocalGoFile(t *testing.T, name string) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)

	content, err := os.ReadFile(filepath.Join(filepath.Dir(file), name))
	require.NoError(t, err)

	return strings.ReplaceAll(string(content), "\r\n", "\n")
}
