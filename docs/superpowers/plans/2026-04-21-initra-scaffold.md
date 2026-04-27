# Initra Scaffold Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 基于设计文档实现一个可运行、可扩展、具备平台层与示例模块的 Golang Web API 脚手架。

**Architecture:** 采用垂直切片架构，将平台基础设施集中在 `internal/platform`，业务模块集中在 `internal/app`。Gin 负责 HTTP 承载，Huma 负责契约与 OpenAPI，服务层通过 `context.Context`、事务接口、仓储接口与缓存接口协作。

**Tech Stack:** Go 1.26、gin、huma、go-jet、atlas、redis、jetcache-go、do、zap、oops、jwt、casbin、viper、PostgreSQL

---

### Task 1: 建立项目基础骨架

**Files:**
- Create: `cmd/server/main.go`
- Create: `internal/boot/*.go`
- Create: `internal/platform/config/*.go`
- Create: `internal/platform/logging/*.go`
- Create: `internal/platform/web/*.go`
- Create: `internal/platform/observability/*.go`
- Test: `internal/platform/config/config_test.go`
- Test: `internal/platform/observability/handler_test.go`

- [ ] **Step 1: 先写配置与健康检查测试**
- [ ] **Step 2: 运行对应测试并确认失败**
- [ ] **Step 3: 实现配置加载、日志初始化、Web 启动与健康检查**
- [ ] **Step 4: 重新运行测试并确认通过**

### Task 2: 建立数据库、缓存与 ID 基础设施

**Files:**
- Create: `internal/platform/database/*.go`
- Create: `internal/platform/cache/*.go`
- Create: `internal/platform/idgen/*.go`
- Create: `internal/gen/jet/.gitkeep`
- Test: `internal/platform/idgen/generator_test.go`
- Test: `internal/platform/database/tx_test.go`

- [ ] **Step 1: 先写事务与 ID 生成器测试**
- [ ] **Step 2: 运行对应测试并确认失败**
- [ ] **Step 3: 实现 PostgreSQL、事务 helper、缓存 key helper、Snowflake 生成器**
- [ ] **Step 4: 重新运行测试并确认通过**

### Task 3: 建立统一错误、认证与授权能力

**Files:**
- Create: `internal/platform/errors/*.go`
- Create: `internal/platform/auth/*.go`
- Test: `internal/platform/errors/mapper_test.go`
- Test: `internal/platform/auth/jwt_test.go`

- [ ] **Step 1: 先写错误映射与 JWT 服务测试**
- [ ] **Step 2: 运行对应测试并确认失败**
- [ ] **Step 3: 实现错误码、标准响应、JWT、当前用户上下文、Casbin 初始化与中间件**
- [ ] **Step 4: 重新运行测试并确认通过**

### Task 4: 实现 auth/user 示例模块

**Files:**
- Create: `internal/app/auth/**`
- Create: `internal/app/user/**`
- Create: `db/migrations/*.sql`
- Test: `internal/app/auth/domain/service_test.go`
- Test: `internal/app/user/domain/service_test.go`

- [ ] **Step 1: 先写 auth 与 user 服务测试**
- [ ] **Step 2: 运行对应测试并确认失败**
- [ ] **Step 3: 实现模块装配、服务、仓储、缓存、策略与 Huma API**
- [ ] **Step 4: 重新运行测试并确认通过**

### Task 5: 工程化完善与交付

**Files:**
- Create: `configs/*`
- Create: `db/atlas.hcl`
- Create: `scripts/*`
- Create: `Makefile`
- Create: `docker-compose.yml`
- Create: `Dockerfile`
- Create: `README.md`
- Test: `test/e2e/*`

- [ ] **Step 1: 完善本地配置、迁移脚本、模块生成脚本与开发命令**
- [ ] **Step 2: 增加 e2e/集成测试样例**
- [ ] **Step 3: 运行 `go test ./...`、`go build ./cmd/server` 等完整验证**
- [ ] **Step 4: 根据验证结果修正直至通过**
