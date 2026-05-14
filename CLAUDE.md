# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目定位

initra 是面向企业内部 Go 服务的快速开发脚手架。所有项目介绍和架构说明必须明确归属到以下三个部分，避免把模板、运行时 package 和 CLI 生成器混为一谈：

- **标准项目模板**：`templates/api` 提供包含 auth/user/file 示例模块、Ent schema 与生成代码、seed 和 Atlas migrations 的 RESTful API 服务模板，`templates/worker` 提供后台 worker 占位骨架；`examples/api` 是 API 模板的可运行验证样例。
- **可复用的 Go package**：根模块 `github.com/teamsillybees/initra` 的 `pkg/*`，沉淀 Web、配置、错误、日志、认证、数据访问、Redis、缓存、文件与对象存储、HTTP Client、任务调度等通用能力。
- **工程化 CLI**：`cmd/initra`，承载生成项目、业务模块、CRUD 样例、迁移文件、配置样例、接口骨架、测试骨架和代码生成命令。

跨部分边界：

- `examples/api` 是独立 Go module 的可运行 API 示例项目，属于标准项目模板。
- `templates/api` 是 CLI API 项目模板，内容与 `examples/api` 保持同步，并保留 auth/user 基础模块、file 本地文件示例模块、Ent schema 与生成代码、seed 数据和 Atlas migrations，属于标准项目模板。
- `templates/worker` 是 CLI worker 项目模板，目前只提供可编译占位骨架。
- `cmd/initra` 只负责生成和维护工程骨架，属于工程化 CLI。
- `internal/` 只服务脚手架仓库自身，标准项目模板、生成项目和外部业务项目都不得 import。

主要技术栈：

- HTTP：Gin + Huma
- 持久化：Ent 类型安全持久化访问 + Atlas migrations
- Migration：Atlas
- ID：snowflake
- Redis：`pkg/redisx` + go-redis/v9，支持 standalone/sentinel，不支持 cluster
- Cache：jetcache-go + Redis；直接使用 Redis 时优先组合 `pkg/redisx`
- DI：samber/do
- Log：zap
- Error：samber/oops + 统一错误码
- Auth：JWT + Casbin
- Config：Viper
- Storage：`pkg/storage` 统一接口 + local/阿里云 OSS/腾讯云 COS/AWS S3/S3 兼容 provider

## 常用命令

```powershell
# 根模块测试
go test ./pkg/... ./cmd/initra/... ./internal/... -count=1
go vet ./pkg/... ./cmd/initra/... ./internal/...

# 示例项目测试
go test ./examples/api/... -count=1
go vet ./examples/api/...

# 构建 CLI
go build -o $env:TEMP\initra.exe ./cmd/initra

# 生成示例项目
go run ./cmd/initra new $target --type api --module example.com/demo-api --replace (Resolve-Path .).Path
```

仓库使用 `go.work` 联调根模块和 `examples/api`，在根目录执行 Go 命令即可覆盖两个模块。

## Go 代码规范

- 遵循 Go 标准项目风格：小接口、显式错误处理、依赖通过构造函数注入。
- 导出类型、函数、常量必须有中文注释；内部 helper 若包含业务规则或边界也应注释。
- 注释完善，清晰表达意图，禁止使用英文注释或无意义的注释，注释符合 Golang 注释规范。
- 手动编辑文件时运行 `gofmt`。

## 架构约束

### 模块组织

本节面向标准项目模板和由模板生成的业务项目。

业务代码按业务模块组织为**单一 flat package**，不拆 controller/service/repository 子目录。模块主文件按职责命名，必要配套能力可用独立文件承载：

```
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
- 禁止使用 `Result` 后缀命名 service/repo 层返回值，列表直接使用 `[]T`，分页统一使用 `pagination.PageResult[T]` 泛型封装。
- `*.dto.go` 只放 HTTP 边界类型：Huma/Gin 包装类型用非导出的 `request`/`response` 后缀；HTTP 查询参数用 `Query` 后缀；HTTP 请求体用 `Body` 后缀；对外 JSON DTO 用 `VO` 后缀。分页 JSON 输出使用 `pagination.PageVO[T]` 泛型，不定义模块专属分页 VO。
- `Response` 只表示 Huma/Gin 内部响应包装，不作为对外 JSON DTO 后缀；统一成功/错误响应等 JSON 结构使用 `VO` 命名。
- `*.handler.go` 可引用 DTO 类型，但不再定义 DTO；它只负责参数转换、调用 service 和包装响应。

### 依赖规则

- 面向标准项目模板：模块之间禁止循环依赖，不互相 import 具体实现。
- 面向标准项目模板：跨模块调用优先依赖接口，不依赖具体实现；接口定义在调用方模块内部，保持小接口、少方法。
- 面向可复用 Go package：共享能力放入 `pkg/*`，禁止业务模块通过互相引用来解决复用问题。
- 面向标准项目模板：示例项目只能依赖可复用 Go package `pkg/*`，不能 import 根仓库 `internal/`。
- 面向标准项目模板和可复用 Go package：包名简短、清晰、全小写，一个模块一个 package。
- 面向 Redis 能力：统一优先使用 `pkg/redisx` 的 client、Key Builder、缓存、Lua registry、SCAN+UNLINK 和 redislock 短锁；禁止生产使用 `KEYS`，禁止记录密码、token、验证码、session value。
- 面向文件与对象存储能力：统一优先使用 `pkg/storage` 和 `pkg/storage/provider`，业务模块只依赖统一接口，不直接依赖云厂商 SDK。
- 面向依赖注入：框架能力包（如 `logging`、`httpclient`、`redisx`、`cache`、`idgen`、`storage`、`auth`、`server`）应提供 `Register(injector, cfg)` 或 `Register(injector, options)` 风格入口函数，在启动时显式调用一次完成装配；业务代码只依赖接口、不感知底层实现，禁止在业务模块中直接使用 `do.Invoke` 或滥用全局 injector。
- 面向启动装配：`internal/boot/providers.go` 应优先调用各 `pkg/*` 的 `Register` 入口，例如 `storageprovider.Register(injector, cfg.Storage)`、`httpclient.Register(injector, cfg.HTTPClient)`、`redisx.Register(injector, cfg.Redis)`；只有项目自身能力（如 examples 的 `data.NewEntClient`）才保留本地 `do.Provide`。
- 面向启动装配：禁止为 package 装配做两阶段传参，例如先 `do.ProvideValue(injector, cfg.HTTPClient)` 再 `httpclient.Register(injector)`；配置应作为 `Register` 参数显式传入，依赖的公共组件（如 `*zap.Logger`）可由 `Register` 内部从 injector 解析。
- 面向 HTTP Client：`httpclient.Register(injector, cfg.HTTPClient)` 统一注册 Factory 和各服务命名 Client，业务模块通过 `httpclient.ClientName("service")` 注入命名 `*httpclient.Client` 或定义最小接口，避免在业务模块中手写 `Factory.Get` provider。

### 错误处理

- 面向可复用 Go package：统一错误码由 `pkg/errors` 定义（`CodeBadRequest`、`CodeUnauthorized`、`CodeInternalError` 等）。
- 面向标准项目模板：业务专属错误码可在 `internal/module/bizerrors/` 中定义，使用 `apperrors.New` 工厂函数创建。
- 面向标准项目模板：禁止在业务模块内部创建 sentinel error（`errors.New`），统一走 bizerrors。

### 路由与安全

- 面向标准项目模板：所有 `/api/` 接口必须通过 `registry.Register` 登记 `RouteSecurity`，否则鉴权中间件按 **fail-closed** 拒绝请求。
- 面向标准项目模板：公开接口设置 `Public: true`。
- 面向标准项目模板：`RouteSecurity` 的 Resource/Action 需与 Casbin policy 文件中定义的一致。
- 面向标准项目模板：file 示例模块默认使用 `storage.provider: local`，提供上传、下载、元信息查询和删除；切换云厂商应通过 `storage` 配置完成。

### 配置

- 面向标准项目模板：业务项目在 `internal/boot/config.go` 中定义自己的配置结构，使用 `pkg/config` 泛型加载。
- 面向可复用 Go package：`pkg/config` 只提供泛型加载能力，不绑定具体业务配置结构；pkg 中的配置结构体应复用 `pkg/config` 的 `Sanitize`、`Validate` 公共方法，避免各自重复实现脱敏与校验；业务项目 boot config 应组合 pkg 中定义的配置结构体（如 `storage.Config`、`redisx.Config`），而非从头定义。
- 面向工程化 CLI 和标准项目模板：模板文件中的模块路径使用 `{{ .ModulePath }}`，禁止硬编码 `github.com/teamsillybees/initra/examples/api`。
- 配置规范使用结构体定义，必须支持默认值、环境变量覆盖配置文件、启动校验和敏感配置脱敏打印。
- 运行环境统一使用 `app.env` 或无前缀环境变量 `APP_ENV` 表示；其他配置环境变量默认使用 `INITRA_` 前缀。
- 密码、Token、Secret、Access Key、Authorization、带密码的 DSN 禁止明文输出到日志。

### 测试

- 面向标准项目模板：业务模块应包含单元测试文件，按复杂度使用 fake 实现测试 service 编排逻辑。
- 面向标准项目模板：涉及 SQL 的集成测试优先使用 `go-sqlmock` 验证 SQL 生成正确性。
- 面向标准项目模板和仓库边界：架构测试确保示例项目不 import 根仓库 `internal/`。
