package server

import (
	"testing"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/samber/do"
	"github.com/stretchr/testify/require"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
	"go.uber.org/zap"
)

func TestRegisterProvidesApp(t *testing.T) {
	injector := do.New()
	manager, err := platformauth.NewJWTManager(platformauth.JWTConfig{
		Issuer:          "initra",
		Secret:          "server-register-test-secret",
		AccessTokenTTL:  time.Minute,
		RefreshTokenTTL: time.Hour,
	})
	require.NoError(t, err)
	enforcer, err := casbin.NewEnforcer()
	require.NoError(t, err)
	do.ProvideValue(injector, zap.NewNop())
	do.ProvideValue(injector, manager)
	do.ProvideValue(injector, enforcer)
	Register(injector, Options{Title: "initra", Version: "test", Env: "test"})

	app := do.MustInvoke[*App](injector)

	require.NotNil(t, app.Engine)
	require.NotNil(t, app.API)
	require.NotNil(t, app.Registry)
}
