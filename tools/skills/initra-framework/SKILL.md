---
name: initra-framework
description: Use when working in a Go codebase that depends on github.com/teamsillybees/initra, especially when adding or reviewing business modules, boot providers, route security, idgen.ID, Ent schema, config, logx, auth, redisx, cache, httpclient, storage, task queues, migrations, or code that might reimplement initra pkg capabilities.
---

# initra-framework

## 目的

这个 skill 只解决一个问题：在使用 initra 的 Go 业务项目中，避免 agent 绕开框架、复制框架、重新发明框架。

initra 的业务项目不是普通 Go Web 项目。它的正确性来自几个边界：业务模块是 flat package；基础设施只在 `internal/boot` 装配；业务 ID 是 `idgen.ID`；`/api/` 路由必须登记安全元信息；Redis、HTTP Client、对象存储、任务队列、日志、配置、错误响应都优先使用 `github.com/teamsillybees/initra/pkg/*`。

## 先做判断

只有在当前项目满足至少一个条件时使用本 skill：

- `go.mod` 依赖 `github.com/teamsillybees/initra`。
- 代码位于 initra 生成的 API 项目结构中，例如 `internal/boot`、`internal/modules`、`configs`、`db/migrations`。
- 用户明确要求按 initra 框架、initra CLI、initra pkg 最佳实践开发。

## 工作协议

1. 先读当前项目的 `AGENTS.md`、`go.mod`、`internal/boot` 和目标模块。
2. 判断需求属于哪项框架能力，再读取对应 `references/*.md`。不要一次性读取所有 reference。
3. 修改代码时保持边界：boot 装配框架能力；module 编写业务逻辑；handler 做 HTTP 适配；service 直接承载业务逻辑和 Ent 操作。
4. 如果 initra 已经提供能力，业务项目不得新增平行基础设施。
5. 如果 initra 没有提供能力，先说明缺口，再添加项目内最小实现。
6. 修改 Go 文件后运行 `gofmt`，并执行与风险匹配的测试。

## 架构不变量

### 模块结构

标准业务模块保持 flat package：

```text
internal/modules/<module>/
  <module>.handler.go
  <module>.dto.go
  <module>.service.go
  <module>.routes.go
  providers.go
  cache.go
  <module>_test.go
```

- `*.dto.go` 只放 HTTP 边界类型：非导出 `request`/`response`、查询 `Query`、请求体 `Body`、对外 JSON `VO`。
- `*.handler.go` 只做 path/query/body 适配、调用 service、包装成功响应。
- `*.service.go` 直接写业务逻辑和 Ent 查询；不要恢复独立 repository 层或 service DTO 层。
- 领域实体放在 `cache.go` 或 service 相关文件里；不要恢复废弃的 `*.model.go`。

### 依赖边界

- 框架 package 在 `internal/boot` 一步式注册，例如 `logx.Register(injector, cfg.Log)`、`httpclient.Register(injector, cfg.HTTPClient)`、`redisx.Register(injector, cfg.Redis)`。
- 模块 `providers.go` 只负责把模块内构造函数连起来，并把框架依赖传给 service。
- service 和 handler 不调用 `do.Invoke`、`do.MustInvoke`，不持有全局 injector。
- 跨模块调用由调用方定义小接口，不 import 对方具体实现。

### ID 与 Schema

- 业务 ID 统一使用 `idgen.ID`；REST path、service 入参、Ent 主键/外键、JSON VO 都不使用 `int64` 或 `string` 表达业务 ID。
- Ent schema 复用 `pkg/entx/mixin`，不要在业务项目复制本地 mixin。
- 对外 OpenAPI/JSON ID 是字符串；示例使用 `"1771234567890123456"`。

### 路由安全

- 所有 `/api/` 路由必须通过 `server.RouteRegistry` 登记 `RouteSecurity`。
- 公开接口显式使用 `AccessModePublic`。
- 只需登录态的接口使用 `AccessModeAuthenticated`。
- 后台管理、运营、审核、退款、风控、配置管理等接口使用 `AccessModePermission`，并让 `Resource`/`Action` 匹配 Casbin policy。

## 能力选择

| 需求 | 使用 | 读取 |
| --- | --- | --- |
| 业务 ID、Ent 主键/外键、OpenAPI ID、schema mixin | `pkg/idgen`、`pkg/entx/mixin` | `references/id.md` |
| 配置加载、默认值、环境变量、脱敏、校验 | `pkg/config` 和各 pkg `Config` | `references/config.md` |
| boot/module providers、DI、`Register` | `samber/do` + initra 注册入口 | `references/di.md` |
| JWT、密码、refresh token、Casbin、路由安全 | `pkg/auth`、`pkg/server` | `references/auth.md` |
| 错误码、业务错误、HTTP 错误响应 | `pkg/errors`、`pkg/response` | `references/errors.md` |
| 结构化日志、trace/request id、脱敏 | `pkg/logx`、`pkg/requestctx` | `references/logging.md` |
| Redis client、key、SCAN、短锁、Redis 缓存 | `pkg/redisx` | `references/redisx.md` |
| 多级业务缓存、详情缓存 | `pkg/cache` | `references/cache.md` |
| 下游 HTTP、第三方服务、webhook | `pkg/httpclient` | `references/httpclient.md` |
| 上传下载、对象存储、预签名、STS | `pkg/storage`、`pkg/storage/provider` | `references/storage.md` |
| 异步任务、延迟任务、worker、scheduler | `pkg/task`、`pkg/task/asynqadapter` | `references/task.md` |

## 禁止清单

默认不要在业务项目中添加这些东西：

- 新 Redis client、Redis cluster client、生产 `KEYS`。
- 临时 `http.Client{}`、`http.DefaultClient`、硬编码远程 base URL。
- 直接 import 云厂商 SDK 或 initra storage provider 具体实现。
- 业务代码直接 import `github.com/hibiken/asynq`。
- 把 Asynq `TaskID` 或 `Unique` 当作长期业务幂等。
- 自定义全局错误响应结构，或在 handler 手写错误 JSON。
- 重新实现 Viper 配置加载器、logger 初始化器、认证中间件。
- 在日志中输出密码、token、验证码、session value、Authorization、access key、带密码 DSN。
- import `github.com/teamsillybees/initra/internal/...`。

## 验证

业务项目初始化本 skill：

```powershell
initra skill       # 写入 .agents/skills/initra-framework
initra skill codex # 同上
initra skill cc    # 写入 .claude/skills/initra-framework
```

快速检查常见违规：

```powershell
go run .agents/skills/initra-framework/scripts/check_initra_usage.go --root .
```

常规验证按改动范围选择：

```powershell
go test ./...
go vet ./...
```

如果修改的是 initra 框架仓库模板或 CLI，还要按仓库 `AGENTS.md` 运行根模块和 `examples` 的对应测试。
