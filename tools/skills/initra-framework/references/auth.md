# auth

JWT、refresh token store、密码哈希、Casbin enforcer、路由安全元信息、认证中间件和授权中间件使用 `github.com/teamsillybees/initra/pkg/auth` 与 `pkg/server`。

## 标准装配

启用 Redis token store 时，先注册 Redis，再注册 auth：

```go
redisx.Register(injector, cfg.Redis)
auth.Register(injector, auth.RegisterOptions{
	AppName:      cfg.App.Name,
	Env:          cfg.App.Env,
	RedisEnabled: cfg.Redis.Enabled,
	JWT: auth.JWTConfig{
		Issuer:          cfg.Auth.JWT.Issuer,
		Secret:          cfg.Auth.JWT.Secret,
		AccessTokenTTL:  cfg.Auth.AccessTokenTTL,
		RefreshTokenTTL: cfg.Auth.RefreshTokenTTL,
	},
	CasbinModelPath:  cfg.Casbin.ModelPath,
	CasbinPolicyPath: cfg.Casbin.PolicyPath,
})
```

`auth.Register` 会提供 `*auth.BcryptPasswordManager`、`*auth.JWTManager` 和 `*casbin.Enforcer`。Redis 启用时 token store 使用 Redis，否则使用 memory store。

## 路由安全

每个 `/api/` 路由都必须注册安全元信息。缺失元信息时，中间件默认 fail-closed。

```go
registry.Register(http.MethodPost, "/api/v1/auth/login", platformauth.RouteSecurity{AccessMode: platformauth.AccessModePublic})
registry.Register(http.MethodGet, "/api/v1/auth/me", platformauth.RouteSecurity{AccessMode: platformauth.AccessModeAuthenticated})
registry.Register(http.MethodGet, "/api/v1/users/{id}", platformauth.RouteSecurity{
	AccessMode: platformauth.AccessModePermission,
	Resource:   "user",
	Action:     "read",
})
```

`AccessModePublic` 用于登录、注册、验证码和公开内容；`AccessModeAuthenticated` 用于登录即可访问的 ToC 接口；`AccessModePermission` 用于后台管理、运营操作、审核、退款、风控、配置管理等需要 RBAC 授权的接口。只有 `AccessModePermission` 需要配置 `Resource` 和 `Action`，并且必须与 `configs/rbac_policy.csv` 保持一致。

## 业务用法

- 使用 `auth.BcryptPasswordManager` 做密码哈希和校验。
- 使用 `auth.JWTManager` 签发、刷新和解析 access/refresh token。
- 登录和刷新接口只有在显式注册为 `AccessModePublic` 时才公开。
- 当前用户身份从模板已有 request context helper 中读取。

## 禁止写法

- 不要在 handler 中绕过中间件。
- 不要在业务模块中自定义 JWT 签名逻辑。
- 不要在 handler 中硬编码 Casbin 决策。
- 不要记录原始 JWT、refresh token、密码哈希或凭证。
- 不要留下未注册到 route registry 的 `/api/` 路由。

## 检查清单

- 启用 Redis token store 时，auth 是否在 Redis 之后注册？
- 所有 API 路由是否都在 `RouteRegistry` 中？
- 公开路由是否显式设置 `AccessModePublic`？
- 登录即可访问的路由是否使用 `AccessModeAuthenticated`，而不是误用 RBAC？
- `AccessModePermission` 路由的 Casbin resource/action 是否与 policy 匹配？
- 认证错误是否返回统一应用错误？
