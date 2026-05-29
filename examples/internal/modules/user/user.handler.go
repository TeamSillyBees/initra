package user

import (
	"context"

	platformauth "github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/idgen"
	"github.com/teamsillybees/initra/pkg/requestctx"
	"github.com/teamsillybees/initra/pkg/response"
)

// Handler 封装 user 模块的 HTTP 适配逻辑。
type Handler struct {
	service *Service
}

// NewHandler 创建 user 模块 HTTP Handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) get(ctx context.Context, input *getUserRequest) (*getUserResponse, error) {
	vo, err := h.service.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	return &getUserResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), vo),
	}, nil
}

func (h *Handler) page(ctx context.Context, input *pageUsersRequest) (*pageUsersResponse, error) {
	pageVO, err := h.service.Page(ctx, input.PageUsersQuery)
	if err != nil {
		return nil, err
	}
	return &pageUsersResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), pageVO),
	}, nil
}

func (h *Handler) create(ctx context.Context, input *createUserRequest) (*createUserResponse, error) {
	vo, err := h.service.Create(ctx, input.Body, operatorID(ctx))
	if err != nil {
		return nil, err
	}
	return &createUserResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), vo),
	}, nil
}

func (h *Handler) update(ctx context.Context, input *updateUserRequest) (*updateUserResponse, error) {
	vo, err := h.service.Update(ctx, input.ID, input.Body, operatorID(ctx))
	if err != nil {
		return nil, err
	}
	return &updateUserResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), vo),
	}, nil
}

func (h *Handler) delete(ctx context.Context, input *deleteUserRequest) (*deleteUserResponse, error) {
	if err := h.service.Delete(ctx, input.ID, operatorID(ctx)); err != nil {
		return nil, err
	}
	return &deleteUserResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), map[string]any{}),
	}, nil
}

func operatorID(ctx context.Context) idgen.ID {
	if principal, ok := platformauth.PrincipalFromContext(ctx); ok {
		return principal.UserID
	}
	return 0
}
