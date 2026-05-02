# initra

`initra` 是面向企业内部 Go Web API 的快速开发脚手架，项目介绍统一按三个部分理解：

1. **标准项目模板**：通过 `templates/basic` 和 `examples/basic` 提供可生成、可运行的 API 服务基础工程模板。
2. **可复用的 Go package**：通过根模块 `pkg/*` 沉淀 Web、配置、错误、日志、认证、数据访问、Redis、缓存、对象存储、HTTP Client、任务调度等通用能力，业务项目通过 `go.mod` 按需引入。
3. **工程化 CLI**：通过 `cmd/initra` 承载生成项目、业务模块、CRUD 样例、迁移文件、配置样例、接口骨架、测试骨架和代码生成命令。

## 目录

```text
cmd/initra          工程化 CLI 入口
pkg/                可复用 Go package
internal/           根仓库内部测试与辅助代码
templates/basic     标准项目模板，供 CLI 生成项目使用
examples/basic      标准项目模板的可运行示例，包含 auth/user
docs/               架构与工程规范文档
scripts/            根仓库测试、检查、构建入口
```

## 标准项目模板

`templates/basic` 是 CLI 默认生成模板，`examples/basic` 是该模板的可运行来源与验证样例，保留完整 auth/user 功能：

- 登录、refresh token、当前用户信息
- 用户 CRUD 与用户详情缓存
- Gin + Huma OpenAPI
- JWT + Casbin 授权
- Atlas schema/migration、go-jet 生成代码、seed 数据

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

生成完整示例项目：

```powershell
$framework = (Resolve-Path .).Path
$target = Join-Path $env:TEMP "demo-api"
go run ./cmd/initra new $target --module example.com/demo-api --replace $framework
```

发布版 CLI 会用自身构建版本写入生成项目 `go.mod`。开发版 CLI 必须传 `--framework-version` 或 `--replace`，避免生成不可复现的 `initra` 依赖。

工程化 CLI 的职责边界是生成和维护工程骨架，包括项目、业务模块、CRUD 样例、迁移文件、配置样例、接口骨架、测试骨架和代码生成命令；运行时能力应沉淀在 `pkg/*`，业务示例应沉淀在标准项目模板中。

## 本地开发

仓库使用 `go.work` 联调根模块和 `examples/basic`：

```powershell
go test ./pkg/... ./cmd/initra/... ./internal/... -count=1
go test ./examples/basic/... -count=1
go vet ./pkg/... ./cmd/initra/... ./internal/...
go vet ./examples/basic/...
```

## 依赖治理

- 面向标准项目模板：生成项目默认采用 Go Modules，不要求业务项目使用 `go.work`。
- 面向可复用 Go package：私有发布时业务项目通过 `GOPRIVATE` 配置私有 Git 域名。
- 面向本地联调：生成项目可用 `replace github.com/teamsillybees/initra => <本地路径>` 指向当前仓库。
- 面向脚手架仓库自身：根仓库用 `go.work` 组织根模块和 `examples/basic` 开发。
