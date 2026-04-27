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
test/                      integration / e2e 测试
```

## 项目初始化

1. 安装命令行工具

Atlas 用于维护数据库迁移，Jet 用于从数据库 schema 生成类型安全的 SQL 构造代码。安装后需要确保 `atlas` 和 `jet` 都在 `PATH` 中。

macOS / Linux 可使用官方脚本安装 Atlas：

```bash
curl -sSf https://atlasgo.sh | sh
```

使用 Homebrew 时可这样安装 Atlas：

```bash
brew install ariga/tap/atlas
```

Windows 环境建议下载 Atlas Windows AMD64 二进制，保存为 `atlas.exe` 并将所在目录加入 `PATH`：

```powershell
# 下载地址：https://atlasbinaries.com/atlas/atlas-windows-amd64-latest.exe
# 示例安装目录：C:\atlas\atlas.exe
$atlasPath = "C:\atlas"
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
[Environment]::SetEnvironmentVariable("Path", "$currentPath;$atlasPath", "User")
```

Jet CLI 通过 Go 工具链安装：

```bash
go install github.com/go-jet/jet/v2/cmd/jet@latest
```

如果 `jet` 无法直接执行，请确认 `$env:GOBIN` 或 `$env:GOPATH\bin` 已加入 `PATH`。

安装完成后验证：

```bash
atlas version
jet -help
```

2. 启动本地依赖

```bash
docker compose up -d postgres redis
```

3. 执行数据库迁移

```bash
atlas -c file://db/atlas.hcl migrate apply --env local
```

4. 重新生成 Jet 代码

```bash
jet -dsn="postgresql://postgres:postgres@127.0.0.1:5432/initra?sslmode=disable" -schema=public -path=./internal/gen/jet
```

5. 初始化默认账号

```bash
psql "postgresql://postgres:postgres@127.0.0.1:5432/initra?sslmode=disable" -f db/seeds/001_seed_admin.sql
```

6. 启动服务

在 IDEA 中运行 `cmd/server`，并设置环境变量 `APP_ENV=local`。

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

本项目不再维护 Makefile 和脚本目录，常用入口直接使用原生命令。Go 运行、测试、格式化、构建等命令交由 IDEA Run Configuration 或内置工具链管理，README 不再重复列出。

### 本地依赖

```bash
docker compose up -d postgres redis
docker compose ps
docker compose logs -f postgres redis
docker compose down
```

### Docker 镜像

```bash
docker build -t initra:latest .
```

### 数据库迁移

```bash
atlas -c file://db/atlas.hcl migrate diff <name> --env local
atlas -c file://db/atlas.hcl migrate apply --env local
atlas -c file://db/atlas.hcl migrate status --env local
```

### Jet 代码生成

```bash
jet -dsn="postgresql://postgres:postgres@127.0.0.1:5432/initra?sslmode=disable" -schema=public -path=./internal/gen/jet
```

### Seed 数据

```bash
psql "postgresql://postgres:postgres@127.0.0.1:5432/initra?sslmode=disable" -f db/seeds/001_seed_admin.sql
```

## Atlas 与数据库

- `db/schema/` 是 Atlas 的 desired database schema 目录，每个表单独一个 SQL 文件
- 迁移目录固定为 `db/migrations`
- `atlas -c file://db/atlas.hcl migrate diff <name> --env local` 会基于 `db/schema/` 和迁移目录生成新的 versioned migration

## 当前实现范围

- 平台层：配置、日志、数据库、缓存、JWT、Casbin、Gin/Huma、统一错误响应、health/ready/version
- 业务层：auth 登录/刷新/me，user CRUD/列表/详情
- 工程化：Dockerfile、docker-compose、Atlas 配置、Jet 代码生成说明、README 命令清单
- 测试：service 单元测试、repository 测试、HTTP e2e 测试、架构边界测试
