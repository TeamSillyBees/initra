# initra

`initra` 是面向企业内部 Go 服务的快速开发脚手架，项目介绍统一按三个部分理解：

1. **标准项目模板**：通过 `templates/api` 提供 RESTful API 服务骨架，通过 `templates/worker` 提供后台 worker 占位骨架；`examples/api` 是 API 模板的可运行验证样例。
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

`api` 模板只生成 Web API 骨架，不强行绑定数据库、Redis、业务模块或 CRUD 样例。默认包含 Gin + Huma、统一响应/错误、JWT/Casbin 中间件装配、配置加载、日志和 health/ready/version 接口。

`worker` 模板面向后台任务、定时任务、消费任务、批处理任务。目前只提供可编译的占位入口，后续 worker 所需框架能力成熟后再扩展。

模板生成的业务项目是独立 Go module，通过 `go.mod` 引入 `github.com/teamsillybees/initra` 的可复用 Go package，不复制根仓库 `pkg/` 源码，也不依赖根仓库 `internal/`。

## 可复用 Go package

- `pkg/config`：泛型配置加载，不绑定业务项目配置结构。
- `pkg/logging`、`pkg/db`、`pkg/cache`、`pkg/idgen`：基础设施封装。
- `pkg/errors`、`pkg/response`、`pkg/requestctx`：统一错误、响应、trace/request id。
- `pkg/auth`：JWT、refresh token、Redis token store、Casbin、路由安全元信息。
- `pkg/server`：Gin + Huma 应用与认证授权中间件装配。
- `pkg/observability`：health、ready、version 接口模块。

业务项目应只 import 实际需要的 `pkg/*`，组合根由业务项目自己的 `internal/boot` 显式组装。

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
