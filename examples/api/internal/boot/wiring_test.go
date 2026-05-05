package boot

import (
	"testing"

	"github.com/samber/do"
	"github.com/stretchr/testify/require"
	authmodule "github.com/teamsillybees/initra/examples/api/internal/module/auth"
	usermodule "github.com/teamsillybees/initra/examples/api/internal/module/user"
)

// TestModuleProvideDoesNotCollide 验证 auth 和 user 模块使用具名依赖后不会在 DI 容器中冲突。
func TestModuleProvideDoesNotCollide(t *testing.T) {
	injector := do.New()

	require.NotPanics(t, func() {
		usermodule.Provide(injector)
		authmodule.Provide(injector)
	})
}
