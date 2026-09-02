# AGENTS.md

本文件指导 coding agents 在本业务项目中开展开发、修改、审查和验证工作。适用范围为项目根目录及其所有子目录；若用户在对话中给出更具体要求，以用户要求为准。

## 工作原则

- 先阅读上下文，再修改文件。优先查看当前目录可用的 README、本文件、相关模块代码、测试、配置、Ent schema 和迁移文件；在 `examples` 中同时查看仓库根 README。
- 保持改动聚焦，只处理用户要求的任务；不要顺手重构无关代码或格式化无关文件。
- 工作区可能存在用户未提交改动。不要回滚、覆盖或清理非本次任务产生的变更。
- 面向项目已有模式实现功能，优先复用现有 package、构造函数、测试 fake 和错误处理方式。
- 当前代码和可执行测试是事实源；README、AGENTS 或 skill 与代码冲突时，按代码修正文档，不保留旧分层或历史兼容描述。
- 修改 Go 文件后运行 `gofmt`；修改 Ent schema、配置、迁移或生成入口后同步检查实际 Go 代码是否仍可格式化、编译和测试。
- 文档、注释和面向开发者的说明优先使用中文。
- 导出的类型、函数和常量使用符合 Go 规范的中文注释；复杂内部逻辑、测试 fake 和关键断言补充能解释意图的注释。

## 项目定位

本项目是基于 initra 的 Go RESTful API 服务。当前包含 `auth`、`user`、`file`、`httpdemo`、`taskdemo` 示例模块，用来展示认证、用户管理、文件存储、HTTP Client 和任务队列等常见能力的组织方式；真实业务开发时应按业务域替换或裁剪示例模块，不要把示例接口当成业务契约。

项目依赖 `github.com/teamsillybees/initra/pkg/*` 提供 Web、配置、错误、日志、认证、数据库、Redis、缓存、文件存储、HTTP Client、任务队列等通用能力。业务代码不得 import `github.com/teamsillybees/initra/internal/*`，也不得复制 initra 根仓库的 `pkg/*` 到业务项目内。

## 常用命令

在项目根目录执行：

```powershell
go test ./... -count=1
go vet ./...
go build ./cmd/server
go generate ./internal/data
initra migrate diff <name> --env local --config-dir configs
initra migrate apply --env local
initra migrate hash
initra skill
initra skill --check
```

`initra skill` 会把 initra 框架使用说明安装或升级到 `.agents/skills/initra-framework`，供 Codex 在业务项目中识别框架边界和最佳实践；`--check` 只校验内容是否为内置最新版本，确需覆盖本地修改时使用 `--force`。

涉及配置、迁移、依赖装配、路由权限或示例模块替换时，应运行与改动风险匹配的 `go test`、`go vet`、构建或迁移验证。若无法运行测试，最终回复中必须说明原因和未验证风险。

## Go 开发规范

- 遵循 Go 标准风格：小接口、显式错误处理、依赖通过构造函数注入。
- 包名保持简短、清晰、全小写；一个业务模块一个 package。
- 不引入不必要的抽象。只有在减少真实重复、隔离复杂度或匹配现有模式时才新增抽象。
- 共享能力优先放入项目内清晰的公共 package 或复用 `github.com/teamsillybees/initra/pkg/*`，不要通过业务模块之间互相 import 具体实现来复用逻辑。
- 业务 ID 统一使用 `github.com/teamsillybees/initra/pkg/idgen.ID`。Ent 主键、外键、auth user ID、REST path params、service 入参和 JSON VO 不使用 `int64` 或 `string` 表达业务 ID；对外 JSON/OpenAPI ID 暴露为字符串。包级生成器没有固定默认节点，每个运行实例必须显式配置唯一的 0–1023 `idgen.node`。仅在对接雪花 ID 生成器、第三方库、手写 SQL 参数或极少数底层基础设施场景调用 `Int64()`。
- Ent schema 使用 `github.com/teamsillybees/initra/pkg/entx/fieldx` 的 ID、审计、软删除等字段助手，按需使用仅定义版本字段的 `fieldx.OptimisticLockVersion()` 和 `pkg/entx/indexx`；Ent Client 注册 `entx.AuditHook`、`entx.RejectDeleteHook`。不要复制本地 schema helper，也不要把版本字段助手描述成完整自动乐观锁。
- 禁止在业务模块中随意使用 `panic` 或吞掉错误；错误应向上返回并保留足够上下文。
- 业务专属错误码集中放在 `internal/modules/bizerrors` 或业务约定的错误包中，通过统一错误工厂创建。

## 模块结构

业务代码按业务模块组织为单一 flat package，不拆 controller/service/repository 子目录。模块文件结构遵循：

```text
internal/modules/<module>/
  <module>.handler.go HTTP 适配：解包请求、转换参数、调用 Service、包装响应
  <module>.dto.go     传输边界类型：request/response、Query、Body、Form、VO、私有 Payload
  <module>.service.go 用例逻辑与依赖编排；需要数据库时直接使用 Ent Client
  <module>.routes.go  路由注册 + Module 结构体
  providers.go        samber/do 依赖注入
  cache.go            可选，缓存适配器或少量领域实体
  *_test.go           单元测试
```

- `*.handler.go` 负责 HTTP 解包、上下文数据提取、参数转换、Service 调用和响应包装，不在 handler 文件定义 DTO。
- 需要数据库的 Service 直接依赖 `*ent.Client`；缓存、密码、HTTP Client、任务 Publisher 等可替换能力由调用方定义最小私有接口。
- `*.dto.go` 保存传输边界类型：Huma/Gin 包装类型用非导出的 `request`/`response` 后缀，查询参数用 `Query`，请求体用 `Body`，multipart 表单用 `Form`，对外 JSON 类型用 `VO`，私有外部 HTTP 载荷可用 `Payload`。
- `Response` 只表示 Huma/Gin 内部响应包装，不作为对外 JSON DTO 后缀；统一成功/错误响应等 JSON 结构使用 `VO` 命名。
- Service 根据用例接受 HTTP 边界 `Body`/`Query`、`idgen.ID`、基础参数或职责明确的专用输入，返回 `VO` 或明确命名的内部结果类型；不要为了分层机械增加转换 DTO。
- 默认不创建 `<module>.repo.go` 或 `<module>.model.go`。领域实体可放在 `cache.go`、`*.service.go` 或明确命名的同包文件中，但不要为了凑结构创建空文件。
- 模块之间禁止循环依赖。跨模块调用优先依赖调用方内部定义的小接口，避免依赖具体实现。

## 路由、安全与配置

- 所有 `/api/` 业务接口必须通过 `registry.Register` 登记 `RouteSecurity`，鉴权中间件默认 fail-closed。
- 公开接口必须显式设置 `AccessModePublic`。
- 登录即可访问的用户侧接口必须设置 `AccessModeAuthenticated`。
- 后台管理、运营操作、审核、退款、风控、配置管理等接口必须设置 `AccessModePermission`，并通过 `Permission` 直接登记稳定权限标识（例如 `system:user:read`）。权限策略唯一事实源是 `sys_role`、`sys_menu`、`sys_role_menu`，禁止重新引入静态 policy 文件。
- access JWT 只承载 `userId`、`sessionId`、`sessionVersion` 和标准声明，不保存角色、权限或租户快照；请求时必须从 Redis 缓存或数据库解析当前有效用户、会话版本、角色和超级管理员状态。全部退出、修改密码或禁用账号必须递增 `sys_user.session_version`。
- 登录必须通过 `auth.LoginGuard` 执行账号/IP 限流与连续失败锁定，凭证类拒绝统一响应并只在安全审计日志记录内部原因。生产 CORS 使用明确白名单，匿名 `/docs`、`/openapi`、`/schemas` 只允许 dev/local/test 开放。
- 业务配置结构放在 `internal/boot/config.go`，通过 `pkg/config` 泛型加载；配置应支持默认值、环境变量覆盖、启动校验和敏感配置脱敏打印。
- 运行环境统一使用环境变量 `APP_ENV` 表示；其他配置环境变量默认使用 `INITRA_` 前缀。
- 密码、Token、Secret、Access Key、Authorization、验证码、session value 和带密码的 DSN 禁止明文输出到日志。

## 框架包使用

- Redis 业务能力优先使用 `pkg/redisx` 的 client、Key Builder、缓存、Lua registry、SCAN+UNLINK 和 redislock 短锁；禁止生产使用 `KEYS`。
- 认证 token store 默认使用 Redis；只有 dev/local/test 环境显式设置 `auth.allow_memory_token_store: true` 才允许内存 store，其他环境必须 fail-closed。
- 缓存能力优先复用 `pkg/cache`，业务侧保留明确的 key 命名、TTL 和失效策略；不要把 Redis value、验证码、session value 或 token 写入日志。
- 文件与对象存储业务能力优先使用 `pkg/storage` 的统一接口和 `pkg/storage/provider` 工厂；业务模块不要直接依赖云厂商 SDK。
- HTTP Client 在启动层通过 `httpclient.Register(injector, cfg.HTTPClient)` 统一装配；业务 Service 依赖 `httpclient.Executor`，模块使用 `httpclient.Provide` 绑定命名客户端。trace/request ID 由请求上下文自动透传，不在业务调用中重复拼 Header。
- `Application` 根据 `task.worker.enabled` 和 `task.scheduler.enabled` 按需解析、启动和关闭 Worker/Scheduler；禁用时不解析对应 provider。启用 Worker 时，`server.shutdown_timeout` 不得小于 `task.worker.shutdown_timeout`。新增任务 handler 或周期任务时通过 `pkg/task` 在启动前完成注册，业务代码不得直接 import `github.com/hibiken/asynq`。
- 任务队列按 at-least-once 模型设计，不承诺 exactly-once；`biz_key` 是业务幂等键，外部副作用任务必须由业务侧保证幂等。

## 依赖装配

- 需要 DI 的框架能力通过各自 `Register(injector, cfg)` 或 `Register(injector, options)` 入口在启动时装配；observability 路由当前通过 `NewModule(...).Register(api, registry)` 注册。
- `internal/boot/providers.go` 优先调用各 `pkg/*` 的 `Register` 入口，例如 `logx.Register(injector, cfg.Log)`、`redisx.Register(injector, cfg.Redis)`、`storageprovider.Register(injector, cfg.Storage)`、`httpclient.Register(injector, cfg.HTTPClient)`、`asynqadapter.Register(injector, cfg.Task)`。
- 业务模块在自己的 `providers.go` 中注册 cache、service、handler、module 等本地对象。模块内部允许 `Handler -> *Service`、`Module -> *Handler` 的具体依赖；Service 对可替换的框架或外部能力定义最小私有接口，数据库访问按当前模式直接依赖 `*ent.Client`。
- 禁止在业务模块中直接使用 `do.Invoke` 或滥用全局 injector。`do` 的使用边界应集中在 `providers.go` 和 `internal/boot` 装配层。
- 禁止为 package 装配做两阶段传参，例如先 `do.ProvideValue(injector, cfg.HTTPClient)` 再 `httpclient.Register(injector)`；配置应作为 `Register` 参数显式传入。

## 数据库与测试

- `internal/data/schema` 是数据库结构主源；`db/migrations` 是版本化迁移历史；`db/seeds` 保存种子数据。
- 手动修改 Ent schema 后运行 `go generate ./internal/data`，再生成迁移。
- 禁止物理外键和数据库级联；Ent edge 只表达逻辑关系，迁移 diff 必须保持 `migrate.WithForeignKeys(false)`。关联写入必须在同一事务内校验父记录有效并使用 `FOR SHARE`，父记录软删除或影响关系有效性的状态更新使用 `FOR UPDATE`，删除逻辑必须显式处理全部有效子关系。逻辑外键列仍需由单列索引或最左前缀复合索引覆盖，禁止重复索引。
- 数据库集成测试优先使用 `go-sqlmock` 或项目已有测试辅助验证 SQL、事务和参数。
- 新增或修改业务模块必须在模块单元测试、`test/integration` 或 `test/e2e` 中提供与风险匹配的覆盖。
- 架构边界相关改动必须保留或补充测试，确保业务项目不 import `github.com/teamsillybees/initra/internal/*`。
- PostgreSQL 连接使用结构化 URL 安全编码凭据和数据库名，并配置应用名、连接超时、连接池上限、空闲回收时间和最长生命周期；只有 dev/local/test 环境允许弱 TLS 配置，其他环境必须使用 `verify-full`。迁移 diff 默认按 `--env` 和 `--config-dir` 读取业务数据库配置，也可通过 `--dev-url` 临时覆盖。
