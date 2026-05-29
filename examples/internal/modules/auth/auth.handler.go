package auth

import (
	"context"

	"github.com/teamsillybees/initra/examples/internal/modules/bizerrors"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/requestctx"
	"github.com/teamsillybees/initra/pkg/response"
)

// Handler 封装 auth 模块的 HTTP 适配逻辑。
type Handler struct {
	service *Service
}

// NewHandler 创建 auth 模块 Handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) login(ctx context.Context, input *loginRequest) (*loginResponse, error) {
	vo, err := h.service.Login(ctx, input.Body)
	if err != nil {
		return nil, err
	}
	return &loginResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), vo),
	}, nil
}

func (h *Handler) refresh(ctx context.Context, input *refreshRequest) (*refreshResponse, error) {
	vo, err := h.service.Refresh(ctx, input.Body)
	if err != nil {
		return nil, err
	}
	return &refreshResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), vo),
	}, nil
}

func (h *Handler) me(ctx context.Context, _ *meRequest) (*meResponse, error) {
	principal, ok := platformauth.PrincipalFromContext(ctx)
	if !ok {
		return nil, bizerrors.Unauthorized("user principal is missing")
	}

	vo, err := h.service.Me(ctx, principal.UserID)
	if err != nil {
		return nil, err
	}
	return &meResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), vo),
	}, nil
}
