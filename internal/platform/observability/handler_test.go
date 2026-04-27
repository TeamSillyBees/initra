package observability_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/stretchr/testify/require"
	"github.com/teamsillybees/initra/internal/platform/auth"
	"github.com/teamsillybees/initra/internal/platform/observability"
	"github.com/teamsillybees/initra/internal/platform/web"
	"go.uber.org/zap"
)

// TestModuleRegistersHealthReadyAndVersionEndpoints 验证观测模块注册健康、就绪和版本接口。
func TestModuleRegistersHealthReadyAndVersionEndpoints(t *testing.T) {
	logger := zap.NewNop()
	jwtManager, err := auth.NewJWTManager(auth.JWTConfig{
		Issuer:          "initra",
		Secret:          "observability-test-secret",
		AccessTokenTTL:  3600,
		RefreshTokenTTL: 7200,
	})
	require.NoError(t, err)

	enforcer, err := casbin.NewEnforcer()
	require.NoError(t, err)

	app, err := web.NewApp(web.Options{
		Title:   "initra",
		Version: "test",
		Env:     "test",
	}, logger, jwtManager, enforcer)
	require.NoError(t, err)

	module := observability.NewModule(observability.BuildInfo{
		Version:   "test",
		Commit:    "abc123",
		BuildTime: "2026-04-21T00:00:00Z",
	})
	module.Register(app.API, app.Registry)

	for _, path := range []string{"/health", "/ready", "/version"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		app.Engine.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code, path)
	}
}
