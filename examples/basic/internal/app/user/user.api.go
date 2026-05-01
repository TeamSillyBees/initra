package user

import (
	"context"

	platformauth "github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/requestctx"
	"github.com/teamsillybees/initra/pkg/response"
)

// --- Request DTOs ---

// GetUserInput 描述用户详情查询的路径参数。
type GetUserInput struct {
	ID int64 `path:"id" example:"1001" doc:"用户 ID"`
}

// ListUsersInput 描述用户列表分页查询参数。
type ListUsersInput struct {
	Page     int    `query:"page" example:"1" doc:"页码"`
	PageSize int    `query:"page_size" example:"20" doc:"每页数量"`
	Keyword  string `query:"keyword" example:"alice" doc:"用户名、昵称、手机号或邮箱关键字"`
}

// CreateUserRequest 描述创建用户请求体。
type CreateUserRequest struct {
	Username     string   `json:"username" example:"alice"`
	Password     string   `json:"password" example:"secret-123"`
	Nickname     string   `json:"nickname" example:"Alice"`
	Phone        string   `json:"phone" example:"13800000000"`
	Email        string   `json:"email" example:"alice@example.com"`
	AvatarURL    string   `json:"avatar_url" example:"https://example.com/avatar.png"`
	RoleCodes    []string `json:"role_codes" example:"admin"`
	IsSuperAdmin bool     `json:"is_super_admin" example:"false"`
	IsEnable     *bool    `json:"is_enable,omitempty" example:"true"`
	SortID       int      `json:"sort_id" example:"0"`
}

// CreateUserInput 描述创建用户接口输入。
type CreateUserInput struct {
	Body CreateUserRequest
}

// UpdateUserRequest 描述更新用户请求体。
type UpdateUserRequest struct {
	Nickname     *string   `json:"nickname,omitempty" example:"Alice Updated"`
	Phone        *string   `json:"phone,omitempty" example:"13800000001"`
	Email        *string   `json:"email,omitempty" example:"alice.updated@example.com"`
	AvatarURL    *string   `json:"avatar_url,omitempty" example:"https://example.com/avatar-updated.png"`
	RoleCodes    *[]string `json:"role_codes,omitempty" example:"viewer"`
	IsSuperAdmin *bool     `json:"is_super_admin,omitempty" example:"false"`
	IsEnable     *bool     `json:"is_enable,omitempty" example:"true"`
	SortID       *int      `json:"sort_id,omitempty" example:"10"`
}

// UpdateUserInput 描述更新用户接口输入。
type UpdateUserInput struct {
	ID   int64 `path:"id" example:"1001" doc:"用户 ID"`
	Body UpdateUserRequest
}

// DeleteUserInput 描述删除用户接口输入。
type DeleteUserInput struct {
	ID int64 `path:"id" example:"1001" doc:"用户 ID"`
}

// --- Response DTOs ---

// UserResponse 是 user 模块对外暴露的用户 DTO。
type UserResponse struct {
	ID           int64    `json:"id"`
	Username     string   `json:"username"`
	Nickname     string   `json:"nickname"`
	Phone        string   `json:"phone"`
	Email        string   `json:"email"`
	AvatarURL    string   `json:"avatar_url"`
	RoleCodes    []string `json:"role_codes"`
	IsSuperAdmin bool     `json:"is_super_admin"`
	IsEnable     bool     `json:"is_enable"`
	SortID       int      `json:"sort_id"`
}

// UserListResponse 是用户分页列表 DTO。
type UserListResponse struct {
	Items    []UserResponse `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// --- Huma output types ---

type getUserOutput struct {
	Body response.SuccessResponse[UserResponse]
}

type listUsersOutput struct {
	Body response.SuccessResponse[UserListResponse]
}

type createUserOutput struct {
	Body response.SuccessResponse[UserResponse]
}

type updateUserOutput struct {
	Body response.SuccessResponse[UserResponse]
}

type deleteUserOutput struct {
	Body response.SuccessResponse[map[string]any]
}

// --- Handler ---

// Handler 封装 user 模块的 HTTP 适配逻辑。
type Handler struct {
	service *Service
}

// NewHandler 创建 user 模块 HTTP Handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Get 返回用户详情。
func (h *Handler) Get(ctx context.Context, input *GetUserInput) (*getUserOutput, error) {
	user, err := h.service.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	return &getUserOutput{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), toUserResponse(user)),
	}, nil
}

// List 返回用户分页列表。
func (h *Handler) List(ctx context.Context, input *ListUsersInput) (*listUsersOutput, error) {
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

	return &listUsersOutput{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), UserListResponse{
			Items:    items,
			Total:    result.Total,
			Page:     result.Page,
			PageSize: result.PageSize,
		}),
	}, nil
}

// Create 创建用户。
func (h *Handler) Create(ctx context.Context, input *CreateUserInput) (*createUserOutput, error) {
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

	return &createUserOutput{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), toUserResponse(user)),
	}, nil
}

// Update 更新用户。
func (h *Handler) Update(ctx context.Context, input *UpdateUserInput) (*updateUserOutput, error) {
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

	return &updateUserOutput{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), toUserResponse(user)),
	}, nil
}

// Delete 删除用户。
func (h *Handler) Delete(ctx context.Context, input *DeleteUserInput) (*deleteUserOutput, error) {
	operatorID := int64(0)
	if principal, ok := platformauth.PrincipalFromContext(ctx); ok {
		operatorID = principal.UserID
	}
	if err := h.service.Delete(ctx, input.ID, operatorID); err != nil {
		return nil, err
	}
	return &deleteUserOutput{
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
