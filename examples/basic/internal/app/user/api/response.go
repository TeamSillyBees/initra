package api

import "github.com/teamsillybees/initra/pkg/response"

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

// getUserOutput 是 Huma 用户详情接口响应包装类型。
type getUserOutput struct {
	Body response.SuccessResponse[UserResponse]
}

// listUsersOutput 是 Huma 用户列表接口响应包装类型。
type listUsersOutput struct {
	Body response.SuccessResponse[UserListResponse]
}

// createUserOutput 是 Huma 创建用户接口响应包装类型。
type createUserOutput struct {
	Body response.SuccessResponse[UserResponse]
}

// updateUserOutput 是 Huma 更新用户接口响应包装类型。
type updateUserOutput struct {
	Body response.SuccessResponse[UserResponse]
}

// deleteUserOutput 是 Huma 删除用户接口响应包装类型。
type deleteUserOutput struct {
	Body response.SuccessResponse[map[string]any]
}
