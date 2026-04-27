package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// staticRouteSecurityLookup 始终返回同一份安全元信息，用于 JWT 中间件单元测试。
type staticRouteSecurityLookup struct {
	security RouteSecurity
}

// Lookup 返回预置安全元信息。
func (s staticRouteSecurityLookup) Lookup(method string, path string) (RouteSecurity, bool) {
	return s.security, true
}

// missingRouteSecurityLookup 模拟路由未登记安全元信息的危险场景。
type missingRouteSecurityLookup struct{}

// Lookup 始终返回未命中。
func (missingRouteSecurityLookup) Lookup(method string, path string) (RouteSecurity, bool) {
	return RouteSecurity{}, false
}

// TestJWTMiddlewareRejectsBlacklistedToken 验证 JWT 中间件拒绝已被黑名单吊销的 token。
func TestJWTMiddlewareRejectsBlacklistedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := &fakeTokenStore{refreshValid: true}
	manager, err := NewJWTManager(JWTConfig{
		Issuer:          "initra",
		Secret:          "middleware-test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Store:           store,
	})
	require.NoError(t, err)

	pair, err := manager.IssueTokenPair(t.Context(), Principal{UserID: 1001, Roles: []string{"admin"}})
	require.NoError(t, err)

	claims, err := manager.parse(pair.AccessToken, TokenTypeAccess)
	require.NoError(t, err)
	store.blacklistedTokenID = claims.ID
	store.accessBlacklisted = true

	engine := gin.New()
	engine.Use(JWTMiddleware(manager, staticRouteSecurityLookup{
		security: RouteSecurity{Resource: "auth", Action: "read"},
	}, nil))
	engine.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestAuthorizationMiddlewareRejectsUnregisteredAPIRoute 验证 /api 路由缺少安全元信息时默认拒绝。
func TestAuthorizationMiddlewareRejectsUnregisteredAPIRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		ctx := WithPrincipal(c.Request.Context(), Principal{
			UserID: 1001,
			Roles:  []string{"admin"},
		})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	engine.Use(AuthorizationMiddleware(nil, missingRouteSecurityLookup{}, nil))
	engine.GET("/api/v1/unregistered", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/unregistered", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
}
