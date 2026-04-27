# initra

`initra` 是一个面向中大型 Golang Web API 项目的脚手架模板，采用垂直切片架构，将平台层基础设施与业务模块显式分离，并内置 auth/user 两个参考模块。

## 技术栈

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

## 目录概览

```text
cmd/server                 程序入口
configs/                   环境配置与 Casbin 策略
db/                        Atlas 配置、desired schema 目录、migration、seed
internal/boot              应用组合根，负责依赖注入、模块装配和 HTTP Server 构建
internal/platform/         配置、日志、数据库、缓存、鉴权、Web 等基础设施
internal/app/auth          登录、刷新 token、当前用户信息
internal/app/user          用户 CRUD 与用户详情缓存
internal/gen/jet           Jet 生成代码目录
scripts/                   Atlas、Jet、模块生成与 bootstrap 脚本
test/                      integration / e2e 测试
```

## 快速开始

1. 启动本地依赖

```bash
make up
```

2. 执行数据库迁移

```bash
make migrate-apply ENV=local
```

3. 重新生成 Jet 代码

```bash
make jet
```

4. 启动服务

```bash
make dev
```

## 默认账号

- 用户名：`admin`
- 密码：`admin123`

默认管理员定义在 `db/seeds/001_seed_admin.sql`；首次初始化数据库后可手动执行该 seed。

## 文档与接口

- OpenAPI JSON：`GET /openapi.json`
- OpenAPI YAML：`GET /openapi.yaml`
- 在线文档：`GET /docs`
- 健康检查：`GET /health`
- 就绪检查：`GET /ready`
- 版本信息：`GET /version`

## 常用命令

```bash
make test
make fmt
make vet
make lint
make new-module NAME=order
make migrate-new NAME=create_order_table
make docker-build
```

Windows / PowerShell 环境可直接使用同名脚本，避免依赖 `sh`：

```powershell
.\scripts\bootstrap.ps1
.\scripts\atlas.ps1 migrate apply --env local
.\scripts\jet.ps1
.\scripts\new_module.ps1 -Name order
.\scripts\new_migration.ps1 -Name create_order_table
$env:APP_ENV = "local"; go run ./cmd/server
```

## Jet 与 Atlas

- `scripts/jet.sh` 默认从本地 PostgreSQL 读取 `public` schema，并输出到 `internal/gen/jet`
- `scripts/jet.ps1` 提供等价的 Windows / PowerShell 入口
- `scripts/atlas.sh` 固定读取 `db/atlas.hcl`
- `scripts/atlas.ps1` 提供等价的 Windows / PowerShell 入口
- `db/schema/` 是 Atlas 的 desired database schema 目录，每个表单独一个 SQL 文件
- 迁移目录固定为 `db/migrations`
- `make migrate-diff NAME=xxx` 会基于 `db/schema/` 和迁移目录生成新的 versioned migration
- 系统表结构以 `paperlingo-server` 中的 `V1.0.0__init_sys.sql` 为基准参考，并按脚手架需要做了适配

## 当前实现范围

- 平台层：配置、日志、数据库、缓存、JWT、Casbin、Gin/Huma、统一错误响应、health/ready/version
- 业务层：auth 登录/刷新/me，user CRUD/列表/详情
- 工程化：Makefile、Dockerfile、docker-compose、Atlas/Jet 脚本、PowerShell 脚本、模块生成脚本
- 测试：service 单元测试、repository 测试、HTTP e2e 测试、架构边界测试
