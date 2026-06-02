# initra

`initra` 是面向企业内部 Go 服务的快速开发脚手架，项目介绍统一按三个部分理解：

1. **标准项目模板**：通过 `templates/api` 提供包含 auth/user/file 示例模块、Ent schema、seed 和 Atlas migrations 的 RESTful API 服务模板；`examples` 是 API 模板的可运行验证样例。API 模板不内置 Ent 生成代码，`initra new` 会在生成项目后自动执行 Ent 代码生成。
2. **可复用的 Go package**：通过根模块 `pkg/*` 沉淀 Web、配置、错误、日志、认证、数据访问、Redis、缓存、文件与对象存储、HTTP Client、任务队列、任务调度等通用能力，业务项目通过 `go.mod` 按需引入。
3. **工程化 CLI**：通过 `cmd/initra` 承载生成项目、业务模块、CRUD 样例、迁移文件、配置样例、接口骨架、测试骨架和代码生成命令。

## 技术栈

- Web 服务：Gin + Huma
- 数据库与 ORM：PostgreSQL + Ent
- 数据库迁移：Atlas
- 认证与授权：JWT + Casbin
- 配置与依赖注入：Viper + samber/do
- 错误处理：基于 oops 统一错误链、业务错误码与 HTTP 响应映射
- 日志与观测：logx/zap + health/ready/version
- 业务 ID：`pkg/idgen.ID` + snowflake，REST/OpenAPI 对外使用 JSON string
- Redis 客户端：go-redis
- 二级缓存：jetcache-go
- 任务队列与调度：Asynq
- 文件与对象存储：local / OSS / COS / S3
- 工程化 CLI：Cobra

## 目录

```text
cmd/initra          工程化 CLI 入口
pkg/                可复用 Go package
internal/           根仓库内部测试与辅助代码
templates/api       RESTful API 项目模板
examples            API 模板的可运行示例
docs/               架构与工程规范文档
```

## 项目模板

`api` 模板生成可直接运行的 Web API 基础项目，默认包含 Gin + Huma、统一响应/错误、JWT/Casbin 中间件装配、配置加载、日志、health/ready/version 接口，以及 auth/user 基础业务模块、file 本地文件示例模块、Ent schema、seed 数据和 Atlas 配置。API 标准模板使用 Ent 作为类型安全持久化访问层，通过 `pkg/entx/mixin` + Runtime Hook 实现雪花 ID、审计字段和软删除等通用自动填充能力；模板目录只保存 schema、`generate.go` 和迁移 diff 入口，`initra new` 会在渲染 API 项目后执行 `go run ./internal/data/entgenerate` 生成缺失的 Ent 代码。业务方仍可通过后续 CLI 命令追加新的模块、CRUD 样例和配置能力。

模板生成的业务项目是独立 Go module，通过 `go.mod` 引入 `github.com/teamsillybees/initra` 的可复用 Go package，不复制根仓库 `pkg/` 源码，也不依赖根仓库 `internal/`。

## 模型命名约定

标准 API 项目按模块保持 flat package。HTTP 边界类型放在 `*.dto.go`：Huma/Gin 包装类型使用非导出的 `request`/`response` 后缀；查询参数使用 `Query`；请求体使用 `Body`；对外 JSON DTO 使用 `VO`；分页 JSON 输出使用 `pagination.PageVO[T]` 泛型。业务 ID 在 Ent、service/repo、auth、REST path params 和 JSON VO 中统一使用 `idgen.ID`，OpenAPI 暴露为 `type: string`，JSON 示例形如 `"1771234567890123456"`。领域实体和 service/repo 入参放在 `*.model.go`，结构体入参统一使用 `DTO` 后缀；禁止使用 `Result` 后缀命名返回值，列表直接使用 `[]T`，分页使用 `pagination.PageResult[T]` 泛型封装。

## 可复用 Go package

- `pkg/config`：泛型配置加载，不绑定业务项目配置结构。
- `pkg/logx`、`pkg/cache`、`pkg/idgen`、`pkg/database`：基础设施封装；`pkg/logx` 统一封装 zap 初始化、console 终端可读输出、JSONL 文件日志、oops 错误字段提取和脱敏策略。其中 `pkg/idgen.ID` 是业务 ID 专用类型，底层为 int64、JSON/OpenAPI 为 string；`pkg/database` 提供 SQL 连接池注册与启动 Ping 检查。
- `pkg/redisx`：Redis 基础能力封装，支持 standalone/sentinel client、Ping/readiness、Key Builder、JSON/Msgpack 缓存、TTL jitter、空值缓存、singleflight、SCAN+UNLINK、Lua script registry、基于 `github.com/bsm/redislock` 的短时间分布式锁，以及 OpenTelemetry/zap hook；不支持 cluster，不封装 KEYS。
- `pkg/entx`：Ent 通用 Hook、上下文工具和可复用 schema mixin，不依赖具体项目生成的 `internal/data/ent`。
- `pkg/errors`、`pkg/response`、`pkg/requestctx`：统一错误、响应、trace/request id。
- `pkg/auth`：JWT、refresh token、Redis token store、Casbin、路由安全元信息。
- `pkg/server`：Gin + Huma 应用与认证授权中间件装配。
- `pkg/observability`：health、ready、version 接口模块。
- `pkg/storage`：统一文件与对象存储接口，支持 local、阿里云 OSS、腾讯云 COS、AWS S3 和 S3 兼容服务，覆盖上传、下载、删除、预签名、临时授权和分片上传。
- `pkg/task`：任务队列抽象层，提供 `Publisher`、`Worker`、`Registry`、`Scheduler`、`Task` 与 `TaskMeta`；默认由 `pkg/task/asynqadapter` 适配 Asynq，业务代码不直接依赖 Asynq。任务队列按 at-least-once 模型设计，`biz_key` 是业务幂等键，外部副作用任务必须由业务侧保证幂等。

业务项目应只 import 实际需要的 `pkg/*`，组合根由业务项目自己的 `internal/boot` 显式组装。

## 配置规范

- 业务项目使用结构体定义配置，配置结构放在自己的 `internal/boot/config.go`。
- 配置加载支持默认值、可选 `configs/config.yaml` 初始值、可选 `configs/config.<env>.yaml` 覆盖、环境变量覆盖和启动校验；配置文件不存在时会跳过，环境变量可独立完成配置。
- 运行环境由显式 `--env`、无前缀环境变量 `APP_ENV` 或默认值 `dev` 决定，不再从 YAML 的 `app.env` 读取；其他配置环境变量默认使用 `INITRA_` 前缀。
- API 模板默认启用 `storage.provider: local`，文件落到 `./var/uploads`，可通过 `storage` 配置分组切换云厂商 provider。

## 工程化 CLI

构建 CLI：

```powershell
go build -o $env:TEMP\initra.exe ./cmd/initra
```

安装正式版 CLI：

```powershell
go install github.com/teamsillybees/initra/cmd/initra@latest
```

生成项目：

```powershell
$framework = (Resolve-Path .).Path
go run ./cmd/initra new $env:TEMP\demo-api --type api --module example.com/demo-api --replace $framework
```

`initra new` 创建 API 项目后，会在初始化 Git 仓库前执行一次 `go run ./internal/data/entgenerate`，把模板中未保存的 Ent 生成代码补齐，然后执行 `git init`。

正式版 CLI 会自动使用自身版本写入生成项目 `go.mod`：

```powershell
initra new $env:TEMP\demo-api --type api --module example.com/demo-api
```

核心命令：

```powershell
initra
initra help [command]
initra new <app> --type api
initra module add <name>
initra crud add <module> --table <table>
initra config add <capability>
initra migrate new <name>
initra migrate diff <name>
initra migrate apply --env <env>
initra migrate hash
initra skill codex
initra skill cc
initra doctor
```

直接执行 `initra` 或 `initra <group>` 会展示对应帮助；`initra help <command>` 可查看任意子命令的参数、示例和说明。

`initra migrate diff <name>` 直接执行当前项目的 `go run ./internal/data/migratediff/main.go <name>`，可追加 `--env`、`--config-dir` 和 `--dev-url`。`initra migrate apply --env <env>` 会执行 `atlas -c file://db/atlas.hcl migrate apply --env <env>` 应用迁移。手动修改迁移文件后，可执行 `initra migrate hash` 重新计算 `atlas.sum`。

在业务项目根目录执行 `initra skill codex` 会写入 `.agents/skills/initra-framework`，供 Codex 使用；执行 `initra skill cc` 会写入 `.claude/skills/initra-framework`，供 Claude Code 使用。

发布版 CLI 会用自身构建版本写入生成项目 `go.mod`。开发版 CLI 必须传 `--framework-version` 或 `--replace`，避免生成不可复现的 `initra` 依赖。

## 本地开发

仓库使用 `go.work` 联调根模块和 `examples`：

```powershell
go test ./pkg/... ./cmd/initra/... ./internal/... -count=1
go test ./examples/... -count=1
go vet ./pkg/... ./cmd/initra/... ./internal/...
go vet ./examples/...
```

## 依赖治理

- 面向标准项目模板：生成项目默认采用 Go Modules，不要求业务项目使用 `go.work`。
- 面向可复用 Go package：私有发布时业务项目通过 `GOPRIVATE` 配置私有 Git 域名。
- 面向本地联调：生成项目可用 `replace github.com/teamsillybees/initra => <本地路径>` 指向当前仓库。
- 面向脚手架仓库自身：根仓库用 `go.work` 组织根模块和 `examples` 开发。
