package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperrors "github.com/teamsillybees/initra/pkg/errors"
	"github.com/teamsillybees/initra/pkg/requestctx"
	"go.uber.org/zap"
)

// HTTP 头常量集中定义，避免不同中间件拼写不一致。
const (
	headerAuthorization = "Authorization"
	headerRequestID     = "X-Request-ID"
	headerTraceID       = "X-Trace-ID"
	routeSecurityKey    = "initra.route_security"
)

// routeSecurityResult 缓存单次请求的路由安全元信息，避免认证和授权中间件重复查表。
type routeSecurityResult struct {
	security RouteSecurity
	found    bool
}

// RequestContextMiddleware 为每个请求补齐 request_id 与 trace_id。
// 该中间件必须排在日志、CORS、认证与授权之前，确保即使请求被提前拒绝，也能留下完整链路信息。
func RequestContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := firstNonEmpty(c.GetHeader(headerRequestID), uuid.NewString())
		traceID := firstNonEmpty(c.GetHeader(headerTraceID), requestID)
		ctx := requestctx.WithRequestID(c.Request.Context(), requestID)
		ctx = requestctx.WithTraceID(ctx, traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Header(headerRequestID, requestID)
		c.Header(headerTraceID, traceID)

		c.Next()
	}
}

// JWTMiddleware 在进入业务处理前完成 JWT 身份解析。
func JWTMiddleware(manager *JWTManager, lookup RouteSecurityLookup, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions || shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		security, ok := resolveRouteSecurity(c, lookup)
		if !ok && isAPIRoute(c.Request.URL.Path) {
			writeError(c, apperrors.New(apperrors.CodeForbidden, "route security metadata is missing"))
			return
		}
		if ok && security.Public {
			c.Next()
			return
		}

		token, ok := bearerToken(c.GetHeader(headerAuthorization))
		if !ok {
			writeError(c, apperrors.New(apperrors.CodeUnauthorized, "authorization token is required"))
			return
		}

		claims, err := manager.ParseAccessToken(c.Request.Context(), token)
		if err != nil {
			if errors.Is(err, ErrTokenStoreFailure) {
				writeError(c, apperrors.Wrap(err, apperrors.CodeInternalError, "validate authorization token failed"))
				return
			}
			writeError(c, apperrors.New(apperrors.CodeUnauthorized, "authorization token is invalid"))
			return
		}

		c.Request = c.Request.WithContext(WithPrincipal(c.Request.Context(), Principal{
			UserID:   claims.UserID,
			Roles:    claims.Roles,
			TenantID: claims.TenantID,
		}))

		c.Next()

		if logger != nil && c.Writer.Status() >= http.StatusInternalServerError {
			logger.Warn("request completed with server error",
				zap.String("trace_id", requestctx.TraceIDFromContext(c.Request.Context())),
				zap.Int("status", c.Writer.Status()),
				zap.String("path", c.Request.URL.Path),
			)
		}
	}
}

// AuthorizationMiddleware 在身份校验后，基于 Casbin 策略继续完成授权判断。
func AuthorizationMiddleware(enforcer *casbin.Enforcer, lookup RouteSecurityLookup, _ *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions || shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		security, ok := resolveRouteSecurity(c, lookup)
		if !ok {
			if isAPIRoute(c.Request.URL.Path) {
				writeError(c, apperrors.New(apperrors.CodeForbidden, "route security metadata is missing"))
				return
			}
			c.Next()
			return
		}
		if security.Public {
			c.Next()
			return
		}
		if security.Resource == "" || security.Action == "" {
			writeError(c, apperrors.New(apperrors.CodeForbidden, "route security metadata is incomplete"))
			return
		}

		principal, ok := PrincipalFromContext(c.Request.Context())
		if !ok {
			writeError(c, apperrors.New(apperrors.CodeUnauthorized, "user principal is missing"))
			return
		}

		for _, role := range principal.Roles {
			allowed, err := enforcer.Enforce(role, security.Resource, security.Action)
			if err != nil {
				writeError(c, apperrors.Wrap(err, apperrors.CodeInternalError, "authorize request failed"))
				return
			}
			if allowed {
				c.Next()
				return
			}
		}

		writeError(c, apperrors.New(apperrors.CodeForbidden, "permission denied"))
	}
}

// RequestLoggerMiddleware 输出结构化请求日志。
func RequestLoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if logger == nil {
			return
		}

		traceID := requestctx.TraceIDFromContext(c.Request.Context())
		requestID := requestctx.RequestIDFromContext(c.Request.Context())
		fields := []zap.Field{
			zap.String("trace_id", traceID),
			zap.String("request_id", requestID),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("latency_ms", time.Since(start).Milliseconds()),
		}

		if principal, ok := PrincipalFromContext(c.Request.Context()); ok {
			fields = append(fields, zap.Int64("user_id", principal.UserID))
		}

		logger.Info("http request completed", fields...)
	}
}

// RecoveryMiddleware 捕获 panic 并统一转换成标准错误响应。
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		if logger != nil {
			logger.Error("panic recovered",
				zap.Any("panic", recovered),
				zap.String("path", c.Request.URL.Path),
			)
		}
		writeError(c, apperrors.New(apperrors.CodeInternalError, "internal error"))
	})
}

// CORSMiddleware 提供最小可用的跨域响应头。
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID, X-Trace-ID")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// resolveRouteSecurity 获取路由安全元信息，并在同一请求内复用查询结果。
func resolveRouteSecurity(c *gin.Context, lookup RouteSecurityLookup) (RouteSecurity, bool) {
	if cached, exists := c.Get(routeSecurityKey); exists {
		if result, ok := cached.(routeSecurityResult); ok {
			return result.security, result.found
		}
	}
	if lookup == nil {
		return RouteSecurity{}, false
	}

	routePath := c.FullPath()
	if routePath == "" {
		routePath = c.Request.URL.Path
	}
	security, ok := lookup.Lookup(c.Request.Method, routePath)
	c.Set(routeSecurityKey, routeSecurityResult{security: security, found: ok})
	return security, ok
}

// shouldSkipAuth 判断无需认证的系统级公开路径。
// 这些路径必须保持较小集合，业务公开接口应通过 RouteSecurity.Public 显式声明。
func shouldSkipAuth(path string) bool {
	return path == "/health" ||
		path == "/ready" ||
		path == "/version" ||
		strings.HasPrefix(path, "/docs") ||
		strings.HasPrefix(path, "/openapi") ||
		strings.HasPrefix(path, "/schemas")
}

// isAPIRoute 判断请求是否属于业务 API 命名空间，供 fail-closed 策略统一使用。
func isAPIRoute(path string) bool {
	return strings.HasPrefix(path, "/api/")
}

// bearerToken 从 Authorization 头中提取 Bearer token，并拒绝空 token 或错误认证方案。
func bearerToken(header string) (string, bool) {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}

// writeError 将平台错误统一转换为带 trace_id 的 JSON 响应。
func writeError(c *gin.Context, err error) {
	traceID := requestctx.TraceIDFromContext(c.Request.Context())
	status, body := apperrors.ToHTTP(err, traceID)
	c.AbortWithStatusJSON(status, body)
}

// firstNonEmpty 返回第一个非空字符串，用于请求 ID 和 trace ID 的兜底选择。
func firstNonEmpty(candidates ...string) string {
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}
	return ""
}
