package auth

import "github.com/teamsillybees/initra/pkg/response"

// LoginBody 描述登录请求体。
type LoginBody struct {
	Username string `json:"username" example:"admin"`
	Password string `json:"password" example:"admin123"`
}

type loginRequest struct {
	Body LoginBody
}

// RefreshBody 描述刷新 token 请求体。
type RefreshBody struct {
	RefreshToken string `json:"refreshToken" example:"token"`
}

type refreshRequest struct {
	Body RefreshBody
}

type meRequest struct{}

// LoginVO 描述登录成功响应体。
type LoginVO struct {
	AccessToken  string         `json:"accessToken"`
	RefreshToken string         `json:"refreshToken"`
	User         UserIdentityVO `json:"user"`
}

// RefreshVO 描述刷新 token 响应体。
type RefreshVO struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// UserIdentityVO 描述当前登录用户的公开身份信息。
type UserIdentityVO struct {
	UserID       int64    `json:"userId"`
	Username     string   `json:"username"`
	Nickname     string   `json:"nickname"`
	RoleCodes    []string `json:"roleCodes"`
	IsSuperAdmin bool     `json:"isSuperAdmin"`
	IsEnable     bool     `json:"isEnable"`
}

type loginResponse struct {
	Body response.SuccessVO[LoginVO]
}

type refreshResponse struct {
	Body response.SuccessVO[RefreshVO]
}

type meResponse struct {
	Body response.SuccessVO[UserIdentityVO]
}
