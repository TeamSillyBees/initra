# 架构说明

`initra` 已拆分为“公共框架库 + CLI 模板 + 示例项目”。

## 边界

- 根模块 `github.com/teamsillybees/initra` 发布稳定 `pkg/*` API。
- 根仓库 `internal/*` 只服务框架仓库自身，业务项目不得 import。
- `examples/basic` 是独立业务项目示例，包含认证和用户管理。
- `templates/basic` 是 CLI 默认模板，内容由 `examples/basic` 同步生成。

## 公共 package 分层

- 基础层：`pkg/errors`、`pkg/requestctx`、`pkg/response` 不依赖 Web 装配层。
- 基础设施层：`pkg/config`、`pkg/logging`、`pkg/database`、`pkg/cache`、`pkg/idgen`、`pkg/password` 按需引入。
- Web 与安全层：`pkg/auth`、`pkg/web`、`pkg/observability` 只在业务项目需要 HTTP/JWT/Casbin 时引入。

框架不提供公共 `boot` 全家桶。业务项目在自己的 `internal/boot` 里显式创建配置、日志、数据库、Redis、JWT、Casbin、Web App，并注册业务模块。

## 示例项目结构

```text
examples/basic/cmd/server          示例服务入口
examples/basic/internal/boot       示例组合根
examples/basic/internal/app/auth   登录、refresh、me
examples/basic/internal/app/user   用户 CRUD 与缓存
examples/basic/internal/gen/jet    示例数据库生成代码
examples/basic/configs             示例配置与 Casbin 策略
examples/basic/db                  Atlas schema、migration、seed
examples/basic/test                架构、集成、e2e 测试
```

示例业务代码仍采用垂直切片，不强制 controller/service/repository 横向分层。

## 生成器

`cmd/initra` 使用 `templates/basic` 生成业务项目。生成后的项目：

- 只复制业务项目源码、配置、数据库文件、脚本和测试。
- 不复制根仓库 `pkg/` 源码。
- 在 `go.mod` 中 require `github.com/teamsillybees/initra`。
- 本地开发时可写入 `replace` 指向框架仓库。

## 标准验证

根模块：

```powershell
go test ./pkg/... ./cmd/initra/... ./internal/... -count=1
go vet ./pkg/... ./cmd/initra/... ./internal/...
go build -o $env:TEMP\initra.exe ./cmd/initra
```

示例项目：

```powershell
go test ./examples/basic/... -count=1
go vet ./examples/basic/...
go build -o $env:TEMP\initra-basic-server.exe ./examples/basic/cmd/server
```
