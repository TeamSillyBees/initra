package web

import (
	"net/http"
	"strings"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	platformauth "github.com/teamsillybees/initra/internal/platform/auth"
	apperrors "github.com/teamsillybees/initra/internal/platform/errors"
	sharedtypes "github.com/teamsillybees/initra/internal/shared/types"
	"go.uber.org/zap"
)

// Options 描述 Web 层初始化所需的最小配置。
type Options struct {
	Title   string
	Version string
	Env     string
}

// App 聚合 Gin Engine、Huma API 和路由安全注册表。
type App struct {
	Engine   *gin.Engine
	API      huma.API
	Registry *RouteRegistry
}

// RouteRegistry 负责记录每个路由对应的安全元信息。
type RouteRegistry struct {
	mu     sync.RWMutex
	routes map[string]platformauth.RouteSecurity
}

// NewRouteRegistry 创建一个空的路由注册表。
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		routes: map[string]platformauth.RouteSecurity{},
	}
}

// Register 注册路由的安全元信息。
func (r *RouteRegistry) Register(method string, path string, security platformauth.RouteSecurity) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes[routeKey(method, path)] = security
}

// Lookup 查询某个路由是否配置了安全元信息。
func (r *RouteRegistry) Lookup(method string, path string) (platformauth.RouteSecurity, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	security, ok := r.routes[routeKey(method, path)]
	return security, ok
}

// NewApp 创建集成 Gin、Huma、JWT 与 Casbin 的 Web 应用。
func NewApp(options Options, logger *zap.Logger, jwtManager *platformauth.JWTManager, enforcer *casbin.Enforcer) (*App, error) {
	configureGinMode(options.Env)
	configureHumaErrors()

	engine := gin.New()
	registry := NewRouteRegistry()
	engine.Use(
		platformauth.RecoveryMiddleware(logger),
		platformauth.RequestContextMiddleware(),
		platformauth.RequestLoggerMiddleware(logger),
		platformauth.CORSMiddleware(),
		platformauth.JWTMiddleware(jwtManager, registry, logger),
		platformauth.AuthorizationMiddleware(enforcer, registry, logger),
	)

	humaConfig := huma.DefaultConfig(options.Title, options.Version)
	humaConfig.OpenAPIPath = "/openapi"
	humaConfig.DocsPath = "/docs"
	humaConfig.SchemasPath = "/schemas"
	humaConfig.CreateHooks = nil

	api := humagin.New(engine, humaConfig)
	return &App{
		Engine:   engine,
		API:      api,
		Registry: registry,
	}, nil
}

// configureGinMode 根据运行环境设置 Gin 模式，避免生产环境输出调试日志。
func configureGinMode(env string) {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "prod", "production":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}

// routeKey 将 HTTP 方法和路径规范化后组成注册表 key。
func routeKey(method string, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + normalizeRoutePath(path)
}

// normalizeRoutePath 将 Gin 的 :id / *path 参数格式转换为 Huma/OpenAPI 的 {id} 格式。
// 这样注册表可以只维护一种路由写法，同时兼容 Gin 运行时 FullPath 返回值。
func normalizeRoutePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	segments := strings.Split(path, "/")
	for index, segment := range segments {
		if strings.HasPrefix(segment, ":") && len(segment) > 1 {
			segments[index] = "{" + strings.TrimPrefix(segment, ":") + "}"
			continue
		}
		if strings.HasPrefix(segment, "*") && len(segment) > 1 {
			segments[index] = "{" + strings.TrimPrefix(segment, "*") + "}"
		}
	}
	return strings.Join(segments, "/")
}

// humaError 实现 Huma StatusError，并承载平台统一错误响应字段。
type humaError struct {
	Status  int            `json:"-"`
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
	TraceID string         `json:"trace_id,omitempty"`
}

// Error 返回错误文本，满足 error 接口。
func (e *humaError) Error() string {
	return e.Message
}

// GetStatus 返回 Huma 写入 HTTP 响应时使用的状态码。
func (e *humaError) GetStatus() int {
	return e.Status
}

// ContentType 固定返回 JSON，保证错误响应格式稳定。
func (e *humaError) ContentType(string) string {
	return "application/json"
}

// configureHumaErrors 将 Huma 默认错误适配到脚手架统一错误模型。
func configureHumaErrors() {
	huma.NewError = func(status int, message string, errs ...error) huma.StatusError {
		return newHumaError(status, "", message, errs...)
	}
	huma.NewErrorWithContext = func(ctx huma.Context, status int, message string, errs ...error) huma.StatusError {
		traceID := sharedtypes.TraceIDFromContext(ctx.Context())
		return newHumaError(status, traceID, message, errs...)
	}
}

// newHumaError 根据底层错误类型决定错误码、状态码、细节和 trace_id。
func newHumaError(status int, traceID string, message string, errs ...error) huma.StatusError {
	if len(errs) > 0 {
		if appErr := apperrors.From(errs[0]); appErr != nil {
			actualStatus, body := apperrors.ToHTTP(appErr, traceID)
			return &humaError{
				Status:  actualStatus,
				Code:    body.Code,
				Message: body.Message,
				Details: body.Details,
				TraceID: body.TraceID,
			}
		}
	}

	code := string(apperrors.CodeInternalError)
	if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
		code = string(apperrors.CodeBadRequest)
	}

	details := map[string]any{}
	if len(errs) > 0 {
		messages := make([]string, 0, len(errs))
		for _, err := range errs {
			messages = append(messages, err.Error())
		}
		details["errors"] = messages
	}

	return &humaError{
		Status:  status,
		Code:    code,
		Message: message,
		Details: details,
		TraceID: traceID,
	}
}
