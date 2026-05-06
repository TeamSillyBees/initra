package auth

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/server"
)

// Module 负责 auth 模块路由注册。
type Module struct {
	handler *Handler
}

// NewModule 创建 auth 模块实例。
func NewModule(handler *Handler) *Module {
	return &Module{handler: handler}
}

// Register 将 auth 模块注册到应用。
func (m *Module) Register(api huma.API, registry *server.RouteRegistry) {
	huma.Register(api, huma.Operation{
		OperationID: "auth-login",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/login",
		Summary:     "用户登录",
		Description: "根据账号密码签发访问令牌和刷新令牌。",
		Tags:        []string{"认证管理"},
	}, m.handler.login)
	registry.Register(http.MethodPost, "/api/v1/auth/login", platformauth.RouteSecurity{Public: true})

	huma.Register(api, huma.Operation{
		OperationID: "auth-refresh",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/refresh",
		Summary:     "刷新令牌",
		Description: "使用 refresh token 换取新的访问令牌。",
		Tags:        []string{"认证管理"},
	}, m.handler.refresh)
	registry.Register(http.MethodPost, "/api/v1/auth/refresh", platformauth.RouteSecurity{Public: true})

	huma.Register(api, huma.Operation{
		OperationID: "auth-me",
		Method:      http.MethodGet,
		Path:        "/api/v1/auth/me",
		Summary:     "获取当前登录用户",
		Description: "读取当前访问令牌对应的用户信息。",
		Tags:        []string{"认证管理"},
	}, m.handler.me)
	registry.Register(http.MethodGet, "/api/v1/auth/me", platformauth.RouteSecurity{Resource: "auth", Action: "read"})
}
