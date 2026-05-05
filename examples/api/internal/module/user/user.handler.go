package user

import (
	"context"

	platformauth "github.com/teamsillybees/initra/pkg/auth"
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

// Get 返回用户详情。
func (h *Handler) Get(ctx context.Context, input *GetUserInput) (*GetUserOutput, error) {
	user, err := h.service.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	return &GetUserOutput{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), toUserResponse(user)),
	}, nil
}

// List 返回用户分页列表。
func (h *Handler) List(ctx context.Context, input *ListUsersInput) (*ListUsersOutput, error) {
	result, err := h.service.List(ctx, ListUsersParams{
		Page:     input.Page,
		PageSize: input.PageSize,
		Keyword:  input.Keyword,
	})
	if err != nil {
		return nil, err
	}

	items := make([]UserResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toUserResponse(item))
	}

	return &ListUsersOutput{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), UserListResponse{
			Items:    items,
			Total:    result.Total,
			Page:     result.Page,
			PageSize: result.PageSize,
		}),
	}, nil
}

// Create 创建用户。
func (h *Handler) Create(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
	operatorID := int64(0)
	if principal, ok := platformauth.PrincipalFromContext(ctx); ok {
		operatorID = principal.UserID
	}

	user, err := h.service.Create(ctx, CreateUserParams{
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

	return &CreateUserOutput{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), toUserResponse(user)),
	}, nil
}

// Update 更新用户。
func (h *Handler) Update(ctx context.Context, input *UpdateUserInput) (*UpdateUserOutput, error) {
	operatorID := int64(0)
	if principal, ok := platformauth.PrincipalFromContext(ctx); ok {
		operatorID = principal.UserID
	}

	user, err := h.service.Update(ctx, UpdateUserParams{
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

	return &UpdateUserOutput{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), toUserResponse(user)),
	}, nil
}

// Delete 删除用户。
func (h *Handler) Delete(ctx context.Context, input *DeleteUserInput) (*DeleteUserOutput, error) {
	operatorID := int64(0)
	if principal, ok := platformauth.PrincipalFromContext(ctx); ok {
		operatorID = principal.UserID
	}
	if err := h.service.Delete(ctx, input.ID, operatorID); err != nil {
		return nil, err
	}
	return &DeleteUserOutput{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), map[string]any{}),
	}, nil
}

func toUserResponse(user *User) UserResponse {
	return UserResponse{
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
