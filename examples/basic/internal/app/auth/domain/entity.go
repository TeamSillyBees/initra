package domain

// Identity 描述 auth 模块关心的最小用户身份聚合。
type Identity struct {
	UserID       int64
	Username     string
	Nickname     string
	PasswordHash string
	RoleCodes    []string
	IsSuperAdmin bool
	IsEnable     bool
}

// LoginInput 描述登录输入。
type LoginInput struct {
	Username string
	Password string
}

// AuthenticatedUser 描述认证成功后返回给客户端的用户信息。
type AuthenticatedUser struct {
	UserID       int64    `json:"user_id"`
	Username     string   `json:"username"`
	Nickname     string   `json:"nickname"`
	RoleCodes    []string `json:"role_codes"`
	IsSuperAdmin bool     `json:"is_super_admin"`
	IsEnable     bool     `json:"is_enable"`
}

// LoginResult 描述登录成功后的响应载荷。
type LoginResult struct {
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token"`
	User         AuthenticatedUser `json:"user"`
}

// RefreshResult 描述刷新 token 的响应载荷。
type RefreshResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
