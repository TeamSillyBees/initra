package taskdemo

import (
	"context"

	"github.com/teamsillybees/initra/pkg/requestctx"
	"github.com/teamsillybees/initra/pkg/response"
)

// Handler 封装 taskdemo 示例模块的 HTTP 适配逻辑。
type Handler struct {
	service *Service
}

// NewHandler 创建 taskdemo 示例模块 HTTP Handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) publishEmail(ctx context.Context, input *publishEmailRequest) (*publishEmailResponse, error) {
	vo, err := h.service.PublishEmail(ctx, input.Body, requestctx.TraceIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	return &publishEmailResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), vo),
	}, nil
}
