# initra

`initra` 是面向企业内部 Go 服务的快速开发脚手架，项目介绍统一按三个部分理解：

1. **标准项目模板**：通过 `templates/api` 提供包含 auth/user/file/httpdemo/taskdemo 示例模块、Ent schema、seed 和 Atlas migrations 的 RESTful API 服务模板；`examples` 是 API 模板的可运行验证样例。API 模板不内置 Ent 生成代码，`initra new` 会在生成项目后自动执行 Ent 代码生成。
2. **可复用的 Go package**：通过根模块 `pkg/*` 沉淀 Web、配置、错误、日志、认证、数据访问、Redis、缓存、文件与对象存储、HTTP Client、任务队列、任务调度等通用能力，业务项目通过 `go.mod` 按需引入。
3. **工程化 CLI**：通过 `cmd/initra` 承载项目、模块、显式代码片段、聚合配置、迁移和 Codex skill 的生成命令。

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

`api` 模板生成 Go Web API 基础项目，当前包含 Gin + Huma、统一响应/错误、JWT/Casbin、配置、日志、health/ready/version，以及 `auth`、`user`、`rbac`、`file`、`httpdemo`、`taskdemo` 示例模块、Ent schema、seed 和 Atlas migrations。RBAC 以数据库中的角色、权限资源和授权关系为唯一事实源；access JWT 只保存用户 ID、会话 ID 和用户级会话版本，请求时从共享缓存或数据库解析当前身份。标准认证接口包含登录、刷新、当前用户、当前会话退出、全部会话退出和修改密码，并为登录提供 Redis 账号/IP 限流与连续失败锁定。Schema 使用 `pkg/entx/fieldx` 组合 ID、审计、软删除等字段，按需使用 `pkg/entx/indexx`；Ent edge 仅表达逻辑关系，物理外键生成被关闭，关联写入通过事务内有效性检查和行锁保持一致。运行时在 Ent Client 注册 `entx.AuditHook` 和 `entx.RejectDeleteHook`。模板不保存 Ent 生成代码，`initra new` 渲染后会执行 `go run ./internal/data/entgenerate`。

`examples` 保留已发布的旧外键迁移原文以兼容存量数据库升级，最新迁移会删除这些约束；模板同步时会把旧版本渲染为 no-op 并重算 `atlas.sum`，因此新生成项目的完整迁移链也不会创建物理外键。

模板生成的业务项目是独立 Go module，通过 `go.mod` 引入 `github.com/teamsillybees/initra` 的可复用 Go package，不复制根仓库 `pkg/` 源码，也不依赖根仓库 `internal/`。

## 模型命名约定

标准 API 项目按业务模块保持 flat package，不拆 controller/service/repository 子目录。HTTP 边界类型放在 `*.dto.go`：Huma/Gin 包装类型使用非导出的 `request`/`response` 后缀，查询参数使用 `Query`，请求体使用 `Body`，对外 JSON 类型使用 `VO`，分页输出使用 `pagination.PageVO[T]`。Handler 负责 HTTP 解包和响应包装；Service 直接操作 Ent，并根据用例接受 Body/Query、`idgen.ID`、基础参数或专用输入，返回 VO 或明确命名的内部结果类型。业务 ID 在 Ent、auth、REST path、service 和 JSON VO 中统一使用 `idgen.ID`，对外 JSON/OpenAPI 表示为字符串。

## 可复用 Go package

- `pkg/config`：泛型配置加载，不绑定业务项目配置结构。
- `pkg/logx`、`pkg/cache`、`pkg/idgen`、`pkg/database`：基础设施封装；`pkg/logx` 统一封装 zap 初始化、console 终端可读输出、JSONL stdout/文件日志、按日期与大小滚动、oops 错误字段提取和脱敏策略。其中 `pkg/idgen.ID` 是业务 ID 专用类型，底层为 int64、JSON/OpenAPI 为 string；包级 ID 生成器没有默认节点，应用必须显式注册实例唯一的 Snowflake node。`pkg/database` 提供带连接数、空闲回收和最长生命周期配置的 SQL 连接池注册与启动 Ping 检查。
- `pkg/redisx`：Redis 基础能力封装，支持 standalone/sentinel client、Ping/readiness、Key Builder、JSON/Msgpack 缓存、TTL jitter、空值缓存、singleflight、SCAN+UNLINK、Lua script registry、基于 `github.com/bsm/redislock` 的短时间分布式锁，以及 OpenTelemetry/zap hook；不支持 cluster，不封装 KEYS。
- `pkg/entx`：Ent 通用 Hook、上下文工具，以及 `fieldx`/`indexx` schema 字段和索引助手；不依赖具体项目生成的 `internal/data/ent`。
- `pkg/errors`、`pkg/response`、`pkg/requestctx`：统一错误、响应、trace/request id。
- `pkg/auth`：带会话 ID/版本的 JWT、refresh token、Redis token store、access 黑名单、账号/IP 登录限流与失败锁定、数据库 Casbin adapter、请求身份解析契约和路由权限元信息；内存状态实现仅允许 dev/local/test 显式 opt-in。
- `pkg/server`：Gin + Huma 应用与认证授权中间件装配、配置化 CORS、环境化文档暴露，以及由 `RouteSecurity` 自动生成的 OpenAPI Bearer/JWT 安全契约。
- `pkg/observability`：health、ready、version 接口模块；`/health` 只检查进程存活，`/ready` 通过带独立超时的 registry 检查必要依赖。
- `pkg/storage`：统一文件与对象存储接口，支持 local、阿里云 OSS、腾讯云 COS、AWS S3 和 S3 兼容服务；分片上传与临时授权通过可选扩展接口提供，具体支持范围由 provider 决定，local provider 不支持 presign。
- `pkg/task`：任务队列抽象层，提供 `Publisher`、`Worker`、`Registry`、`Scheduler`、`Task` 与 `TaskMeta`；默认由 `pkg/task/asynqadapter` 适配 Asynq，业务代码不直接依赖 Asynq。任务队列按 at-least-once 模型设计，`biz_key` 是业务幂等键，外部副作用任务必须由业务侧保证幂等。启用 Worker 时，应用总关闭超时必须不小于 Worker 关闭超时。

业务项目应只 import 实际需要的 `pkg/*`，组合根由业务项目自己的 `internal/boot` 显式组装。

## 配置规范

- 业务项目使用结构体定义配置，配置结构放在自己的 `internal/boot/config.go`。
- 配置加载支持默认值、可选 `configs/config.yaml` 初始值、可选 `configs/config.<env>.yaml` 覆盖、环境变量覆盖和启动校验；配置文件不存在时会跳过，环境变量可独立完成配置，配置文件中的未知字段会导致启动失败。
- 服务进程通过无前缀环境变量 `APP_ENV` 选择运行环境，未设置时使用 `dev`；迁移 diff 默认通过 `--env` 和 `--config-dir` 读取对应业务数据库配置，也可用 `--dev-url` 覆盖，apply/hash 通过 `--env` 选择 Atlas 环境。其他配置环境变量默认使用 `INITRA_` 前缀。
- 标准模板默认启用 Redis 作为共享 token store；只有 dev/local/test 环境显式设置 `auth.allow_memory_token_store: true` 才允许关闭 Redis，其他环境一律 fail-closed。`idgen.node` 没有可用于生产的默认值，每个实例必须配置唯一的 0–1023 节点号。
- `auth.login_protection` 默认启用账号/IP 登录速率限制和连续失败锁定；`server.cors` 使用明确白名单，`server.docs` 默认仅在 dev/local/test 开放，其他环境必须关闭匿名文档路由。
- 数据库连接使用结构化 PostgreSQL URL 安全编码凭据并配置连接超时与连接池生命周期；只有 dev/local/test 环境允许弱 TLS 配置，其他环境必须使用 `verify-full`。
- API 模板的基础配置启用 `storage.provider: local`，默认路径统一为 `./var/uploads`；可通过 `storage` 配置分组切换云厂商 provider。

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

`initra new` 在目标同级临时目录依次完成依赖下载、Ent 生成和全项目测试，再初始化 Git 并原子移动到目标目录；失败不会留下半成品。`--app-name` 会派生独立的展示名称和安全 `app.slug`，每次生成还会创建随机 JWT secret 与一次性管理员密码，密码只在项目成功落盘后输出一次。

核心命令：

```powershell
initra
initra help [command]
initra new <dir> --type api
initra module add <name>
initra snippet add <module> --table <table>
initra config add <capability>
initra migrate new <name>
initra migrate diff <name> --env <env> --config-dir configs
initra migrate apply --env <env>
initra migrate hash
initra skill [--check|--force]
initra skill codex [--check|--force]
initra doctor [--json]
```

`module add` 生成符合当前 flat-package 标准的模块骨架，但仍需在项目 boot 中完成模块注册；`snippet add` 只生成显式命名的数据表常量，不承诺 CRUD、持久化或路由接线；`config add` 会事务化更新能力配置类型、聚合 `boot.Config` 和主配置 YAML。

直接执行 `initra` 或 `initra <group>` 会展示对应帮助；`initra help <command>` 可查看任意子命令的参数、示例和说明。

`initra migrate diff` 默认从 `--env` 对应的业务配置构造数据库 URL，使生成的 diff 基于该业务库；需要临时覆盖时可传 `--dev-url`。`initra migrate apply --env <env>` 会执行 `atlas -c file://db/atlas.hcl migrate apply --env <env>` 应用迁移。手动修改迁移文件后，可执行 `initra migrate hash` 重新计算 `atlas.sum`。

在业务项目根目录执行 `initra skill` 或 `initra skill codex`，会写入 Codex 使用的 `.agents/skills/initra-framework`；`--check` 只校验版本和内容，`--force` 用于覆盖已修改的内置文件。`initra doctor --json` 可输出适合 agent 消费的稳定检查报告，必需工具或配置异常时返回非零状态。

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

## 工程治理

- 安全问题按 [SECURITY.md](SECURITY.md) 私密报告。
- 开发和提交要求见 [CONTRIBUTING.md](CONTRIBUTING.md)，面向使用者的变更记录见 [CHANGELOG.md](CHANGELOG.md)。
- LICENSE/内部授权、CODEOWNERS 与二进制发布信任链仍需由仓库所有者明确决定，当前不作推断。
