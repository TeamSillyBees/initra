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

## 约束

- service 不接收 `gin.Context`。
- handler 不写业务逻辑。
- Repository 只使用 Jet 生成 SQL。
- 错误统一通过 `internal/platform/errors` 归一化。
- 成功/失败响应都统一带 `trace_id`。
- `internal/platform` 不允许导入 `internal/app/*`，该规则由 `internal/architecture` 测试固定。
