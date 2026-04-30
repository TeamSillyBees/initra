# 架构说明

`initra` 采用“平台层 + 业务模块”的垂直切片架构。

## 目录边界

- `cmd/server`：程序入口，只负责启动流程与信号处理。
- `internal/boot`：应用组合根，负责加载配置、初始化依赖注入容器、注册业务模块并构建 HTTP Server。
- `internal/platform`：配置、日志、数据库、缓存、鉴权、Web 装配、可观测能力等基础设施，不反向依赖业务模块。
- `internal/app/auth`：登录、刷新 token、当前用户信息。
- `internal/app/user`：用户 CRUD 与用户详情缓存。
- `internal/gen/jet`：Jet 生成代码目录，当前仓库内置最小占位代码，正式开发请使用 `make jet` 重新生成。

## 核心链路

1. `internal/boot.Bootstrap` 加载配置并初始化 do 容器。
2. 平台层创建 logger、DB、Redis、cache manager、JWT、Casbin、Gin/Huma。
3. `observability`、`auth`、`user` 模块向 Huma 注册 operation，同时向路由注册表登记安全策略。
4. Gin 中间件依次完成 recovery、trace/request id 注入、请求日志、CORS、JWT 认证、Casbin 授权。
5. 业务 handler 只负责 DTO 与上下文接入，具体业务编排落在 domain service，持久化访问通过 infra repository 完成。

## 新增业务功能扩展点

新增业务功能时优先新增一个垂直切片模块，而不是把代码散落到平台层：

- `internal/app/<module>/domain`：定义领域实体、service、仓储接口、缓存接口和业务错误。
- `internal/app/<module>/api`：定义 Huma input/output、HTTP handler 和 DTO 转换。
- `internal/app/<module>/infra`：实现 domain 所需接口，例如 Jet repository、Redis cache、第三方客户端适配。
- `internal/app/<module>/wire.go`：把模块内部依赖注册进 do 容器；跨模块同类型依赖必须使用命名服务。
- `internal/app/<module>/module.go`：注册 Huma operation，并同步登记 `RouteSecurity`，保证接口文档和授权策略一致。
- `internal/boot/app.go`：新增模块的 `Provide` 和 `Register` 都必须在组合根接入，否则依赖不会初始化、路由也不会暴露。
- `db/schema`、`db/migrations`、`internal/gen/jet`：有表结构变更时按 schema -> migration -> Jet 生成代码的顺序推进。
- `configs/rbac_policy.csv`：新增受保护资源时维护 Casbin policy；公开接口只在路由注册处显式设置 `Public: true`。
- `test` 与模块内测试：业务规则放 service 单测，SQL 行为放 integration，完整 HTTP/鉴权链路放 e2e。

## 约束

- service 不接收 `gin.Context`。
- handler 不写业务逻辑。
- Repository 只使用 Jet 生成 SQL。
- 错误统一通过 `internal/platform/errors` 归一化。
- 成功/失败响应都统一带 `trace_id`。
- `internal/platform` 不允许导入 `internal/app/*`，该规则由 `internal/architecture` 测试固定。
