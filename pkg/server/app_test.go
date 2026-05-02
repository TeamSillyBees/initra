package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
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
		Secret:          "server-test-secret",
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

// TestNewAppAcceptsValidJWTForProtectedAPIRoute 验证受保护 /api 路由会完成 JWT 校验与 Casbin 授权。
func TestNewAppAcceptsValidJWTForProtectedAPIRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager, err := platformauth.NewJWTManager(platformauth.JWTConfig{
		Issuer:          "initra",
		Secret:          "server-test-secret",
		AccessTokenTTL:  time.Minute,
		RefreshTokenTTL: time.Hour,
	})
	require.NoError(t, err)

	modelPath, policyPath := writeWebCasbinFiles(t)
	enforcer, err := platformauth.NewEnforcer(modelPath, policyPath)
	require.NoError(t, err)

	app, err := NewApp(Options{Title: "initra", Version: "test", Env: "test"}, nil, manager, enforcer)
	require.NoError(t, err)

	app.Registry.Register(http.MethodGet, "/api/v1/protected", platformauth.RouteSecurity{
		Resource: "user",
		Action:   "read",
	})
	app.Engine.GET("/api/v1/protected", func(c *gin.Context) {
		principal, ok := platformauth.PrincipalFromContext(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, int64(1001), principal.UserID)
		c.Status(http.StatusNoContent)
	})

	pair, err := manager.IssueTokenPair(t.Context(), platformauth.Principal{
		UserID: 1001,
		Roles:  []string{"admin"},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()

	app.Engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

// TestNewAppAllowsPublicAPIRouteWithoutToken 验证显式 Public 的 /api 路由不会触发 JWT 校验。
func TestNewAppAllowsPublicAPIRouteWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app, err := NewApp(Options{Title: "initra", Version: "test", Env: "test"}, nil, nil, nil)
	require.NoError(t, err)

	app.Registry.Register(http.MethodGet, "/api/v1/public", platformauth.RouteSecurity{Public: true})
	app.Engine.GET("/api/v1/public", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public", nil)
	rec := httptest.NewRecorder()

	app.Engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

// TestNewAppRejectsAPIRouteMissingSecurityMetadata 验证 /api 路由缺少安全元信息时默认拒绝。
func TestNewAppRejectsAPIRouteMissingSecurityMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app, err := NewApp(Options{Title: "initra", Version: "test", Env: "test"}, nil, nil, nil)
	require.NoError(t, err)

	app.Engine.GET("/api/v1/unregistered", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/unregistered", nil)
	rec := httptest.NewRecorder()

	app.Engine.ServeHTTP(rec, req)

	var body struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Equal(t, "FORBIDDEN", body.Code)
	require.Equal(t, "route security metadata is missing", body.Message)
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

// writeWebCasbinFiles 为 Web 链路测试写入最小 Casbin 模型和策略文件。
func writeWebCasbinFiles(t *testing.T) (string, string) {
	t.Helper()

	dir := t.TempDir()
	modelPath := filepath.Join(dir, "rbac_model.conf")
	policyPath := filepath.Join(dir, "rbac_policy.csv")

	model := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`
	policy := `
p, admin, user, read
`

	require.NoError(t, os.WriteFile(modelPath, []byte(strings.TrimSpace(model)), 0o600))
	require.NoError(t, os.WriteFile(policyPath, []byte(strings.TrimSpace(policy)), 0o600))
	return modelPath, policyPath
}
