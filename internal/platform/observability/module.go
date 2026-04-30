package observability

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	platformauth "github.com/teamsillybees/initra/internal/platform/auth"
	"github.com/teamsillybees/initra/internal/platform/web"
	sharedtypes "github.com/teamsillybees/initra/internal/shared/types"
)

// BuildInfo 描述版本、构建时间和提交号等可观测信息。
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}

// Module 暴露健康检查、就绪检查与版本信息接口。
type Module struct {
	info BuildInfo
}

// NewModule 创建 observability 模块。
func NewModule(info BuildInfo) *Module {
	return &Module{info: info}
}

// Register 将健康检查和版本信息接口注册到 Huma。
func (m *Module) Register(api huma.API, registry *web.RouteRegistry) {
	// healthData 是健康检查响应的最小载荷，避免公开内部依赖状态。
	type healthData struct {
		Status string `json:"status"`
	}
	// healthOutput 包装 Huma 所需的响应结构。
	type healthOutput struct {
		Body sharedtypes.SuccessResponse[healthData]
	}
	// versionOutput 包装版本接口响应结构。
	type versionOutput struct {
		Body sharedtypes.SuccessResponse[BuildInfo]
	}

	registerPublic := func(method string, path string) {
		registry.Register(method, path, platformauth.RouteSecurity{Public: true})
	}

	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "健康检查",
		Description: "返回服务是否存活。",
		Tags:        []string{"系统观测"},
	}, func(ctx context.Context, _ *struct{}) (*healthOutput, error) {
		return &healthOutput{
			Body: sharedtypes.OK(sharedtypes.TraceIDFromContext(ctx), healthData{Status: "ok"}),
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
	}, func(ctx context.Context, _ *struct{}) (*healthOutput, error) {
		return &healthOutput{
			Body: sharedtypes.OK(sharedtypes.TraceIDFromContext(ctx), healthData{Status: "ready"}),
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
	}, func(ctx context.Context, _ *struct{}) (*versionOutput, error) {
		return &versionOutput{
			Body: sharedtypes.OK(sharedtypes.TraceIDFromContext(ctx), m.info),
		}, nil
	})
	registerPublic(http.MethodGet, "/version")
}
