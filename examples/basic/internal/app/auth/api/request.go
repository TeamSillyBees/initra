package api

// LoginRequest 描述登录请求体。
type LoginRequest struct {
	Username string `json:"username" example:"alice"`
	Password string `json:"password" example:"secret-123"`
}

// LoginInput 描述登录接口输入。
type LoginInput struct {
	Body LoginRequest
}

// RefreshRequest 描述刷新 token 请求体。
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" example:"token"`
}

// RefreshInput 描述刷新接口输入。
type RefreshInput struct {
	Body RefreshRequest
}

// MeInput 仅作为当前用户接口的占位输入。
type MeInput struct{}
