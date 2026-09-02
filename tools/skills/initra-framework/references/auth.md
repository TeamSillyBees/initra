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
	LoginProtection: cfg.Auth.LoginProtection,
})
server.Register(injector, server.Options{
	Title: cfg.App.Name, Version: version, Env: cfg.App.Env,
	TrustedProxies: cfg.Server.TrustedProxies,
	CORS: cfg.Server.CORS,
	Docs: cfg.Server.Docs,
})
```

Redis 启用时 refresh token store、请求身份缓存和多实例权限变更通知都使用 Redis。Redis 关闭时不会隐式退化：只有 `AllowMemoryStore=true` 且环境是 dev、local 或 test 才允许进程内 token store；请求身份直接查库。共享环境和多副本部署必须使用共享 Redis。

access JWT 只保存 `userId`、`sessionId`、`sessionVersion` 和标准声明，不保存角色、权限或租户快照。`logout-all`、修改密码或禁用账号会递增 `sys_user.session_version`；中间件和 refresh 轮转都校验当前版本。角色禁用、收回或用户禁用后，请求时解析出的当前身份立即影响授权。

标准 auth 模块提供登录、refresh、当前用户、`logout`、`logout-all` 和修改密码。`logout` 原子消费当前 refresh token 并拉黑配对 access JTI；全量退出和改密通过用户级会话版本撤销全部旧会话。登录前必须调用 `auth.LoginGuard`，按账号和可信来源 IP 限流，失败时累计锁定，成功后清除连续失败状态；所有凭证类拒绝统一返回 `LOGIN_FAILED`，详细原因只进入安全审计日志。

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
- `RouteSecurity` 同时驱动 OpenAPI operation 的 Bearer security 和 401/403 响应，不要手写与运行时不一致的安全声明。

## HTTP 暴露

- `server.cors` 必须使用明确 origin、method、header 白名单；共享环境禁止通配符。
- `server.trusted_proxies` 只登记实际受信任的反向代理地址或 CIDR，登录 IP 防护不会无条件信任转发头。
- `server.docs.enabled` 只允许在 dev/local/test 开启；共享环境关闭 `/docs`、`/openapi*`、`/schemas/*`。

## 禁止

- 不要在 handler 层绕过统一 JWT/Casbin 中间件。
- 不要让 `/api/` 路由缺少 `RouteSecurity`；中间件按 fail-closed 处理。
- 不要把角色或权限固化进 access JWT。
- 不要把 token、session value 或 Redis 身份缓存内容写入日志。
