package logx

import (
	"testing"

	"github.com/samber/do"
	"github.com/stretchr/testify/require"
)

// TestRegisterProvidesLogger 验证 DI 提供 logx Logger。
func TestRegisterProvidesLogger(t *testing.T) {
	injector := do.New()
	Register(injector, Config{Level: "debug"})

	require.NotNil(t, do.MustInvoke[*Logger](injector))
}
