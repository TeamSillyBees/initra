package logging

import (
	"testing"

	"github.com/samber/do"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRegisterProvidesLogger(t *testing.T) {
	injector := do.New()
	Register(injector, Config{Level: "info", Format: "json", Output: "stdout"})

	logger := do.MustInvoke[*zap.Logger](injector)

	require.NotNil(t, logger)
}
