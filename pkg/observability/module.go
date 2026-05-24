package observability

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/requestctx"
	"github.com/teamsillybees/initra/pkg/response"
	"github.com/teamsillybees/initra/pkg/server"
)

// BuildInfoVO 描述版本、构建时间和提交号等可观测 JSON 信息。
type BuildInfoVO struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

// Module 暴露健康检查、就绪检查与版本信息接口。
type Module struct {
	info BuildInfoVO
}

// NewModule 创建 observability 模块。
func NewModule(info BuildInfoVO) *Module {
	return &Module{info: info}
}

// Register 将健康检查和版本信息接口注册到 Huma。
func (m *Module) Register(api huma.API, registry *server.RouteRegistry) {
	type emptyRequest struct{}
	type healthVO struct {
		Status string `json:"status"`
	}
	type healthResponse struct {
		Body response.SuccessVO[healthVO]
	}
	type versionResponse struct {
		Body response.SuccessVO[BuildInfoVO]
	}

	registerPublic := func(method string, path string) {
		registry.Register(method, path, platformauth.RouteSecurity{AccessMode: platformauth.AccessModePublic})
	}

	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "健康检查",
		Description: "返回服务是否存活。",
		Tags:        []string{"系统观测"},
	}, func(ctx context.Context, _ *emptyRequest) (*healthResponse, error) {
		return &healthResponse{
			Body: response.OK(requestctx.TraceIDFromContext(ctx), healthVO{Status: "ok"}),
		}, nil
	})
	registerPublic(http.MethodGet, "/health")

	huma.Register(api, huma.Operation{
		OperationID: "ready-check",
		Method:      http.MethodGet,
		Path:        "/ready",
		Summary:     "就绪检查",
		Description: "返回服务是否完成初始化并具备对外提供服务的能力。",
		Tags:        []string{"系统观测"},
	}, func(ctx context.Context, _ *emptyRequest) (*healthResponse, error) {
		return &healthResponse{
			Body: response.OK(requestctx.TraceIDFromContext(ctx), healthVO{Status: "ready"}),
		}, nil
	})
	registerPublic(http.MethodGet, "/ready")

	huma.Register(api, huma.Operation{
		OperationID: "version-info",
		Method:      http.MethodGet,
		Path:        "/version",
		Summary:     "版本信息",
		Description: "返回构建版本、提交号和构建时间。",
		Tags:        []string{"系统观测"},
	}, func(ctx context.Context, _ *emptyRequest) (*versionResponse, error) {
		return &versionResponse{
			Body: response.OK(requestctx.TraceIDFromContext(ctx), m.info),
		}, nil
	})
	registerPublic(http.MethodGet, "/version")
}
