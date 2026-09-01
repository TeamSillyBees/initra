# AGENTS.md

本文件指导 Codex 在本仓库中开展开发、修改、审查和验证工作。适用范围为仓库根目录及其所有子目录；若用户在对话中给出更具体的要求，以用户要求为准。

## 工作原则

- 先阅读上下文，再修改文件。优先查看 `README.md`、本文件、相关包代码、测试和模板文件。
- 保持改动聚焦，只处理用户要求的任务；不要顺手重构无关代码或格式化无关文件。
- 工作区可能存在用户未提交改动。不要回滚、覆盖或清理非本次任务产生的变更。
- 面向仓库已有模式实现功能，优先复用现有 package、构造函数、测试 fake 和错误处理方式。
- 当前代码和可执行测试是架构事实源；README、AGENTS 或 skill 与代码冲突时，按代码修正文档，不保留旧入口、旧分层或历史兼容描述。
- 修改 Go 文件后运行 `gofmt`；修改模板时同步检查生成后的实际 Go 代码是否仍可格式化和编译。
- 文档、注释和面向开发者的说明优先使用中文。
- 导出的类型、函数和常量使用符合 Go 规范的中文注释；复杂内部逻辑、测试 fake 和关键断言补充能解释意图的注释，避免机械重复代码含义。

## 项目定位

`initra` 是面向企业内部 Go 服务的快速开发脚手架。理解和描述本仓库时必须区分三类内容：

- **标准项目模板**：`templates/api` 提供包含 auth/user/file/httpdemo/taskdemo 示例模块、Ent schema、seed 和 Atlas migrations 的 RESTful API 服务模板；`examples` 是 API 模板的可运行验证样例。模板不保存 Ent 生成代码，`initra new` 在渲染后执行生成入口。
- **可复用 Go package**：根模块 `github.com/teamsillybees/initra` 的 `pkg/*`，沉淀 Web、配置、错误、日志、认证、数据库、Redis、缓存、文件与对象存储、HTTP Client、任务队列、任务调度等通用能力；
- **工程化 CLI**：`cmd/initra`，负责生成项目、模块、显式代码片段、聚合配置、迁移文件和 Codex skill。

重要边界：

- `examples` 是独立 Go module 的可运行 API 示例项目，属于标准项目模板。
- `templates/api` 是 CLI API 项目模板，内容应与 `examples` 保持同步。
- `cmd/initra` 只负责生成和维护工程骨架，不承载运行时业务能力。
- `internal/` 只服务脚手架仓库自身；标准项目模板、生成项目和外部业务项目不得 import 根仓库 `internal/`。

## 常用命令

在仓库根目录执行：

```powershell
# 根模块测试与静态检查
go test ./pkg/... ./cmd/initra/... ./internal/... -count=1
go vet ./pkg/... ./cmd/initra/... ./internal/...

# 示例项目测试与静态检查
go test ./examples/... -count=1
go vet ./examples/...

# 构建 CLI
go build -o $env:TEMP\initra.exe ./cmd/initra

# 生成示例项目进行验证
$target = Join-Path $env:TEMP "demo-api"
go run ./cmd/initra new $target --type api --module example.com/demo-api --replace (Resolve-Path .).Path
```

仓库使用 `go.work` 联调根模块和 `examples`。涉及模板生成、包边界或示例项目行为时，应同时验证根模块和 `examples`。

## Go 开发规范

- 遵循 Go 标准风格：小接口、显式错误处理、依赖通过构造函数注入。
- 包名保持简短、清晰、全小写；一个业务模块一个 package。
- 不引入不必要的抽象。只有在减少真实重复、隔离复杂度或匹配现有模式时才新增抽象。
- 共享能力放入 `pkg/*`，不要通过业务模块之间互相 import 来复用逻辑。
- 业务 ID 统一使用 `pkg/idgen.ID`。Ent 主键、外键、auth user ID、REST path params、service 入参和 JSON VO 不使用 `int64` 或 `string` 表达业务 ID；对外 JSON/OpenAPI ID 暴露为字符串。`pkg/idgen` 没有固定默认节点，每个运行实例必须显式配置唯一的 0–1023 Snowflake node。仅在对接雪花 ID 生成器、第三方库、手写 SQL 参数或极少数底层基础设施场景调用 `Int64()`。
- Ent schema 使用 `pkg/entx/fieldx` 的 `ID()`、`Audit()`、`SoftDelete()` 等字段助手，按需使用仅定义版本字段的 `fieldx.OptimisticLockVersion()` 和 `pkg/entx/indexx`；审计操作人和物理删除保护由 Ent Client 注册 `entx.AuditHook`、`entx.RejectDeleteHook`。不要复制本地 schema helper，也不要把版本字段助手描述成完整的自动乐观锁实现。
- 标准模板彻底禁止物理外键与数据库级联，Ent edge 只表达代码生成、查询和业务建模所需的逻辑关系，迁移 diff 必须保持 `migrate.WithForeignKeys(false)`。关联写入必须在同一事务内校验父记录仍有效，并通过 `FOR SHARE` 与父记录软删除/状态更新的 `FOR UPDATE` 协作消除并发竞态；删除父记录时必须显式拒绝、软删除或迁移全部有效子关系。逻辑外键列仍必须由单列索引或以该列为最左前缀的复合索引覆盖，禁止保留被复合索引完全覆盖的重复索引。
- `examples` 为兼容已发布的 Atlas 迁移校验保留旧外键迁移原文，并由后续迁移删除已有外键；`tools/sync_api_templates.go` 必须把该历史版本渲染为 no-op 并重新计算模板 `atlas.sum`，确保新生成项目从未创建物理外键。禁止直接改写已发布的 examples 迁移历史。
- Redis 业务能力优先使用 `pkg/redisx` 的 client、Key Builder、缓存、Lua registry、SCAN+UNLINK 和 redislock 短锁；禁止生产使用 `KEYS`，禁止记录密码、token、验证码、session value。
- 认证 token store 默认使用 Redis；内存 store 必须显式 opt-in 且仅允许用于 dev/local/test。其他任何环境关闭 Redis 或启用内存 store 时都必须启动失败。
- 文件与对象存储业务能力优先使用 `pkg/storage` 的统一接口和 `pkg/storage/provider` 工厂；业务模块不要直接依赖云厂商 SDK。
- 任务队列业务能力优先使用 `pkg/task` 的 `Publisher`、`Worker`、`Registry`、`Scheduler`、`Task` 和 `TaskMeta`；业务代码不得直接 import `github.com/hibiken/asynq`，Asynq 类型只允许出现在 `pkg/task/asynqadapter`。
- `pkg/task` 按 at-least-once 模型设计，不承诺 exactly-once；`biz_key` 是业务幂等键，不是 Asynq `TaskID`，也不是 Asynq `Unique`，外部副作用任务必须由业务侧保证幂等。
- 框架能力包（如 `logx`、`httpclient`、`redisx`、`cache`、`idgen`、`storage`、`auth`、`server`、`task/asynqadapter`）通过 `Register(injector, cfg)` 或 `Register(injector, options)` 风格入口在启动时装配；业务对象通过构造函数接收最小必要依赖，禁止在 service/handler 中直接使用 `do.Invoke` 或持有全局 injector。
- `internal/boot/providers.go` 应优先调用各 `pkg/*` 的 `Register` 入口，例如 `storageprovider.Register(injector, cfg.Storage)`、`httpclient.Register(injector, cfg.HTTPClient)`、`redisx.Register(injector, cfg.Redis)`、`asynqadapter.Register(injector, cfg.Task)`；只有项目自身能力（如 examples 的 `data.NewEntClientFromDB`）才保留本地 `do.Provide`。
- 禁止为 package 装配做两阶段传参，例如先 `do.ProvideValue(injector, cfg.HTTPClient)` 再 `httpclient.Register(injector)`；配置应作为 `Register` 参数显式传入，公共组件（如 `*logx.Logger`）可由 `Register` 内部从 injector 解析。
- HTTP Client 装配由 `httpclient.Register(injector, cfg.HTTPClient)` 统一完成；业务模块优先通过 `httpclient.ProvideConsumer` 将命名 client 注入最小接口，或显式解析 `httpclient.ClientName("service")`，不要手写 `Factory.Get` provider。
- 标准模板会根据 `task.worker.enabled` 和 `task.scheduler.enabled` 在 `Application` 生命周期中按需解析、启动和关闭 Worker/Scheduler；禁用时不解析对应 provider。启用 Worker 时，`server.shutdown_timeout` 不得小于 `task.worker.shutdown_timeout`。新增任务 handler 或周期任务时仍需在启动前完成 Registry/Scheduler 注册。
- 禁止在业务模块中随意使用 `panic` 或吞掉错误；错误应向上返回并保留足够上下文。

## 架构约束

### 标准项目模板

业务代码按业务模块组织为单一 flat package，不拆 controller/service/repository 子目录。模块主文件按职责命名，必要配套能力可用独立文件承载。模块文件结构应遵循：

```text
internal/modules/<module>/
  <module>.handler.go HTTP 适配：解包请求、转换参数、调用 Service、包装响应
  <module>.dto.go     传输边界类型：request/response、Query、Body、Form、VO、私有 Payload
  <module>.service.go 用例逻辑与依赖编排；需要数据库时直接使用 Ent Client
  <module>.routes.go  路由注册 + Module 结构体
  providers.go        samber/do 依赖注入
  cache.go            可选，缓存适配器或领域实体
  *_test.go           单元测试
```

- `*.handler.go` 负责 HTTP path/query/body 解包、调用 service 和包装响应。Service 根据用例直接接受 Body/Query、`idgen.ID`、基础参数或明确的专用输入，返回 VO 或明确命名的内部结果类型；不要为了分层机械增加转换 DTO。
- 需要数据库的 `*.service.go` 直接通过 Ent Client 操作数据库，不拆独立 Repository 层；缓存、密码、HTTP Client 和任务 Publisher 等可替换能力由模块定义最小接口。
- 默认不创建 `<module>.model.go` 或 `<module>.repo.go`；领域实体放在 `cache.go`、`*.service.go` 或职责明确的同包文件中。
- 不为简单用例机械创建 service 层 DTO；只有当输入不属于 HTTP 边界、需要跨入口复用或隔离复杂用例时，才定义职责明确的专用类型。

- `*.dto.go` 只放传输边界类型：Huma/Gin 包装用非导出的 `request`/`response` 后缀，查询参数用 `Query`，请求体用 `Body`，multipart 表单用 `Form`，对外 JSON 类型用 `VO`，私有外部 HTTP 载荷可用 `Payload`。
- `Response` 只表示 Huma/Gin 内部响应包装，不作为对外 JSON DTO 后缀；统一成功/错误响应等 JSON 结构使用 `VO` 命名。
- `*.handler.go` 可引用 DTO 类型，但不定义 DTO；它只负责 HTTP 适配、调用 service 和包装响应。
- 模块之间禁止循环依赖。
- 跨模块调用优先依赖调用方内部定义的小接口，避免依赖具体实现。
- 业务模块应保持独立，不互相 import 具体实现。
- 示例项目只能依赖根模块的 `pkg/*`，不能 import 根仓库 `internal/`。

### 模板同步

- 修改 `examples` 中的示例代码时，使用 `go run ./tools/sync_api_templates.go` 同步模板，并用 `go run ./tools/sync_api_templates.go --dry-run` 复查。当前 Windows 换行差异可能产生假漂移，需结合实际 diff 判断。

### 错误处理

- 可复用错误码由 `pkg/errors` 定义，如 `CodeBadRequest`、`CodeUnauthorized`、`CodeInternalError`。
- 业务专属错误集中放在 `internal/modules/bizerrors/`；仅该错误门面调用 `pkg/errors` 的 New/Wrap，其他业务模块调用 `bizerrors` 暴露的语义函数。
- 业务模块内部不要新增 sentinel error，例如直接 `errors.New`；统一使用业务错误门面。
- HTTP 响应错误映射应复用现有 mapper 和 response 机制，不要在 handler 中手写不一致的错误响应结构。

### 路由、安全与配置

- 所有 `/api/` 接口必须通过 `registry.Register` 登记 `RouteSecurity`，鉴权中间件默认 fail-closed。
- 公开接口，例如登录、注册、验证码和公开内容，必须显式设置 `AccessModePublic`。
- 登录即可访问的 ToC 接口必须设置 `AccessModeAuthenticated`，只做认证不做 RBAC 授权。
- 后台管理、运营操作、审核、退款、风控、配置管理等接口必须设置 `AccessModePermission`，并通过 `Permission` 直接登记稳定权限标识（例如 `system:user:read`）。权限策略唯一事实源是 `sys_role`、`sys_menu`、`sys_role_menu`，禁止重新引入静态 policy 文件。
- access JWT 只承载用户和会话身份，不保存角色或权限；请求时必须通过 `auth.IdentityResolver` 从共享缓存或数据库解析当前有效用户、角色和超级管理员状态。
- API 模板的 file 示例模块使用 `storage.provider: local` 展示上传、下载、元信息查询和删除；切换云厂商时只调整 `storage` 配置与 provider。
- 业务项目在自己的 `internal/boot/config.go` 定义配置结构，并通过 `pkg/config` 泛型加载。
- `pkg/config` 提供 `LoadInto`、环境覆盖和 `Sanitize`，不绑定业务配置结构。业务项目 boot config 组合 pkg 已有配置结构体（如 `storage.Config`、`redisx.Config`），在聚合 `Validate()` 中调用各配置自身的 `Validate()`，再校验业务字段。
- 配置规范使用结构体定义，必须支持默认值、环境变量覆盖配置文件、启动校验和敏感配置脱敏打印。
- 运行环境统一环境变量 `APP_ENV` 表示；其他配置环境变量默认使用 `INITRA_` 前缀。
- 密码、Token、Secret、Access Key、Authorization、带密码的 DSN 禁止明文输出到日志。
- PostgreSQL 连接使用结构化 URL 安全编码凭据和数据库名，并配置 `application_name`、连接超时、连接池上限、空闲回收时间和最长生命周期；只有 dev/local/test 环境允许弱 TLS 配置，其他任何环境必须使用 `verify-full` 同时校验证书和主机名。`migrate diff` 默认按 `--env` 和 `--config-dir` 读取业务数据库配置，也可通过 `--dev-url` 临时覆盖。

## 测试要求

- 窄改动至少运行相关包测试；影响共享 package、CLI、模板或生成项目时运行对应完整命令。
- 新增或修改业务模块必须在模块单元测试、`test/integration` 或 `test/e2e` 中提供与风险匹配的覆盖。
- 数据库集成测试优先使用 `go-sqlmock` 验证 SQL 生成和参数。
- 架构边界相关改动必须保留或补充测试，确保示例项目不 import 根仓库 `internal/`。
- 若无法运行测试，必须在最终回复中说明原因和未验证风险。

## Codex 工作流

1. 明确任务影响范围：根模块、`pkg/*`、`cmd/initra`、`examples`、`templates/api` 或文档。
2. 使用 `rg` / `rg --files` 搜索相关代码与测试，避免凭记忆修改。
3. 修改前确认工作区状态，保护用户已有改动。
4. 按最小可行范围编辑文件；手动编辑优先使用补丁方式。
5. 运行与改动风险匹配的 `go test`、`go vet`、构建或生成验证。
6. 当代码主线发生变化时，同步更新 README、AGENTS、模板和 skill 中受影响的事实，不保留旧入口或旧 schema 描述。
7. 最终回复简要说明改了什么、验证了什么，以及任何未完成或未验证事项。
