package redisx

import (
	"testing"

	"github.com/samber/do"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRegisterProvidesDisabledRedisClient(t *testing.T) {
	injector := do.New()
	do.ProvideValue(injector, zap.NewNop())
	Register(injector, Config{Enabled: false})

	client := do.MustInvoke[UniversalClient](injector)

	require.Nil(t, client)
}
