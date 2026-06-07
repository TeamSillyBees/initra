# 认证、授权与路由安全

## Boot 注册

```go
auth.Register(injector, auth.RegisterOptions{
	AppName:          cfg.App.Name,
	Env:              cfg.App.Env,
	RedisEnabled:     cfg.Redis.Enabled,
	JWT:              cfg.Auth.JWT,
	CasbinModelPath:  cfg.Casbin.ModelPath,
	CasbinPolicyPath: cfg.Casbin.PolicyPath,
})
server.Register(injector, server.Options{Title: cfg.App.Name, Version: version, Env: cfg.App.Env})
```

Redis 启用时 refresh token store 使用 Redis；否则退化为内存 store。

## RouteSecurity

所有 `/api/` 路由都必须注册安全元信息：

```go
registry.Register(http.MethodGet, "/api/v1/users/{id}", platformauth.RouteSecurity{
	AccessMode: platformauth.AccessModePermission,
	Resource:   "user",
	Action:     "read",
})
```

- 登录、注册、验证码等公开接口使用 `AccessModePublic`。
- 只需要登录态的 ToC 接口使用 `AccessModeAuthenticated`。
- 后台管理、运营、审核、退款、风控、配置管理等接口使用 `AccessModePermission`，并匹配 Casbin policy。

## 禁止

- 不要在 handler 层绕过统一 JWT/Casbin 中间件。
- 不要让 `/api/` 路由缺少 `RouteSecurity`；中间件按 fail-closed 处理。
- 不要把 token 或 session value 写入日志。
