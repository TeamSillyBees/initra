# AGENTS.md

本文件指导 Codex 在本仓库中开展开发、修改、审查和验证工作。适用范围为仓库根目录及其所有子目录；若用户在对话中给出更具体的要求，以用户要求为准。

## 工作原则

- 先阅读上下文，再修改文件。优先查看 `README.md`、本文件、相关包代码、测试和模板文件。
- 保持改动聚焦，只处理用户要求的任务；不要顺手重构无关代码或格式化无关文件。
- 工作区可能存在用户未提交改动。不要回滚、覆盖或清理非本次任务产生的变更。
- 面向仓库已有模式实现功能，优先复用现有 package、构造函数、测试 fake 和错误处理方式。
- 修改 Go 文件后运行 `gofmt`；修改模板时同步检查生成后的实际 Go 代码是否仍可格式化和编译。
- 文档、注释和面向开发者的说明优先使用中文。Go 导出类型、函数、常量必须有符合 Go 规范的中文注释。

## 项目定位

`initra` 是面向企业内部 Go 服务的快速开发脚手架。理解和描述本仓库时必须区分三类内容：

- **标准项目模板**：`templates/api` 提供 RESTful API 服务骨架，`templates/worker` 提供后台 worker 占位骨架；`examples/api` 是 API 模板的可运行验证样例。
- **可复用 Go package**：根模块 `github.com/teamsillybees/initra` 的 `pkg/*`，沉淀 Web、配置、错误、日志、认证、数据库、Redis、缓存、对象存储、HTTP Client、任务调度等通用能力。
- **工程化 CLI**：`cmd/initra`，负责生成项目、业务模块、CRUD 样例、迁移文件、配置样例、接口骨架、测试骨架和代码生成命令。

重要边界：

- `examples/api` 是独立 Go module 的可运行 API 示例项目，属于标准项目模板。
- `templates/api` 是 CLI API 项目模板，内容应与 `examples/api` 保持同步。
- `templates/worker` 是 CLI worker 项目模板，目前只提供可编译占位骨架。
- `cmd/initra` 只负责生成和维护工程骨架，不承载运行时业务能力。
- `internal/` 只服务脚手架仓库自身；标准项目模板、生成项目和外部业务项目不得 import 根仓库 `internal/`。

## 常用命令

在仓库根目录执行：

```powershell
# 根模块测试与静态检查
go test ./pkg/... ./cmd/initra/... ./internal/... -count=1
go vet ./pkg/... ./cmd/initra/... ./internal/...

# 示例项目测试与静态检查
go test ./examples/api/... -count=1
go vet ./examples/api/...

# 构建 CLI
go build -o $env:TEMP\initra.exe ./cmd/initra

# 生成示例项目进行验证
$target = Join-Path $env:TEMP "demo-api"
go run ./cmd/initra new $target --type api --module example.com/demo-api --replace (Resolve-Path .).Path
```

仓库使用 `go.work` 联调根模块和 `examples/api`。涉及模板生成、包边界或示例项目行为时，应同时验证根模块和 `examples/api`。

## Go 开发规范

- 遵循 Go 标准风格：小接口、显式错误处理、依赖通过构造函数注入。
- 包名保持简短、清晰、全小写；一个业务模块一个 package。
- 不引入不必要的抽象。只有在减少真实重复、隔离复杂度或匹配现有模式时才新增抽象。
- 共享能力放入 `pkg/*`，不要通过业务模块之间互相 import 来复用逻辑。
- 新增或修改导出 API 时，同步补齐中文注释和必要测试。
- 禁止在业务模块中随意使用 `panic` 或吞掉错误；错误应向上返回并保留足够上下文。

## 架构约束

### 标准项目模板

业务代码按业务模块组织为单一 flat package，不拆 controller/service/repository 子目录。模块主文件按职责命名，必要配套能力可用独立文件承载。模块文件结构应遵循：

```text
internal/module/<module>/
  <module>.handler.go Handler + 请求/响应 DTO + Huma output 类型
  <module>.service.go 业务逻辑 + 私有接口定义
  <module>.repo.go    数据库实现
  <module>.model.go   领域实体 + 输入/输出参数类型
  <module>.routes.go  路由注册 + Module 结构体
  providers.go        samber/do 依赖注入
  cache.go            可选，缓存适配器
  <module>_test.go    单元测试
```

- 模块之间禁止循环依赖。
- 跨模块调用优先依赖调用方内部定义的小接口，避免依赖具体实现。
- 业务模块应保持独立，不互相 import 具体实现。
- 示例项目只能依赖根模块的 `pkg/*`，不能 import 根仓库 `internal/`。

### 模板同步

- 修改 `examples/api` 中的模板来源代码时，检查是否需要同步到 `templates/api/*.tmpl`。
- 修改 `templates/api` 或 `templates/worker` 时，确认生成项目仍能通过 `go test`、`go vet` 和必要的 CLI 生成验证。
- 模板文件中的模块路径必须使用 `{{ .ModulePath }}`，禁止硬编码 `github.com/teamsillybees/initra/examples/api`。
- 发布版 CLI 可写入自身构建版本；开发版生成项目必须使用 `--framework-version` 或 `--replace`，避免不可复现依赖。

### 错误处理

- 可复用错误码由 `pkg/errors` 定义，如 `CodeBadRequest`、`CodeUnauthorized`、`CodeInternalError`。
- 业务专属错误码可放在 `internal/module/bizerrors/`，通过 `apperrors.New` 工厂函数创建。
- 业务模块内部不要新增 sentinel error，例如直接 `errors.New`；统一使用业务错误定义。
- HTTP 响应错误映射应复用现有 mapper 和 response 机制，不要在 handler 中手写不一致的错误响应结构。

### 路由、安全与配置

- 所有 `/api/` 接口必须通过 `registry.Register` 登记 `RouteSecurity`，鉴权中间件默认 fail-closed。
- 公开接口，例如登录接口，必须显式设置 `Public: true`。
- `RouteSecurity` 的 `Resource`、`Action` 必须与 Casbin policy 文件保持一致。
- 业务项目在自己的 `internal/boot/config.go` 定义配置结构，并通过 `pkg/config` 泛型加载。
- `pkg/config` 只提供通用加载能力，不绑定任何具体业务配置结构。

## 测试要求

- 窄改动至少运行相关包测试；影响共享 package、CLI、模板或生成项目时运行对应完整命令。
- 标准项目模板每个业务模块应包含单元测试，使用 fake 实现测试 service 编排逻辑。
- 数据库集成测试优先使用 `go-sqlmock` 验证 SQL 生成和参数。
- 架构边界相关改动必须保留或补充测试，确保示例项目不 import 根仓库 `internal/`。
- 若无法运行测试，必须在最终回复中说明原因和未验证风险。

## Codex 工作流

1. 明确任务影响范围：根模块、`pkg/*`、`cmd/initra`、`examples/api`、`templates/api`、`templates/worker` 或文档。
2. 使用 `rg` / `rg --files` 搜索相关代码与测试，避免凭记忆修改。
3. 修改前确认工作区状态，保护用户已有改动。
4. 按最小可行范围编辑文件；手动编辑优先使用补丁方式。
5. 运行与改动风险匹配的 `go test`、`go vet`、构建或生成验证。
6. 最终回复简要说明改了什么、验证了什么，以及任何未完成或未验证事项。
