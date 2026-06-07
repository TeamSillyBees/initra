# AGENTS.md

本文件指导 coding agents 在本业务项目中开展开发、修改、审查和验证工作。适用范围为项目根目录及其所有子目录；若用户在对话中给出更具体要求，以用户要求为准。

## 工作原则

- 先阅读上下文，再修改文件。优先查看 `README.md`、本文件、相关模块代码、测试、配置、Ent schema 和迁移文件。
- 保持改动聚焦，只处理用户要求的任务；不要顺手重构无关代码或格式化无关文件。
- 工作区可能存在用户未提交改动。不要回滚、覆盖或清理非本次任务产生的变更。
- 面向项目已有模式实现功能，优先复用现有 package、构造函数、测试 fake 和错误处理方式。
- 修改 Go 文件后运行 `gofmt`；修改 Ent schema、配置、迁移或生成入口后同步检查实际 Go 代码是否仍可格式化、编译和测试。
- 文档、注释和面向开发者的说明优先使用中文。
- 类型、函数、常量必须有符合 Go 规范的中文注释；测试代码也应提供必要注释，方便理解测试意图。

## 项目定位

本项目是基于 initra 的 Go RESTful API 服务。当前包含 `auth`、`user`、`file`、`httpdemo`、`taskdemo` 示例模块，用来展示认证、用户管理、文件存储、HTTP Client 和任务队列等常见能力的组织方式；真实业务开发时应按业务域替换或裁剪示例模块，不要把示例接口当成业务契约。

项目依赖 `github.com/teamsillybees/initra/pkg/*` 提供 Web、配置、错误、日志、认证、数据库、Redis、缓存、文件存储、HTTP Client、任务队列等通用能力。业务代码不得 import `github.com/teamsillybees/initra/internal/*`，也不得复制 initra 根仓库的 `pkg/*` 到业务项目内。

## 常用命令

在项目根目录执行：

```powershell
go test ./... -count=1
go vet ./...
go build ./cmd/server
go generate ./internal/data
initra migrate diff <name> --env local
initra migrate apply --env local
initra migrate hash
initra skill
```

`initra skill` 会把 initra 框架使用说明初始化到 `.agents/skills/initra-framework`，供 Codex 在业务项目中识别框架边界和最佳实践；使用 Claude Code 时执行 `initra skill cc`。

涉及配置、迁移、依赖装配、路由权限或示例模块替换时，应运行与改动风险匹配的 `go test`、`go vet`、构建或迁移验证。若无法运行测试，最终回复中必须说明原因和未验证风险。

## Go 开发规范

- 遵循 Go 标准风格：小接口、显式错误处理、依赖通过构造函数注入。
- 包名保持简短、清晰、全小写；一个业务模块一个 package。
- 不引入不必要的抽象。只有在减少真实重复、隔离复杂度或匹配现有模式时才新增抽象。
- 共享能力优先放入项目内清晰的公共 package 或复用 `github.com/teamsillybees/initra/pkg/*`，不要通过业务模块之间互相 import 具体实现来复用逻辑。
- 业务 ID 统一使用 `github.com/teamsillybees/initra/pkg/idgen.ID`。Ent 主键、外键、auth user ID、REST path params、service 入参和 JSON VO 不再使用 `int64` 或 `string` 表达业务 ID；对外 JSON/OpenAPI ID 暴露为字符串。仅在对接雪花 ID 生成器、第三方库、手写 SQL 参数或极少数底层基础设施场景调用 `Int64()`。
- Ent schema 统一复用 `github.com/teamsillybees/initra/pkg/entx/mixin` 中的 ID、审计、软删除和乐观锁 mixin，业务项目不要复制本地 schema mixin。
- 禁止在业务模块中随意使用 `panic` 或吞掉错误；错误应向上返回并保留足够上下文。
- 业务专属错误码集中放在 `internal/modules/bizerrors` 或业务约定的错误包中，通过统一错误工厂创建。

## 模块结构

业务代码按业务模块组织为单一 flat package，不拆 controller/service/repository 子目录。模块文件结构遵循：

```text
internal/modules/<module>/
  <module>.handler.go HTTP 适配方法，调用 service 并包装响应
  <module>.dto.go     HTTP 边界类型：内部 Request/Response、Query、Body、VO
  <module>.service.go 业务逻辑 + Ent 数据库操作 + 必要的私有小接口
  <module>.routes.go  路由注册 + Module 结构体
  providers.go        samber/do 依赖注入
  cache.go            可选，缓存适配器或少量领域实体
  <module>_test.go    单元测试
```

- `*.handler.go` 可引用 DTO 类型，但不定义 DTO；它只负责参数转换、调用 service 和包装响应。
- `*.service.go` 直接通过 Ent Client 操作数据库，不默认拆独立 Repository 层。只有在确有跨存储适配、复杂事务隔离或外部系统适配需求时，才新增有明确职责的私有接口。
- `*.dto.go` 只放 HTTP 边界类型：Huma/Gin 包装类型用非导出的 `request`/`response` 后缀；HTTP 查询参数用 `Query` 后缀；HTTP 请求体用 `Body` 后缀；对外 JSON DTO 用 `VO` 后缀。
- `Response` 只表示 Huma/Gin 内部响应包装，不作为对外 JSON DTO 后缀；统一成功/错误响应等 JSON 结构使用 `VO` 命名。
- 不再定义 service 层 DTO，例如 `CreateUserDTO`、`UpdateUserDTO`；service 方法直接接受 HTTP 边界 `Body`/`Query` 类型并返回 `VO`。
- 不再新增默认的 `<module>.repo.go` 或 `<module>.model.go`。领域实体可放在 `cache.go`、`*.service.go` 或明确命名的专用文件中，但不要为了凑结构创建空文件。
- 模块之间禁止循环依赖。跨模块调用优先依赖调用方内部定义的小接口，避免依赖具体实现。

## 路由、安全与配置

- 所有 `/api/` 业务接口必须通过 `registry.Register` 登记 `RouteSecurity`，鉴权中间件默认 fail-closed。
- 公开接口必须显式设置 `AccessModePublic`。
- 登录即可访问的用户侧接口必须设置 `AccessModeAuthenticated`。
- 后台管理、运营操作、审核、退款、风控、配置管理等接口必须设置 `AccessModePermission`，其 `Resource`、`Action` 必须与 Casbin policy 文件保持一致。
- 业务配置结构放在 `internal/boot/config.go`，通过 `pkg/config` 泛型加载；配置应支持默认值、环境变量覆盖、启动校验和敏感配置脱敏打印。
- 运行环境统一使用环境变量 `APP_ENV` 表示；其他配置环境变量默认使用 `INITRA_` 前缀。
- 密码、Token、Secret、Access Key、Authorization、验证码、session value 和带密码的 DSN 禁止明文输出到日志。

## 框架包使用

- Redis 业务能力优先使用 `pkg/redisx` 的 client、Key Builder、缓存、Lua registry、SCAN+UNLINK 和 redislock 短锁；禁止生产使用 `KEYS`。
- 缓存能力优先复用 `pkg/cache`，业务侧保留明确的 key 命名、TTL 和失效策略；不要把 Redis value、验证码、session value 或 token 写入日志。
- 文件与对象存储业务能力优先使用 `pkg/storage` 的统一接口和 `pkg/storage/provider` 工厂；业务模块不要直接依赖云厂商 SDK。
- HTTP Client 由 `httpclient.Register(injector, cfg.HTTPClient)` 统一装配；业务模块通过命名 client 或最小接口依赖 `pkg/httpclient`，不要手写全局 client 工厂。
- 任务队列业务能力优先使用 `pkg/task` 的 `Publisher`、`Worker`、`Registry`、`Scheduler`、`Task` 和 `TaskMeta`；业务代码不得直接 import `github.com/hibiken/asynq`。
- 任务队列按 at-least-once 模型设计，不承诺 exactly-once；`biz_key` 是业务幂等键，外部副作用任务必须由业务侧保证幂等。

## 依赖装配

- 框架能力包应通过各自 `Register(injector, cfg)` 或 `Register(injector, options)` 风格入口函数在启动时显式调用一次完成装配。
- `internal/boot/providers.go` 优先调用各 `pkg/*` 的 `Register` 入口，例如 `logx.Register(injector, cfg.Log)`、`redisx.Register(injector, cfg.Redis)`、`storageprovider.Register(injector, cfg.Storage)`、`httpclient.Register(injector, cfg.HTTPClient)`、`asynqadapter.Register(injector, cfg.Task)`。
- 业务模块在自己的 `providers.go` 中注册 handler、service、module 等本地对象；业务代码只依赖接口，不感知底层实现。
- 禁止在业务模块中直接使用 `do.Invoke` 或滥用全局 injector。`do` 的使用边界应集中在 `providers.go` 和 `internal/boot` 装配层。
- 禁止为 package 装配做两阶段传参，例如先 `do.ProvideValue(injector, cfg.HTTPClient)` 再 `httpclient.Register(injector)`；配置应作为 `Register` 参数显式传入。

## 数据库与测试

- `internal/data/schema` 是数据库结构主源；`db/migrations` 是版本化迁移历史；`db/seeds` 保存种子数据。
- 手动修改 Ent schema 后运行 `go generate ./internal/data`，再生成迁移。
- 数据库集成测试优先使用 `go-sqlmock` 或项目已有测试辅助验证 SQL、事务和参数。
- 标准业务模块应包含单元测试，使用 fake 实现测试 service 编排逻辑。
- 架构边界相关改动必须保留或补充测试，确保业务项目不 import `github.com/teamsillybees/initra/internal/*`。
