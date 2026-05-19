---
name: initra-framework
description: 当 Codex 在基于 initra 的 Go 业务项目中开发、修改、审查或验证代码时使用，尤其适用于 Redis、缓存、分布式锁、HTTP Client、文件与对象存储、任务队列、统一错误与响应、日志、配置、认证、JWT、Casbin、路由安全、依赖注入、boot providers、业务模块结构或 Agent 操作流程相关任务。该 skill 指导 Codex 在业务模块新增基础设施代码前，优先复用 initra 框架能力。
---

# initra-framework

## 核心规则

将 initra 视为业务项目的框架能力源。编写 Redis、缓存、HTTP Client、存储、任务队列、错误处理、日志、配置、认证、路由安全或依赖注入代码前，先确认 `github.com/teamsillybees/initra/pkg/*` 是否已经提供对应能力。

每个任务先读取 `assets/capabilities.yaml`，再只加载匹配的 `references/` 文件。

## Agent 操作流程

1. 确认项目引入了 `github.com/teamsillybees/initra`，并查看 `AGENTS.md`、`go.mod`、`internal/boot` 和目标业务模块。
2. 使用下方能力路由表，将需求映射到一个或多个 initra 能力。
3. 修改代码前读取对应的 `references/*.md`。
4. 按现有 boot/provider 与 flat module 模式实现。
5. 业务 service 只依赖模块内小接口；框架组件只在 boot/provider 层初始化。
6. 修改 Go 文件后运行 `gofmt`，并执行匹配风险范围的 `go test` / `go vet`。
7. 如果 initra 未提供该能力，明确说明缺口，再添加最小的项目内抽象。

## 能力路由

- Redis、Key、TTL、验证码、token/session 存储、分布式锁、Redis 缓存：`references/redisx.md`
- 远程 HTTP API、第三方服务、webhook、内部服务调用：`references/httpclient.md`
- 上传、下载、对象存储、本地存储、OSS、COS、S3、预签名 URL、STS：`references/storage.md`
- 异步任务、延迟任务、指定时间任务、Worker、任务处理器、周期任务、biz_key：`references/task.md`
- 本地/远端业务缓存、用户/资料/详情缓存：`references/cache.md`
- 错误码、业务错误、HTTP 错误响应、错误映射：`references/errors.md`
- Logger、结构化日志、脱敏、trace/request id 日志：`references/logging.md`
- 配置结构、默认值、环境变量覆盖、校验、安全日志输出：`references/config.md`
- JWT、refresh token、密码管理、Casbin、路由安全：`references/auth.md`
- `samber/do`、`Register`、boot providers、module providers、命名依赖：`references/di.md`

## 业务编码规则

使用标准 flat module 结构：

```text
internal/module/<module>/
  <module>.handler.go
  <module>.dto.go
  <module>.service.go
  <module>.repo.go
  <module>.model.go
  <module>.routes.go
  providers.go
  cache.go
  <module>_test.go
```

- HTTP 边界类型放在 `*.dto.go`：Huma/Gin 包装类型使用非导出的 `request` / `response`；查询参数使用 `Query`；请求体使用 `Body`；对外 JSON DTO 使用 `VO`。
- 领域模型和 service/repo 入参放在 `*.model.go`；service/repo 入参结构体使用 `DTO`，不要使用 `Params`。
- Handler 只做传输层适配：转换请求数据、调用 service、包装响应。
- Service 不做框架初始化。通过构造函数注入模块内小接口，并返回 `apperrors.AppError`。
- 每个 `/api/` 路由都必须注册到 `server.RouteRegistry`；公开路由必须设置 `RouteSecurity{Public: true}`。
- `Resource` 和 `Action` 必须与 Casbin policy 保持一致。
- 生成项目和业务项目不得 import `github.com/teamsillybees/initra/internal/...`。

## 装配规则

- 框架 package 使用一步式注册，例如 `logging.Register(injector, cfg.Log)` 或 `httpclient.Register(injector, cfg.HTTPClient)`。
- 任务队列使用 `asynqadapter.Register(injector, cfg.Task)` 注册，业务模块只注入 `task.Publisher` 或在 worker 侧注册 `task.Registry` handler。
- 配置作为参数直接传入 `Register`；不要先 `do.ProvideValue` 暂存配置，再调用无参注册函数。
- Boot 代码负责应用级 providers。模块 `providers.go` 负责模块内构造函数和命名依赖。
- 业务 service 和 handler 不应调用 `do.Invoke` 或 `do.MustInvoke`。
- 优先依赖 package 接口（如 `storage.Service`、`httpclient.Factory`）或模块内小接口，不依赖具体云厂商 SDK。

## 默认禁止

除非 reference 明确允许，否则不要在业务模块中添加：

- 新 Redis client 或 Redis cluster client。
- 生产环境 `KEYS` 扫描。
- 临时 `http.Client{}` 或 `http.DefaultClient` 远程服务调用。
- 直接使用云厂商 SDK 的存储代码。
- 业务代码直接 import `github.com/hibiken/asynq`。
- 把 Asynq `TaskID` 或 `Unique` 当作长期业务幂等机制。
- 自定义全局错误响应结构。
- 重复实现基于 Viper 的配置加载器。
- 自定义 logger 初始化。
- 在 handler 或 service 方法内初始化基础设施。
- 输出密码、token、验证码、session value、Authorization header、access key 或带密码 DSN 的日志。

## 验证

业务项目可先执行 `initra skill init` 写入本 skill 文档，再使用内置检查脚本做快速静态检查：

```powershell
go run .agents/skills/initra-framework/scripts/check_initra_usage.go --root .
```

涉及 initra 模板或生成项目改动时，还要运行项目匹配的 `go test`、`go vet`、CLI 构建和生成项目验证命令。

## 资源

- `assets/capabilities.yaml`：机器可读的能力索引。
- `assets/forbidden-patterns.yaml`：静态检查禁用模式。
- `assets/version.json`：skill 与框架来源版本元数据。
- `references/*.md`：每项能力的详细使用规则。
- `examples/*.go`：boot provider、Redis 用法、HTTP Client adapter、任务队列发布的可改写示例。
- `scripts/check_initra_usage.go`：无额外依赖的常见违规静态扫描脚本。
