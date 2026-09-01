# 认证、授权与路由安全

## Boot 注册

标准项目先注册 `internal/accesscontrol`，由它提供 `auth.PolicyLoader` 和 `auth.IdentityResolver`，再注册认证与 Web 层：

```go
accesscontrol.Provide(injector, accesscontrol.Options{
	AppName:      cfg.App.Slug,
	Env:          cfg.App.Env,
	InstanceID:   cfg.App.InstanceID,
	CacheTTL:     cfg.Cache.RemoteTTL,
	RedisEnabled: cfg.Redis.Enabled,
})
auth.Register(injector, auth.RegisterOptions{
	AppName:          cfg.App.Slug,
	Env:              cfg.App.Env,
	RedisEnabled:     cfg.Redis.Enabled,
	AllowMemoryStore: cfg.Auth.AllowMemoryTokenStore,
	JWT: auth.JWTConfig{
		Issuer:          cfg.Auth.JWT.Issuer,
		Secret:          cfg.Auth.JWT.Secret,
		AccessTokenTTL:  cfg.Auth.AccessTokenTTL,
		RefreshTokenTTL: cfg.Auth.RefreshTokenTTL,
	},
	CasbinModelPath: cfg.Casbin.ModelPath,
})
server.Register(injector, server.Options{Title: cfg.App.Name, Version: version, Env: cfg.App.Env})
```

Redis 启用时 refresh token store、请求身份缓存和多实例权限变更通知都使用 Redis。Redis 关闭时不会隐式退化：只有 `AllowMemoryStore=true` 且环境是 dev、local 或 test 才允许进程内 token store；请求身份直接查库。共享环境和多副本部署必须使用共享 Redis。

access JWT 只保存用户和会话身份，不保存角色或权限。角色禁用、收回或用户禁用后，请求时解析出的当前身份立即影响授权。

## RouteSecurity

所有 `/api/` 路由都必须注册安全元信息。权限路由直接登记稳定权限标识：

```go
registry.Register(http.MethodGet, "/api/v1/users/{id}", platformauth.RouteSecurity{
	AccessMode: platformauth.AccessModePermission,
	Permission: "system:user:read",
})
```

- 登录、注册、验证码等公开接口使用 `AccessModePublic`。
- 只需要登录态的 ToC 接口使用 `AccessModeAuthenticated`。
- 后台管理、运营、审核、退款、风控、配置管理等接口使用 `AccessModePermission`。
- 权限唯一事实源是数据库中的 `sys_role`、`sys_menu`、`sys_role_menu`；禁止增加静态 Casbin policy 文件或在代码中维护第二份角色权限映射。

## 禁止

- 不要在 handler 层绕过统一 JWT/Casbin 中间件。
- 不要让 `/api/` 路由缺少 `RouteSecurity`；中间件按 fail-closed 处理。
- 不要把角色或权限固化进 access JWT。
- 不要把 token、session value 或 Redis 身份缓存内容写入日志。
