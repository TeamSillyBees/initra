package file

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/server"
)

// Module 负责 file 示例模块路由注册。
type Module struct {
	handler *Handler
}

// NewModule 创建 file 示例模块实例。
func NewModule(handler *Handler) *Module {
	return &Module{handler: handler}
}

// Register 将 file 示例模块的 Huma operation 和安全策略注册到应用。
func (m *Module) Register(api huma.API, registry *server.RouteRegistry) {
	huma.Register(api, huma.Operation{
		OperationID:   "upload-local-file",
		Method:        http.MethodPost,
		Path:          "/api/v1/files/local",
		Summary:       "上传本地文件",
		Description:   "上传文件到当前配置的 local 存储目录。",
		Tags:          []string{"文件示例"},
		DefaultStatus: http.StatusCreated,
	}, m.handler.upload)
	registry.Register(http.MethodPost, "/api/v1/files/local", platformauth.RouteSecurity{Resource: "file", Action: "write"})

	huma.Register(api, huma.Operation{
		OperationID: "download-local-file",
		Method:      http.MethodGet,
		Path:        "/api/v1/files/local/download",
		Summary:     "下载本地文件",
		Description: "通过对象 key 下载 local 存储中的文件。",
		Tags:        []string{"文件示例"},
	}, m.handler.download)
	registry.Register(http.MethodGet, "/api/v1/files/local/download", platformauth.RouteSecurity{Resource: "file", Action: "read"})

	huma.Register(api, huma.Operation{
		OperationID: "stat-local-file",
		Method:      http.MethodGet,
		Path:        "/api/v1/files/local/meta",
		Summary:     "查询本地文件元信息",
		Description: "通过对象 key 查询 local 存储中的文件元信息。",
		Tags:        []string{"文件示例"},
	}, m.handler.stat)
	registry.Register(http.MethodGet, "/api/v1/files/local/meta", platformauth.RouteSecurity{Resource: "file", Action: "read"})

	huma.Register(api, huma.Operation{
		OperationID: "delete-local-file",
		Method:      http.MethodDelete,
		Path:        "/api/v1/files/local",
		Summary:     "删除本地文件",
		Description: "通过对象 key 删除 local 存储中的文件。",
		Tags:        []string{"文件示例"},
	}, m.handler.delete)
	registry.Register(http.MethodDelete, "/api/v1/files/local", platformauth.RouteSecurity{Resource: "file", Action: "delete"})
}
