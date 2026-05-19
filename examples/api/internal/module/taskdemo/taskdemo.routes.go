package taskdemo

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/server"
)

// Module 负责 taskdemo 示例模块路由注册。
type Module struct {
	handler *Handler
}

// NewModule 创建 taskdemo 示例模块实例。
func NewModule(handler *Handler) *Module {
	return &Module{handler: handler}
}

// Register 将 taskdemo 示例模块的 Huma operation 和安全策略注册到应用。
func (m *Module) Register(api huma.API, registry *server.RouteRegistry) {
	huma.Register(api, huma.Operation{
		OperationID: "taskdemo-publish-email",
		Method:      http.MethodPost,
		Path:        "/api/v1/task-demo/email",
		Summary:     "任务队列发布示例",
		Description: "通过 pkg/task 发布 demo:send_email 异步任务。",
		Tags:        []string{"任务队列示例"},
	}, m.handler.publishEmail)
	registry.Register(http.MethodPost, "/api/v1/task-demo/email", platformauth.RouteSecurity{Resource: "taskdemo", Action: "create"})
}
