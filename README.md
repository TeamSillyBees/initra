# initra

`initra` 是面向企业内部 Go 服务的快速开发脚手架，项目介绍统一按三个部分理解：

1. **标准项目模板**：通过 `templates/api` 提供包含 auth/user、Ent schema 与生成代码、seed 和 Atlas migrations 的 RESTful API 服务模板，通过 `templates/worker` 提供后台 worker 占位骨架；`examples/api` 是 API 模板的可运行验证样例。
2. **可复用的 Go package**：通过根模块 `pkg/*` 沉淀 Web、配置、错误、日志、认证、数据访问、Redis、缓存、对象存储、HTTP Client、任务调度等通用能力，业务项目通过 `go.mod` 按需引入。
3. **工程化 CLI**：通过 `cmd/initra` 承载生成项目、业务模块、CRUD 样例、迁移文件、配置样例、接口骨架、测试骨架和代码生成命令。

## 目录

```text
cmd/initra          工程化 CLI 入口
pkg/                可复用 Go package
internal/           根仓库内部测试与辅助代码
templates/api       RESTful API 项目模板
templates/worker    worker 项目占位模板
examples/api        API 模板的可运行示例
docs/               架构与工程规范文档
scripts/            根仓库测试、检查、构建入口
```

## 项目模板

`api` 模板生成可直接运行的 Web API 基础项目，默认包含 Gin + Huma、统一响应/错误、JWT/Casbin 中间件装配、配置加载、日志、health/ready/version 接口，以及 auth/user 基础业务模块、Ent schema 与生成代码、seed 数据和 Atlas 配置。API 标准模板使用 Ent 作为类型安全持久化访问层，使用 Ent Mixin + Runtime Hook 实现雪花 ID、审计字段和软删除等通用自动填充能力。业务方仍可通过后续 CLI 命令追加新的模块、CRUD 样例和配置能力。

`worker` 模板面向后台任务、定时任务、消费任务、批处理任务。目前只提供可编译的占位入口，后续 worker 所需框架能力成熟后再扩展。

模板生成的业务项目是独立 Go module，通过 `go.mod` 引入 `github.com/teamsillybees/initra` 的可复用 Go package，不复制根仓库 `pkg/` 源码，也不依赖根仓库 `internal/`。

## 模型命名约定

标准 API 项目按模块保持 flat package。HTTP 边界类型放在 `*.dto.go`：Huma/Gin 包装类型使用非导出的 `request`/`response` 后缀；查询参数使用 `Query`；请求体使用 `Body`；对外 JSON DTO 使用 `VO`；分页 JSON 输出使用 `pagination.PageVO[T]` 泛型。领域实体和 service/repo 入参放在 `*.model.go`，结构体入参统一使用 `DTO` 后缀；禁止使用 `Result` 后缀命名返回值，列表直接使用 `[]T`，分页使用 `pagination.PageResult[T]` 泛型封装。

## 可复用 Go package

- `pkg/config`：泛型配置加载，不绑定业务项目配置结构。
- `pkg/logging`、`pkg/cache`、`pkg/idgen`：基础设施封装。
- `pkg/redisx`：Redis 基础能力封装，支持 standalone/sentinel client、Ping/readiness、Key Builder、JSON/Msgpack 缓存、TTL jitter、空值缓存、singleflight、SCAN+UNLINK、Lua script registry、基于 `github.com/bsm/redislock` 的短时间分布式锁，以及 OpenTelemetry/zap hook；不支持 cluster，不封装 KEYS。
- `pkg/entx`：Ent 通用 Hook 和上下文工具，不依赖具体项目生成的 `internal/ent`。
- `pkg/errors`、`pkg/response`、`pkg/requestctx`：统一错误、响应、trace/request id。
- `pkg/auth`：JWT、refresh token、Redis token store、Casbin、路由安全元信息。
- `pkg/server`：Gin + Huma 应用与认证授权中间件装配。
- `pkg/observability`：health、ready、version 接口模块。

业务项目应只 import 实际需要的 `pkg/*`，组合根由业务项目自己的 `internal/boot` 显式组装。

## 配置规范

- 业务项目使用结构体定义配置，配置结构放在自己的 `internal/boot/config.go`。
- 配置加载支持默认值、YAML 配置文件、环境变量覆盖和启动校验。
- 运行环境统一使用 `app.env` 或无前缀环境变量 `APP_ENV` 表示；其他配置环境变量默认使用 `INITRA_` 前缀。

## 工程化 CLI

构建 CLI：

```powershell
go build -o $env:TEMP\initra.exe ./cmd/initra
```

生成项目：

```powershell
$framework = (Resolve-Path .).Path
go run ./cmd/initra new $env:TEMP\demo-api --type api --module example.com/demo-api --replace $framework
go run ./cmd/initra new $env:TEMP\demo-worker --type worker --module example.com/demo-worker --replace $framework
```

规划中的核心命令：

```powershell
initra new <app> --type api
initra new <app> --type worker
initra module add <name>
initra crud add <module> --table <table>
initra config add <capability>
initra migrate new <name>
initra migrate diff <name>
initra doctor
```

发布版 CLI 会用自身构建版本写入生成项目 `go.mod`。开发版 CLI 必须传 `--framework-version` 或 `--replace`，避免生成不可复现的 `initra` 依赖。

## 本地开发

仓库使用 `go.work` 联调根模块和 `examples/api`：

```powershell
go test ./pkg/... ./cmd/initra/... ./internal/... -count=1
go test ./examples/api/... -count=1
go vet ./pkg/... ./cmd/initra/... ./internal/...
go vet ./examples/api/...
```

## 依赖治理

- 面向标准项目模板：生成项目默认采用 Go Modules，不要求业务项目使用 `go.work`。
- 面向可复用 Go package：私有发布时业务项目通过 `GOPRIVATE` 配置私有 Git 域名。
- 面向本地联调：生成项目可用 `replace github.com/teamsillybees/initra => <本地路径>` 指向当前仓库。
- 面向脚手架仓库自身：根仓库用 `go.work` 组织根模块和 `examples/api` 开发。
