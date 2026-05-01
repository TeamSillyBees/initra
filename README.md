# initra

`initra` 是面向企业内部 Go Web API 项目的公共框架库与项目生成器。根模块发布稳定的 `pkg/*` 能力，业务项目通过 `go.mod` 引入；认证、用户管理等示例业务放在 `examples/basic` 和 CLI 模板中。

## 目录

```text
cmd/initra          项目生成 CLI
pkg/                稳定公共 package
internal/           根仓库内部测试与辅助代码
templates/basic     CLI 使用的完整项目模板
examples/basic      可运行示例项目，包含 auth/user
docs/               架构与工程规范文档
scripts/            根仓库测试、检查、构建入口
```

## 公共能力

- `pkg/config`：泛型配置加载，不绑定业务项目配置结构。
- `pkg/logging`、`pkg/database`、`pkg/cache`、`pkg/idgen`、`pkg/password`：基础设施薄封装。
- `pkg/errors`、`pkg/response`、`pkg/requestctx`：统一错误、响应、trace/request id。
- `pkg/auth`：JWT、refresh token、Redis token store、Casbin、路由安全元信息。
- `pkg/web`：Gin + Huma 应用与认证授权中间件装配。
- `pkg/observability`：health、ready、version 接口模块。

业务项目应只 import 实际需要的 `pkg/*`，组合根由业务项目自己的 `internal/boot` 显式组装。

## CLI

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

发布版 CLI 会用自身构建版本写入生成项目 `go.mod`。开发版 CLI 必须传 `--framework-version` 或 `--replace`，避免生成不可复现的框架依赖。

## 本地开发

仓库使用 `go.work` 联调根模块和 `examples/basic`：

```powershell
go test ./pkg/... ./cmd/initra/... ./internal/... -count=1
go test ./examples/basic/... -count=1
go vet ./pkg/... ./cmd/initra/... ./internal/...
go vet ./examples/basic/...
```


## 示例项目

`examples/basic` 是 CLI 默认模板的来源，保留完整 auth/user 功能：

- 登录、refresh token、当前用户信息
- 用户 CRUD 与用户详情缓存
- Gin + Huma OpenAPI
- JWT + Casbin 授权
- Atlas schema/migration、go-jet 生成代码、seed 数据

示例项目是独立 Go module，只通过 `replace github.com/teamsillybees/initra => ../..` 引入根公共库，不依赖根仓库 `internal`。

## 依赖治理

- 默认采用 Go Modules。
- 私有发布时业务项目通过 `GOPRIVATE` 配置私有 Git 域名。
- 本地联调用 `replace github.com/teamsillybees/initra => <本地路径>`。
- 根仓库用 `go.work` 组织开发，不要求业务项目使用 `go.work`。
