package auth

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

// LoginDTO 描述登录输入参数。
type LoginDTO struct {
	Username string
	Password string
}
