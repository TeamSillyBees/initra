# AGENTS.md

本文件用于指导 Codex 或其他代码代理在本仓库中工作。请优先遵循这里的说明；如果与用户的明确要求冲突，以用户要求为准，并在回复中说明取舍。

## 项目定位

`initra` 是一个面向中大型 Golang Web API 项目的脚手架模板。项目采用垂直切片架构，平台基础设施位于 `internal/platform`，业务模块位于 `internal/app/<module>`，应用组合根位于 `internal/boot`。

主要技术栈：

- HTTP：Gin + Huma
- SQL：go-jet/jet
- Migration：Atlas
- ID：snowflake
- Cache：jetcache-go + Redis
- DI：samber/do
- Log：zap
- Error：samber/oops + 统一错误码
- Auth：JWT + Casbin
- Config：Viper

## 工作原则

- 修改代码前先阅读相关模块现有实现和测试，不要凭猜测改结构。
- 保持脚手架简洁、可复制、可扩展；避免引入复杂抽象和过度封装。
- 代码注释使用中文，说明设计意图、边界条件和非显而易见的实现原因，避免机械复述代码。
- 新增功能或修复行为时优先补测试，再改实现。
- 不要回滚或覆盖用户已有改动，除非用户明确要求。
- 不要手工编辑 `internal/gen/jet` 下的生成代码；需要变更表结构时，先改 `db/schema`，再重新生成。

## 目录约定

- `cmd/server`：程序入口，只负责启动上下文、构建信息和应用运行。
- `configs`：环境配置、Casbin 模型和策略。
- `db/schema`：Atlas desired database schema，每个表一个 SQL 文件。
- `db/migrations`：Atlas 生成或维护的迁移文件。
- `db/seeds`：初始化种子数据。
- `internal/boot`：应用组合根，负责配置、DI、模块注册和 HTTP Server 构建。
- `internal/platform`：平台基础设施，不允许反向依赖业务模块。
- `internal/app/auth`：登录、refresh token、当前用户信息。
- `internal/app/user`：用户 CRUD、用户详情缓存、角色关系维护。
- `internal/gen/jet`：go-jet 生成代码，只能通过生成脚本更新。
- `test`：integration 和 e2e 测试。

## Go 代码规范

- 遵循 Go 标准项目风格：小接口、显式错误处理、依赖通过构造函数注入。
- 导出类型、函数、常量必须有中文注释；内部 helper 若包含业务规则或边界也应注释。
- 错误统一使用 `internal/platform/errors` 中的错误码和包装函数。
- 日志使用 zap，避免记录敏感信息，例如 JWT 字符串、密码、refresh token。
- 业务模块不要直接依赖 Gin、Redis、Viper 等平台细节；通过 domain 接口隔离。
- `internal/platform` 不得 import `internal/app/...`。
- 手动编辑文件时运行 `gofmt`。

## go-jet 使用约定

- 仓储层使用 go-jet statement 构造 SQL，不要拼接业务 SQL 字符串。
- Jet import 使用 dot import，让查询接近 SQL 语义，并避免写出 `postgres.AND()` 这类硬编码 dialect 前缀：

```go
import (
	. "github.com/go-jet/jet/v2/postgres"
	. "github.com/teamsillybees/initra/internal/gen/jet/table"
)
```

- 不要手工维护 `internal/gen/jet/table/*.go`；表结构变更后通过脚本重新生成。

## Atlas 与数据库约定

- `db/schema` 是 desired database schema 目录，不是 migration 目录。
- 每张表结构单独一个 SQL 文件。
- 需要生成迁移时使用 Atlas diff，不要把 desired schema 当作直接迁移文件执行。
- 系统表以历史 `V1.0.0__init_sys.sql` 为设计参考，但以当前 `db/schema` 为准。

常用命令：

```powershell
.\scripts\atlas.ps1 migrate diff <name> --env local
.\scripts\atlas.ps1 migrate apply --env local
.\scripts\jet.ps1
```

## JWT 与鉴权约定

- access token 保持 stateless，不要把登录后签发的 JWT 字符串写入缓存。
- 服务端主动吊销 access token 时，通过 Redis 黑名单保存 token jti，并设置为该 token 剩余 TTL。
- refresh token 需要缓存到 Redis，并设置合理 TTL；刷新时采用消费旧 token 并签发新 token 的轮转语义。
- 受保护接口必须验证 JWT 签名、issuer、token type、exp、iat 和黑名单。
- `/api/` 路由授权必须 fail-closed：除非显式 `Public=true`，否则缺少安全元信息应拒绝访问。
- Casbin resource/action 必须与模块 policy 常量和路由注册保持一致。

## DI 约定

- 使用 `samber/do` 注册依赖。
- 同类型依赖若在多个模块中出现，使用 `ProvideNamed` 避免服务名冲突。
- 模块注册入口位于 `internal/app/<module>/wire.go` 和 `module.go`。

## 测试与验证

提交或交付前至少运行：

```powershell
go test ./... -count=1
go vet ./...
go build -o $env:TEMP\initra-server-check.exe ./cmd/server
```

如修改了数据库 schema 或仓储 SQL：

```powershell
.\scripts\atlas.ps1 migrate diff <name> --env local
.\scripts\jet.ps1
go test ./test/integration ./internal/app/... -count=1
```

如修改了认证、授权、中间件或路由注册：

```powershell
go test ./internal/platform/auth ./internal/platform/web ./test/e2e -count=1
```

## 生成代码与脚手架代码

- `internal/gen/jet` 是生成代码，不要补手写注释或手工修复。
- 新业务模块优先通过 `scripts/new_module.ps1` 或 `make new-module NAME=xxx` 生成骨架，再按实际业务调整。
- 脚手架示例代码应保持清晰、完整、可运行，不能留下旧表结构、旧字段或死逻辑。

## 回复用户时

- 用中文简洁说明修改点、验证命令和结果。
- 如果某些验证无法运行，要明确说明原因。
- 如果发现与用户要求相关的风险，应直接指出并给出可执行建议。
