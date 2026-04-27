package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	platformauth "github.com/teamsillybees/initra/internal/platform/auth"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// TestNewAppAllowsCORSPreflightBeforeAuth 验证 CORS 预检在 JWT 校验前被正确放行。
func TestNewAppAllowsCORSPreflightBeforeAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app, err := NewApp(Options{Title: "initra", Version: "test", Env: "test"}, nil, nil, nil)
	require.NoError(t, err)

	app.Registry.Register(http.MethodGet, "/api/v1/protected", platformauth.RouteSecurity{
		Resource: "user",
		Action:   "read",
	})
	app.Engine.GET("/api/v1/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/protected", nil)
	rec := httptest.NewRecorder()

	app.Engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
}

// TestNewAppLogsUnauthorizedRequests 验证认证失败请求也会经过请求日志中间件。
func TestNewAppLogsUnauthorizedRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	manager, err := platformauth.NewJWTManager(platformauth.JWTConfig{
		Issuer:          "initra",
		Secret:          "web-test-secret",
		AccessTokenTTL:  time.Minute,
		RefreshTokenTTL: time.Hour,
	})
	require.NoError(t, err)

	app, err := NewApp(Options{Title: "initra", Version: "test", Env: "test"}, logger, manager, nil)
	require.NoError(t, err)

	app.Registry.Register(http.MethodGet, "/api/v1/protected", platformauth.RouteSecurity{
		Resource: "user",
		Action:   "read",
	})
	app.Engine.GET("/api/v1/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	rec := httptest.NewRecorder()

	app.Engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.NotEmpty(t, logs.FilterMessage("http request completed").All())
}

// TestRouteRegistryMatchesGinColonParams 验证 Gin 参数路径能匹配 Huma/OpenAPI 参数路径。
func TestRouteRegistryMatchesGinColonParams(t *testing.T) {
	registry := NewRouteRegistry()
	expected := platformauth.RouteSecurity{Resource: "user", Action: "read"}

	registry.Register(http.MethodGet, "/api/v1/users/{id}", expected)

	actual, ok := registry.Lookup(http.MethodGet, "/api/v1/users/:id")
	require.True(t, ok)
	require.Equal(t, expected, actual)
}
