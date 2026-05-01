package api

import "github.com/teamsillybees/initra/pkg/response"

// LoginResponse 描述登录成功响应体。
type LoginResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	User         UserClaims `json:"user"`
}

// RefreshResponse 描述刷新 token 响应体。
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UserClaims 描述当前登录用户的公开身份信息。
type UserClaims struct {
	UserID       int64    `json:"user_id"`
	Username     string   `json:"username"`
	Nickname     string   `json:"nickname"`
	RoleCodes    []string `json:"role_codes"`
	IsSuperAdmin bool     `json:"is_super_admin"`
	IsEnable     bool     `json:"is_enable"`
}

// loginOutput 是 Huma 登录接口响应包装类型。
type loginOutput struct {
	Body response.SuccessResponse[LoginResponse]
}

// refreshOutput 是 Huma 刷新接口响应包装类型。
type refreshOutput struct {
	Body response.SuccessResponse[RefreshResponse]
}

// meOutput 是 Huma 当前用户接口响应包装类型。
type meOutput struct {
	Body response.SuccessResponse[UserClaims]
}
