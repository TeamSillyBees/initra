# AGENTS.md

本文件指导 Codex 在本仓库中开展开发、修改、审查和验证工作。适用范围为仓库根目录及其所有子目录；若用户在对话中给出更具体的要求，以用户要求为准。

## 工作原则

- 先阅读上下文，再修改文件。优先查看 `README.md`、本文件、相关包代码、测试和模板文件。
- 保持改动聚焦，只处理用户要求的任务；不要顺手重构无关代码或格式化无关文件。
- 工作区可能存在用户未提交改动。不要回滚、覆盖或清理非本次任务产生的变更。
- 面向仓库已有模式实现功能，优先复用现有 package、构造函数、测试 fake 和错误处理方式。
- 修改 Go 文件后运行 `gofmt`；修改模板时同步检查生成后的实际 Go 代码是否仍可格式化和编译。
- 文档、注释和面向开发者的说明优先使用中文。
- 代码注释完善，类型、函数、常量必须有符合 Go 规范的中文注释。符合 Go 代码风格。测试代码同样也要提供注释方便理解。

## 项目定位

`initra` 是面向企业内部 Go 服务的快速开发脚手架。理解和描述本仓库时必须区分三类内容：

- **标准项目模板**：`templates/api` 提供包含 auth/user/file 示例模块、Ent schema 与生成代码、seed 和 Atlas migrations 的 RESTful API 服务模板；`examples` 是 API 模板的可运行验证样例。
- **可复用 Go package**：根模块 `github.com/teamsillybees/initra` 的 `pkg/*`，沉淀 Web、配置、错误、日志、认证、数据库、Redis、缓存、文件与对象存储、HTTP Client、任务队列、任务调度等通用能力；其中 Redis 基础能力统一放在 `pkg/redisx`，支持 standalone/sentinel，不支持 cluster；任务队列能力统一放在 `pkg/task`，默认底层适配 Asynq。
- **工程化 CLI**：`cmd/initra`，负责生成项目、业务模块、CRUD 样例、迁移文件、配置样例、接口骨架、测试骨架和代码生成命令。

重要边界：

- `examples` 是独立 Go module 的可运行 API 示例项目，属于标准项目模板。
- `templates/api` 是 CLI API 项目模板，内容应与 `examples` 保持同步，并保留 auth/user 基础模块、file 本地文件示例模块、Ent schema 与生成代码、seed 数据和 Atlas migrations。
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
- Redis 业务能力优先使用 `pkg/redisx` 的 client、Key Builder、缓存、Lua registry、SCAN+UNLINK 和 redislock 短锁；禁止生产使用 `KEYS`，禁止记录密码、token、验证码、session value。
- 文件与对象存储业务能力优先使用 `pkg/storage` 的统一接口和 `pkg/storage/provider` 工厂；业务模块不要直接依赖云厂商 SDK。
- 任务队列业务能力优先使用 `pkg/task` 的 `Publisher`、`Worker`、`Registry`、`Scheduler`、`Task` 和 `TaskMeta`；业务代码不得直接 import `github.com/hibiken/asynq`，Asynq 类型只允许出现在 `pkg/task/asynqadapter`。
- `pkg/task` 按 at-least-once 模型设计，不承诺 exactly-once；`biz_key` 是业务幂等键，不是 Asynq `TaskID`，也不是 Asynq `Unique`，外部副作用任务必须由业务侧保证幂等。
- 框架能力包（如 `logging`、`httpclient`、`redisx`、`cache`、`idgen`、`storage`、`auth`、`server`、`task/asynqadapter`）应提供 `Register(injector, cfg)` 或 `Register(injector, options)` 风格入口函数，在启动时显式调用一次完成装配；业务代码只依赖接口、不感知底层实现，禁止在业务模块中直接使用 `do.Invoke` 或滥用全局 injector。
- `internal/boot/providers.go` 应优先调用各 `pkg/*` 的 `Register` 入口，例如 `storageprovider.Register(injector, cfg.Storage)`、`httpclient.Register(injector, cfg.HTTPClient)`、`redisx.Register(injector, cfg.Redis)`、`asynqadapter.Register(injector, cfg.Task)`；只有项目自身能力（如 examples 的 `data.NewEntClient`）才保留本地 `do.Provide`。
- 禁止为 package 装配做两阶段传参，例如先 `do.ProvideValue(injector, cfg.HTTPClient)` 再 `httpclient.Register(injector)`；配置应作为 `Register` 参数显式传入，依赖的公共组件（如 `*zap.Logger`）可由 `Register` 内部从 injector 解析。
- HTTP Client 装配由 `httpclient.Register(injector, cfg.HTTPClient)` 统一完成，业务模块通过 `httpclient.ClientName("service")` 注入命名 `*httpclient.Client` 或定义最小接口，不要在业务模块中手写 `Factory.Get` provider。
- 禁止在业务模块中随意使用 `panic` 或吞掉错误；错误应向上返回并保留足够上下文。

## 架构约束

### 标准项目模板

业务代码按业务模块组织为单一 flat package，不拆 controller/service/repository 子目录。模块主文件按职责命名，必要配套能力可用独立文件承载。模块文件结构应遵循：

```text
internal/module/<module>/
  <module>.handler.go Handler + HTTP 适配方法 + DTO/领域模型转换
  <module>.dto.go     HTTP 边界类型：内部 Request/Response、Query、Body、VO
  <module>.service.go 业务逻辑 + 私有接口定义
  <module>.repo.go    数据库实现
  <module>.model.go   领域实体 + service/repo DTO 与结果类型
  <module>.routes.go  路由注册 + Module 结构体
  providers.go        samber/do 依赖注入
  cache.go            可选，缓存适配器
  <module>_test.go    单元测试
```

- `*.model.go` 只放模块内部稳定模型，不放 `json`、`path`、`query` 等传输层 tag；service/repo 层的结构体入参统一使用 `DTO` 后缀，不再使用 `Params`。
- `*.dto.go` 只放 HTTP 边界类型：Huma/Gin 包装类型用非导出的 `request`/`response` 后缀；HTTP 查询参数用 `Query` 后缀；HTTP 请求体用 `Body` 后缀；对外 JSON DTO 用 `VO` 后缀。
- `Response` 只表示 Huma/Gin 内部响应包装，不作为对外 JSON DTO 后缀；统一成功/错误响应等 JSON 结构使用 `VO` 命名。
- `*.handler.go` 可引用 DTO 类型，但不再定义 DTO；它只负责参数转换、调用 service 和包装响应。
- 模块之间禁止循环依赖。
- 跨模块调用优先依赖调用方内部定义的小接口，避免依赖具体实现。
- 业务模块应保持独立，不互相 import 具体实现。
- 示例项目只能依赖根模块的 `pkg/*`，不能 import 根仓库 `internal/`。

### 模板同步

- 修改 `examples` 中的示例代码时，使用 `tools/sync_api_templates.go` 进行模板同步

### 错误处理

- 可复用错误码由 `pkg/errors` 定义，如 `CodeBadRequest`、`CodeUnauthorized`、`CodeInternalError`。
- 业务专属错误码可放在 `internal/module/bizerrors/`，通过 `apperrors.New` 工厂函数创建。
- 业务模块内部不要新增 sentinel error，例如直接 `errors.New`；统一使用业务错误定义。
- HTTP 响应错误映射应复用现有 mapper 和 response 机制，不要在 handler 中手写不一致的错误响应结构。

### 路由、安全与配置

- 所有 `/api/` 接口必须通过 `registry.Register` 登记 `RouteSecurity`，鉴权中间件默认 fail-closed。
- 公开接口，例如登录、注册、验证码和公开内容，必须显式设置 `AccessModePublic`。
- 登录即可访问的 ToC 接口必须设置 `AccessModeAuthenticated`，只做认证不做 RBAC 授权。
- 后台管理、运营操作、审核、退款、风控、配置管理等接口必须设置 `AccessModePermission`，其 `Resource`、`Action` 必须与 Casbin policy 文件保持一致。
- API 模板的 file 示例模块使用 `storage.provider: local` 展示上传、下载、元信息查询和删除；切换云厂商时只调整 `storage` 配置与 provider。
- 业务项目在自己的 `internal/boot/config.go` 定义配置结构，并通过 `pkg/config` 泛型加载。
- `pkg/config` 只提供通用加载能力，不绑定任何具体业务配置结构；pkg 中的配置结构体应复用 `pkg/config` 的 `Sanitize`、`Validate` 公共方法，避免各自重复实现脱敏与校验；业务项目 boot config 应组合 pkg 中定义的配置结构体（如 `storage.Config`、`redisx.Config`），而非从头定义。
- 配置规范使用结构体定义，必须支持默认值、环境变量覆盖配置文件、启动校验和敏感配置脱敏打印。
- 运行环境统一环境变量 `APP_ENV` 表示；其他配置环境变量默认使用 `INITRA_` 前缀。
- 密码、Token、Secret、Access Key、Authorization、带密码的 DSN 禁止明文输出到日志。

## 测试要求

- 窄改动至少运行相关包测试；影响共享 package、CLI、模板或生成项目时运行对应完整命令。
- 标准项目模板每个业务模块应包含单元测试，使用 fake 实现测试 service 编排逻辑。
- 数据库集成测试优先使用 `go-sqlmock` 验证 SQL 生成和参数。
- 架构边界相关改动必须保留或补充测试，确保示例项目不 import 根仓库 `internal/`。
- 若无法运行测试，必须在最终回复中说明原因和未验证风险。

## Codex 工作流

1. 明确任务影响范围：根模块、`pkg/*`、`cmd/initra`、`examples`、`templates/api` 或文档。
2. 使用 `rg` / `rg --files` 搜索相关代码与测试，避免凭记忆修改。
3. 修改前确认工作区状态，保护用户已有改动。
4. 按最小可行范围编辑文件；手动编辑优先使用补丁方式。
5. 运行与改动风险匹配的 `go test`、`go vet`、构建或生成验证。
6. 当代码主线发生变化时，同步更新对应模块 README，避免保留已废弃入口或旧 schema 描述。
7. 最终回复简要说明改了什么、验证了什么，以及任何未完成或未验证事项。
