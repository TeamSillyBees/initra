# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

`initra` 是面向企业内部 Go 服务的快速开发脚手架，包含三部分：

1. **可复用 Go package**（`pkg/*`）：Web、配置、错误、日志、认证、数据访问、Redis、缓存、文件存储、HTTP Client、任务队列等通用能力
2. **标准项目模板**（`templates/api`）：通过 `initra new` 生成 API 项目
3. **工程化 CLI**（`cmd/initra`）：Cobra 驱动的项目生成与管理工具

`examples` 是 API 模板的可运行示例，也是本地联调的验证项目。

## 技术栈

Web：Gin + Huma v2 | 数据库：PostgreSQL + Ent + Atlas 迁移 | 认证授权：JWT + Casbin RBAC
配置/DI：Viper + samber/do | 错误：oops + 统一错误码 | 日志：zap
ID 生成：snowflake | Redis：go-redis | 缓存：jetcache-go | 任务队列：Asynq（通过 `pkg/task` 抽象层）
存储：local / OSS / COS / S3 | CLI：Cobra

## 常用命令

```powershell
# 根模块测试与检查
go test ./pkg/... ./cmd/initra/... ./internal/... -count=1
go vet ./pkg/... ./cmd/initra/... ./internal/...

# 示例项目测试与检查
go test ./examples/... -count=1
go vet ./examples/...

# 构建 CLI
go build -o initra.exe ./cmd/initra

# 运行示例项目
go run ./examples/cmd/server

# 通过 CLI 生成新项目（本地开发版）
$framework = (Resolve-Path .).Path
go run ./cmd/initra new $env:TEMP\demo-api --type api --module example.com/demo-api --replace $framework

# Ent 代码生成（在示例项目目录执行）
cd examples && go run ./internal/data/entgenerate

# 数据库迁移（在项目目录执行）
go run ./internal/data/migratediff/main.go <name>   # 生成迁移 diff
```

仓库使用 `go.work` 联调根模块和 `examples`，在仓库根目录执行 Go 命令即可覆盖两个模块。

## 核心架构

### 模块分层（示例项目 `examples`）

每个业务模块在 `internal/modules/<name>/` 下按 flat package 组织：

| 文件 | 职责 |
|------|------|
| `*.dto.go` | HTTP 边界类型：`Body`（请求体）、`Query`（查询参数）、`VO`（对外 JSON DTO）、非导出 `request`/`response`（Huma 包装类型） |
| `*.model.go` | 领域实体和 service/repo 的结构体入参，使用 `DTO` 后缀 |
| `*.handler.go` | HTTP Handler，每个方法对应一个 Huma operation |
| `*.service.go` | 业务逻辑层 |
| `*.repo.go` | 数据访问层（Ent client） |
| `*.routes.go` | `Module` 结构体，负责注册 Huma operation 和 Casbin 安全策略 |
| `providers.go` | `Provide()` 函数，用 samber/do 注册该模块的依赖链 |
| `cache.go` | 缓存层（jetcache-go） |

依赖注入方向：`Handler → Service → Repository + Cache`，全部通过 `providers.go` 在 do 容器中组装。

### 启动流程

`cmd/server/main.go` → `boot.Bootstrap()` → 依次执行：
1. `LoadConfig` — Viper 加载配置
2. `registerProviders` — 基础设施注入（DB、Redis、Logger、JWT、Casbin、Storage、Task 等）
3. `registerModules` — 业务模块依赖注入
4. `registerRoutes` — 注册 Huma routes + Casbin 安全策略
5. 返回 `*Application` 聚合根，调用 `app.Run(ctx)` 启动 HTTP Server

### 统一响应格式

所有 API 响应由 `pkg/response.SuccessVO[T]` 包裹，JSON 格式：

```json
{"code": "OK", "message": "success", "data": {...}, "traceId": "..."}
```

错误响应由 `pkg/errors.ErrorVO` 包裹，通过 Huma 中间件自动映射。

### JSON 命名约定

- **对外 API DTO**（`*.dto.go` 中的 Body/Query/VO）：使用 **lowerCamelCase** 的 JSON tag，如 `json:"userId"`、`json:"roleCodes"`
- **内部 DB 模型**（ent 生成的 `data/ent/*.go`）：使用 snake_case JSON tag（由 Ent 自动生成，不要手动修改）
- **内部任务队列 payload**：使用 snake_case（如 `sendEmailPayload`）

### 命名约定

- HTTP 类型：非导出 `request`/`response` 后缀、`Query`、`Body`、`VO`
- 领域 DTO：`DTO` 后缀（如 `PageUsersDTO`）
- 禁止使用 `Result` 后缀命名返回值；列表直接用 `[]T`，分页用 `pagination.PageVO[T]`
- 分页参数：使用 `pagination.PageQuery` 嵌入 + `query` tag

### CLI 命令体系

```
initra new <app> --type api           生成 API 项目
initra module add <name>              添加业务模块
initra crud add <module>              添加 CRUD 样例
initra config add <capability>        添加配置能力
initra migrate new|diff|apply|hash    数据库迁移管理
initra skill codex                    注入 Codex skill
initra skill cc                       注入 Claude Code skill
initra doctor                         诊断环境
```

## 配置规范

- 项目配置结构体放在 `internal/boot/config.go`，使用 `mapstructure` tag
- 加载顺序：默认值 → `configs/config.yaml` → `configs/config.<env>.yaml` → 环境变量覆盖 → `Validate()` 校验
- 环境变量默认前缀 `INITRA_`，运行环境由 `APP_ENV` 或 `--env` 决定
- `pkg/config.LoadInto[T]()` 是泛型配置加载入口

## 关键设计决策

- **任务队列不直接依赖 Asynq**：`pkg/task` 定义抽象接口，`pkg/task/asynqadapter` 提供 Asynq 适配，业务代码只依赖 `task.Publisher` / `task.TaskMeta`
- **Ent 代码不入库**：模板只保存 schema、mixin、`generate.go`，`initra new` 生成项目后自动执行 Ent 代码生成
- **统一错误模型**：`pkg/errors.AppError` 携带 Code + Message + HTTP Status + Details + Cause，通过 `oops` 包装底层错误
- **路由安全元信息**：每个路由注册时必须同时注册 `RouteSecurity{Resource, Action}`，Casbin 中间件据此鉴权
- **泛型配置加载**：`pkg/config` 不绑定任何业务配置结构，通过泛型 `LoadInto[T]` 实现
