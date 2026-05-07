package httpclient

import (
	"testing"

	"github.com/samber/do"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRegisterProvidesFactory(t *testing.T) {
	injector := do.New()
	Register(injector, Config{Enabled: false}, zap.NewNop())

	factory := do.MustInvoke[Factory](injector)

	require.NotNil(t, factory)
}
