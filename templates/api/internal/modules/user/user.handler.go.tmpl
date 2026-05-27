package user

import (
	"context"

	platformauth "github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/idgen"
	"github.com/teamsillybees/initra/pkg/pagination"
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
	user, err := h.service.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	return &getUserResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), toUserVO(user)),
	}, nil
}

func (h *Handler) page(ctx context.Context, input *pageUsersRequest) (*pageUsersResponse, error) {
	pageDTO, err := input.PageUsersQuery.Normalize()
	if err != nil {
		return nil, err
	}

	result, err := h.service.Page(ctx, PageUsersDTO{
		Page:    pageDTO,
		Keyword: input.Keyword,
	})
	if err != nil {
		return nil, err
	}

	items := make([]UserVO, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toUserVO(item))
	}

	return &pageUsersResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), pagination.NewPageVO(items, result.Total, result.Page)),
	}, nil
}

func (h *Handler) create(ctx context.Context, input *createUserRequest) (*createUserResponse, error) {
	var operatorID idgen.ID
	if principal, ok := platformauth.PrincipalFromContext(ctx); ok {
		operatorID = principal.UserID
	}

	user, err := h.service.Create(ctx, CreateUserDTO{
		Username:     input.Body.Username,
		Password:     input.Body.Password,
		Nickname:     input.Body.Nickname,
		Phone:        input.Body.Phone,
		Email:        input.Body.Email,
		AvatarURL:    input.Body.AvatarURL,
		RoleCodes:    input.Body.RoleCodes,
		IsSuperAdmin: input.Body.IsSuperAdmin,
		IsEnable:     input.Body.IsEnable,
		SortID:       input.Body.SortID,
		OperatorID:   operatorID,
	})
	if err != nil {
		return nil, err
	}

	return &createUserResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), toUserVO(user)),
	}, nil
}

func (h *Handler) update(ctx context.Context, input *updateUserRequest) (*updateUserResponse, error) {
	var operatorID idgen.ID
	if principal, ok := platformauth.PrincipalFromContext(ctx); ok {
		operatorID = principal.UserID
	}

	user, err := h.service.Update(ctx, UpdateUserDTO{
		ID:           input.ID,
		Nickname:     input.Body.Nickname,
		Phone:        input.Body.Phone,
		Email:        input.Body.Email,
		AvatarURL:    input.Body.AvatarURL,
		RoleCodes:    input.Body.RoleCodes,
		IsSuperAdmin: input.Body.IsSuperAdmin,
		IsEnable:     input.Body.IsEnable,
		SortID:       input.Body.SortID,
		OperatorID:   operatorID,
	})
	if err != nil {
		return nil, err
	}

	return &updateUserResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), toUserVO(user)),
	}, nil
}

func (h *Handler) delete(ctx context.Context, input *deleteUserRequest) (*deleteUserResponse, error) {
	var operatorID idgen.ID
	if principal, ok := platformauth.PrincipalFromContext(ctx); ok {
		operatorID = principal.UserID
	}
	if err := h.service.Delete(ctx, input.ID, operatorID); err != nil {
		return nil, err
	}
	return &deleteUserResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), map[string]any{}),
	}, nil
}

func toUserVO(user *User) UserVO {
	return UserVO{
		ID:           user.ID,
		Username:     user.Username,
		Nickname:     user.Nickname,
		Phone:        user.Phone,
		Email:        user.Email,
		AvatarURL:    user.AvatarURL,
		RoleCodes:    append([]string(nil), user.RoleCodes...),
		IsSuperAdmin: user.IsSuperAdmin,
		IsEnable:     user.IsEnable,
		SortID:       user.SortID,
	}
}
