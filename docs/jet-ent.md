# initra：从 go-jet 迁移到 Ent 的重构任务书

## 1. 背景

`initra` 是面向企业内部 Go 服务的快速开发脚手架，当前项目必须区分三类内容：

1. 标准项目模板：
   - `templates/api`
   - `templates/worker`
   - `examples/api`

2. 可复用 Go package：
   - 根模块 `github.com/teamsillybees/initra` 的 `pkg/*`

3. 工程化 CLI：
   - `cmd/initra`

当前 API 模板使用 `go-jet` 生成类型安全 SQL 代码。现在需要将 `go-jet` 替换为 `Ent`，并基于 Ent Mixin + Hook 实现统一自动填充能力，包括：

- 雪花 ID
- `created_at`
- `updated_at`
- `created_by`
- `updated_by`
- 可选：`deleted_at`
- 可选：`version`

本次重构必须保持 `initra` 的脚手架定位，不允许把运行时业务能力塞进 `cmd/initra`。

---

## 2. 当前 Jet 影响范围

需要重点处理以下路径：

```text
go.mod
go.sum
cmd/initra/main.go
cmd/initra/main_test.go

examples/api/go.mod
examples/api/go.sum
examples/api/README.md
examples/api/scripts/jet.ps1
examples/api/tools/jetgen/
examples/api/internal/gen/jet/
examples/api/internal/module/auth/auth.repo.go
examples/api/internal/module/user/user.repo.go
examples/api/internal/boot/providers.go
examples/api/test/architecture/architecture_test.go
examples/api/test/integration/auth_repository_test.go
examples/api/test/integration/user_repository_test.go

templates/api/go.mod.tmpl
templates/api/go.sum.tmpl
templates/api/README.md.tmpl
templates/api/scripts/jet.ps1.tmpl
templates/api/tools/jetgen/
templates/api/internal/gen/jet/
templates/api/internal/module/auth/auth.repo.go.tmpl
templates/api/internal/module/user/user.repo.go.tmpl
templates/api/internal/boot/providers.go.tmpl
templates/api/test/architecture/architecture_test.go.tmpl
templates/api/test/integration/auth_repository_test.go.tmpl
templates/api/test/integration/user_repository_test.go.tmpl

README.md
AGENTS.md
CLAUDE.md
````

---

## 3. 重构目标

### 3.1 必须实现

1. 删除 `go-jet` 运行时依赖。
2. 删除 `tools/jetgen`。
3. 删除 `scripts/jet.ps1`。
4. 删除 `internal/gen/jet`。
5. 引入 Ent schema 与 Ent 生成代码。
6. `examples/api` 可正常编译、测试、启动。
7. `templates/api` 与 `examples/api` 保持同步。
8. `cmd/initra new <app> --type api` 生成的新项目必须是 Ent 版本。
9. `auth/user` 基础模块继续可用。
10. DB schema、seed、Atlas migrations 继续可用。
11. 自动填充能力由 Ent Hook/Mixin 统一完成。
12. `service` 层不再手工设置雪花 ID、创建时间、更新时间、创建人、更新人。
13. `repository` 层使用 Ent Client，不再使用 Jet SQL Builder。

### 3.2 不做的事

1. 不把 Ent entity 直接暴露给 HTTP handler。
2. 不把 Ent entity 直接作为 API 响应对象。
3. 不移除现有业务领域模型。
4. 不让 `cmd/initra` 承载运行时业务能力。
5. 不允许标准模板或外部业务项目 import 根仓库 `internal/`。
6. 不引入 GORM。
7. 不改动 `templates/worker` 的定位，只保证它仍可编译。

---

## 4. 目标架构

迁移后采用以下分层：

```text
HTTP DTO / VO
    ↓
module service
    ↓
module repository
    ↓
internal/ent
    ↓
database
```

规则：

1. `handler` 只处理请求、响应、Huma Operation。
2. `service` 只处理业务规则、权限编排、缓存编排、事务编排。
3. `repository` 只处理 Ent 查询、持久化、领域模型转换。
4. `internal/ent` 只保存 Ent schema 和 Ent 生成代码。
5. `pkg/entx` 只保存可复用的 Ent 通用扩展，不依赖具体项目的 `internal/ent`。

---

## 5. 目标目录结构

在根模块中新增：

```text
pkg/
  entx/
    context.go
    audit_hook.go
    delete_guard_hook.go
    field.go
```

在 `examples/api` 中新增：

```text
examples/api/internal/
  ent/
    generate.go
    schema/
      mixin/
        id.go
        audit.go
        soft_delete.go
        optimistic_lock.go
      sys_user.go
      sys_role.go
      sys_user_role.go
      sys_menu.go
      sys_role_menu.go
      sys_config.go
      sys_dict_collection.go
      sys_dict_item.go
    ... Ent generated files ...

  data/
    ent_client.go
    tx.go
    soft_delete.go
```

在 `templates/api` 中同步新增模板文件：

```text
templates/api/internal/
  ent/
    generate.go.tmpl
    schema/
      mixin/
        id.go.tmpl
        audit.go.tmpl
        soft_delete.go.tmpl
        optimistic_lock.go.tmpl
      sys_user.go.tmpl
      sys_role.go.tmpl
      sys_user_role.go.tmpl
      sys_menu.go.tmpl
      sys_role_menu.go.tmpl
      sys_config.go.tmpl
      sys_dict_collection.go.tmpl
      sys_dict_item.go.tmpl
    ... Ent generated files as templates ...

  data/
    ent_client.go.tmpl
    tx.go.tmpl
    soft_delete.go.tmpl
```

---

## 6. 依赖调整

### 6.1 根模块 `go.mod`

删除：

```text
github.com/go-jet/jet/v2
tool github.com/go-jet/jet/v2/generator/postgres
```

新增：

```text
entgo.io/ent
```

如当前 Go 版本继续使用 `tool` 指令，可新增：

```text
tool entgo.io/ent/cmd/ent
```

如果 tool 指令不稳定，使用 `go generate` 直接执行：

```go
//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/versioned-migration ./schema
```

### 6.2 `examples/api/go.mod`

删除：

```text
github.com/go-jet/jet/v2
```

新增：

```text
entgo.io/ent
```

### 6.3 `templates/api/go.mod.tmpl`

同步执行同样修改。

---

## 7. Ent schema 设计规则

### 7.1 命名规则

Ent schema 使用业务表对应的实体名，但避免和模块领域模型冲突。

示例：

```text
sys_user             -> SysUser
sys_role             -> SysRole
sys_user_role        -> SysUserRole
sys_menu             -> SysMenu
sys_role_menu        -> SysRoleMenu
sys_config           -> SysConfig
sys_dict_collection  -> SysDictCollection
sys_dict_item        -> SysDictItem
```

### 7.2 表名规则

所有 Ent schema 必须通过 annotation 明确表名。

示例：

```go
func (SysUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_user"},
	}
}
```

### 7.3 字段规则

1. bigint 主键统一使用 `field.Int64("id")`。
2. PostgreSQL `VARCHAR(n)` 映射为 `field.String(...).MaxLen(n)`。
3. `TEXT` 映射为 `field.Text(...)` 或 `field.String(...)`，按字段语义选择。
4. `BOOLEAN` 映射为 `field.Bool(...)`。
5. `INTEGER` 映射为 `field.Int(...)`。
6. `TIMESTAMP` 映射为 `field.Time(...)`。
7. 可空字段使用 `Optional().Nillable()`。
8. 必填字符串字段使用 `NotEmpty()`。
9. 唯一约束使用 `Unique()` 或 `index.Fields(...).Unique()`。
10. 普通索引使用 `index.Fields(...)`。

---

## 8. Mixin 设计

### 8.1 IDMixin

路径：

```text
examples/api/internal/ent/schema/mixin/id.go
templates/api/internal/ent/schema/mixin/id.go.tmpl
```

要求：

```go
type ID struct {
	mixin.Schema
}

func (ID) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			Positive(),
	}
}
```

不要在 Mixin 中直接调用全局雪花生成器。

原因：

* 雪花节点号来自配置。
* 生成器应在运行时由 DI 容器注入。
* ID 自动填充应通过 Ent runtime hook 完成。

### 8.2 AuditMixin

字段：

```text
created_at
updated_at
created_by
updated_by
```

要求：

```go
type Audit struct {
	mixin.Schema
}

func (Audit) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
		field.Int64("created_by").Optional().Nillable(),
		field.Int64("updated_by").Optional().Nillable(),
	}
}
```

### 8.3 SoftDeleteMixin

字段：

```text
deleted_at
```

要求：

```go
type SoftDelete struct {
	mixin.Schema
}

func (SoftDelete) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (SoftDelete) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("deleted_at"),
	}
}
```

### 8.4 OptimisticLockMixin

可选字段：

```text
version
```

要求：

```go
type OptimisticLock struct {
	mixin.Schema
}

func (OptimisticLock) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("version").Default(1),
	}
}
```

默认不强制所有表启用。只给高并发更新表使用。

---

## 9. 自动填充设计

### 9.1 新增 `pkg/entx`

路径：

```text
pkg/entx/context.go
pkg/entx/audit_hook.go
pkg/entx/delete_guard_hook.go
pkg/entx/field.go
```

### 9.2 `pkg/entx/context.go`

实现以下能力：

```go
package entx

import "context"

type operatorIDKey struct{}

func WithOperatorID(ctx context.Context, operatorID int64) context.Context {
	return context.WithValue(ctx, operatorIDKey{}, operatorID)
}

func OperatorIDFromContext(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(operatorIDKey{}).(int64)
	return v, ok
}
```

要求：

1. HTTP 请求中优先从认证 principal 获取 operatorID。
2. seed、定时任务、系统任务可以使用 `entx.WithOperatorID(ctx, 0)`。
3. 没有 operatorID 时，`created_by`、`updated_by` 可以保持 NULL。

### 9.3 `pkg/entx/audit_hook.go`

实现运行时 Hook，统一处理：

| 操作                | 自动字段                                           |
| ----------------- | ---------------------------------------------- |
| create            | id、created_at、updated_at、created_by、updated_by |
| update/update_one | updated_at、updated_by                          |
| delete/delete_one | 默认拒绝，避免误物理删除                                   |

要求：

```go
type IDGenerator interface {
	NextID() int64
}

type OperatorFunc func(ctx context.Context) (int64, bool)

type AuditHookOptions struct {
	IDGen    IDGenerator
	Now      func() time.Time
	Operator OperatorFunc
}
```

Hook 逻辑：

1. `Now` 为 nil 时使用 `time.Now`。
2. create 时：

   * 如果 mutation 有 `id` 字段且尚未设置，则填充雪花 ID。
   * 如果 mutation 有 `created_at` 字段，则填充当前时间。
   * 如果 mutation 有 `updated_at` 字段，则填充当前时间。
   * 如果存在 operatorID 且 mutation 有 `created_by`，则填充。
   * 如果存在 operatorID 且 mutation 有 `updated_by`，则填充。
3. update/update_one 时：

   * 如果 mutation 有 `updated_at`，则填充当前时间。
   * 如果存在 operatorID 且 mutation 有 `updated_by`，则填充。
4. 不要因为某个 schema 没有某个字段就返回错误。
5. `SetField` 失败时，如果错误原因是字段不存在，应忽略。
6. 类型错误必须返回错误。

### 9.4 `pkg/entx/delete_guard_hook.go`

实现物理删除保护：

```go
func RejectDeleteHook() ent.Hook
```

规则：

1. 拦截 `ent.OpDelete` 和 `ent.OpDeleteOne`。
2. 默认返回错误。
3. 错误信息明确提示：使用 repository 的 soft delete 方法，不允许直接物理删除。
4. 如未来确实需要物理删除，通过单独的 context flag 显式放行，不在本次实现。

---

## 10. Ent Client 装配

### 10.1 新增 `examples/api/internal/data/ent_client.go`

职责：

1. 从现有 `*sql.DB` 构造 `*ent.Client`。
2. 注册 audit hook。
3. 注册 delete guard hook。
4. 导入 `ent/runtime`，确保 schema hook 可注册。

示例结构：

```go
package data

import (
	"context"
	"database/sql"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"{{ module }}/internal/ent"
	_ "{{ module }}/internal/ent/runtime"

	"github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/entx"
	"github.com/teamsillybees/initra/pkg/idgen"
)

func NewEntClient(db *sql.DB, generator *idgen.Generator) *ent.Client {
	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	client.Use(entx.AuditHook(entx.AuditHookOptions{
		IDGen: generator,
		Now:   time.Now,
		Operator: func(ctx context.Context) (int64, bool) {
			if id, ok := entx.OperatorIDFromContext(ctx); ok {
				return id, true
			}
			principal, ok := auth.PrincipalFromContext(ctx)
			if !ok {
				return 0, false
			}
			return principal.UserID, true
		},
	}))

	client.Use(entx.RejectDeleteHook())

	return client
}
```

模板版本中将 module import 替换为：

```go
"{{ .ModulePath }}/internal/ent"
_ "{{ .ModulePath }}/internal/ent/runtime"
```

### 10.2 修改 `examples/api/internal/boot/providers.go`

新增 provider：

```go
do.Provide(injector, func(i *do.Injector) (*ent.Client, error) {
	sqlDB := do.MustInvoke[*sql.DB](i)
	generator := do.MustInvoke[*idgen.Generator](i)
	return data.NewEntClient(sqlDB, generator), nil
})
```

注意：

1. 继续保留 `*sql.DB` provider。
2. `*sql.DB` 仍由 `pkg/db.Open` 创建。
3. Ent Client 使用既有 `*sql.DB`，不要重复创建连接池。
4. Shutdown 中由 `*sql.DB` 统一关闭底层连接。
5. 不要重复关闭同一个底层连接池。

---

## 11. 事务封装

### 11.1 新增 `examples/api/internal/data/tx.go`

要求：

```go
package data

import (
	"context"
	"fmt"

	"{{ module }}/internal/ent"
)

func WithinTx(ctx context.Context, client *ent.Client, fn func(context.Context, *ent.Client) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()

	if err := fn(ctx, tx.Client()); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("%w: rollback failed: %v", err, rerr)
		}
		return err
	}

	return tx.Commit()
}
```

规则：

1. repository 内部统一使用 `data.WithinTx`。
2. 回调参数使用 `*ent.Client`，不要把 `*ent.Tx` 泄漏到 service。
3. service 不直接感知 Ent 事务对象。
4. 如果跨模块事务后续复杂化，再引入 UnitOfWork，不在本次实现。

---

## 12. Repository 改造规则

### 12.1 通用规则

1. repository 持有 `*ent.Client`。
2. repository 不再持有 `*sql.DB`。
3. repository 不再持有 `idgen`。
4. repository 不再 import `github.com/go-jet/jet/v2/postgres`。
5. repository 不再 import `internal/gen/jet/table`。
6. repository 负责 Ent entity 与领域模型转换。
7. service 层继续使用领域模型 `User`，不直接使用 `*ent.SysUser`。
8. 查询必须默认排除 `deleted_at IS NOT NULL` 的记录。
9. 删除必须改为软删除。
10. 对唯一冲突、未找到、DB 错误进行现有 `apperrors` 风格包装。

### 12.2 `auth.repo.go`

将 Jet 查询替换为 Ent 查询。

目标方法：

```go
FindByUsername(ctx context.Context, username string) (*LoginUser, error)
FindByID(ctx context.Context, id int64) (*LoginUser, error)
```

规则：

1. 使用 `client.SysUser.Query()`。
2. 查询条件包括：

   * username/id
   * `deleted_at IS NULL`
3. 查询未命中返回 nil，不直接返回 DB 错误。
4. 禁用用户仍返回用户，由 service 决定是否拒绝登录。
5. 查询结果转换为 auth 模块内部结构体。

### 12.3 `user.repo.go`

将 Jet CRUD 替换为 Ent Builder。

Repository 结构：

```go
type Repository struct {
	client *ent.Client
}

func NewRepository(client *ent.Client) *Repository {
	return &Repository{client: client}
}
```

#### Create

要求：

1. 使用 `data.WithinTx`。
2. 创建 `sys_user`。
3. 不手工设置：

   * id
   * created_at
   * updated_at
   * created_by
   * updated_by
4. 角色关系使用 `sys_user_role` 显式实体创建。
5. `sys_user_role.id` 也由 Hook 填充。
6. 创建成功后，将生成的 ID 回写到领域模型 `user.ID`。

#### FindByID

要求：

1. 查询 `SysUser`。
2. 过滤 `deleted_at IS NULL`。
3. 加载用户角色编码。
4. 转换为领域模型 `*User`。
5. 未命中返回 nil。

#### FindByUsername

同 `FindByID`，条件改为 username。

#### Page

要求：

1. 支持 keyword 模糊搜索：

   * username
   * nickname
   * phone
   * email
2. 默认过滤软删除。
3. 排序：

   * `sort_id ASC`
   * `id ASC`
4. 分页：

   * `Limit(pageDTO.Limit())`
   * `Offset(pageDTO.Offset())`
5. 总数使用 `Count(ctx)`。
6. 列表查询后批量加载角色编码，避免 N+1。

#### Update

要求：

1. 使用 `data.WithinTx`。
2. 只更新允许修改的字段。
3. 不手工设置 `updated_at`、`updated_by`。
4. 如果角色编码发生变化，先软删除或物理删除旧关系需明确：

   * 推荐：对 `sys_user_role` 使用软删除。
   * 如果保留唯一约束 `user_id + role_id`，软删除后重复添加会冲突，需要把唯一约束改为部分唯一索引或改用恢复旧记录。
5. 更新成功后，重新查询并返回完整领域模型。

#### Delete

要求：

1. 不调用 Ent `DeleteOneID`。
2. 使用 Update 设置：

   * `deleted_at = now`
   * `updated_by = operatorID`
   * `updated_at` 由 Hook 自动处理
3. 删除用户时，用户角色关系同步软删除。
4. 删除完成后清理缓存。

---

## 13. Service 层改造规则

### 13.1 `user.service.go`

删除 `Service` 中的字段：

```go
idgen idGenerator
now   func() time.Time
```

删除构造函数参数：

```go
idgen idGenerator
now func() time.Time
```

Create 中删除手工赋值：

```go
ID
CreatedAt
UpdatedAt
CreatedBy
UpdatedBy
```

Update 中删除手工赋值：

```go
UpdatedAt
UpdatedBy
```

保留：

1. 参数校验。
2. 密码 hash。
3. 默认角色逻辑。
4. 默认启用状态逻辑。
5. 缓存清理逻辑。
6. 领域模型组装。

Create 时 operatorID 的传递方式：

1. handler 已经能拿到 operatorID。
2. handler 或 service 在调用 repository 前，将 operatorID 放入 context：

   * 推荐在 handler 层将请求 principal 转换为 context operator。
   * 如果现有 DTO 已有 `OperatorID`，service 可在调用 repo 前执行 `ctx = entx.WithOperatorID(ctx, input.OperatorID)`。
3. 不再把 operatorID 作为字段手工写入领域模型。

### 13.2 DTO 保持兼容

当前 DTO 中如果已有 `OperatorID`，本次可以先保留，减少改动范围。

后续可单独重构为：

```go
requestctx.OperatorIDFromContext(ctx)
```

---

## 14. Provider 改造

### 14.1 `examples/api/internal/module/user/providers.go`

修改：

```go
repo := NewRepository(do.MustInvoke[*ent.Client](i))
```

删除：

```go
*sql.DB
idgen
```

Service 构造同步删除 idgen 和 now 参数。

### 14.2 `examples/api/internal/module/auth/providers.go`

修改：

```go
repo := NewRepository(do.MustInvoke[*ent.Client](i))
```

删除 Jet/SQL 相关依赖。

### 14.3 模板同步

同样修改：

```text
templates/api/internal/module/user/providers.go.tmpl
templates/api/internal/module/auth/providers.go.tmpl
```

---

## 15. Ent schema 初始覆盖范围

必须为当前已有表建立 Ent schema：

```text
sys_user
sys_role
sys_user_role
sys_menu
sys_role_menu
sys_config
sys_dict_collection
sys_dict_item
```

这些 schema 必须覆盖当前 `examples/api/db/schema/*.sql` 中已有字段、索引、唯一约束和表名。

---

## 16. 外键策略

当前 SQL schema 中存在外键：

```text
sys_user_role.role_id -> sys_role.id
sys_user_role.user_id -> sys_user.id
sys_role_menu.role_id -> sys_role.id
sys_role_menu.menu_id -> sys_menu.id
sys_dict_item.collection_id -> sys_dict_collection.id
```

迁移到 Ent 后：

1. Ent schema 中可以声明 edge，保留关系表达和查询便利性。
2. 如果项目希望避免生产库强制外键，则迁移生成阶段关闭外键 DDL。
3. repository 层必须手工校验引用记录存在。
4. `sys_user_role`、`sys_role_menu` 必须作为显式 Ent schema，不使用 Ent 隐式 M2M。
5. 原因：这两张关系表有独立主键、审计字段、软删除字段，不是纯 join table。

---

## 17. Atlas 改造

### 17.1 当前状态

当前 `examples/api/db/atlas.hcl` 使用：

```hcl
schema {
  src = "file://db/schema"
}
```

### 17.2 目标状态

改为 Ent schema 作为目标源：

```hcl
schema {
  src = "ent://internal/ent/schema"
}
```

所有 env 均需要同步修改：

```text
local
dev
test
prod
```

### 17.3 `db/schema` 的处理

推荐：

1. 保留 `db/schema/*.sql` 作为历史参考或 DBA 阅读材料。
2. 不再作为 Atlas diff 的 schema source。
3. 在 README 中明确：数据库结构主源是 `internal/ent/schema`。
4. `db/migrations` 仍然是版本化迁移历史。
5. `db/seeds` 继续保存种子数据。

### 17.4 Atlas 命令保持

继续保留：

```text
scripts/atlas.ps1
```

命令语义不变：

```powershell
.\scripts\atlas.ps1 migrate diff add_xxx --env local
.\scripts\atlas.ps1 migrate apply --env local
.\scripts\atlas.ps1 migrate status --env local
```

---

## 18. Ent 生成脚本

删除：

```text
examples/api/scripts/jet.ps1
templates/api/scripts/jet.ps1.tmpl
examples/api/tools/jetgen
templates/api/tools/jetgen
```

新增：

```text
examples/api/scripts/ent.ps1
templates/api/scripts/ent.ps1.tmpl
```

脚本功能：

```powershell
go generate ./internal/ent
```

`internal/ent/generate.go`：

```go
package ent

//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/versioned-migration ./schema
```

---

## 19. CLI 改造

### 19.1 doctor 命令

当前 `cmd/initra/main.go` 中检查：

```text
go-jet
jet -version
```

改为检查：

```text
ent
go run entgo.io/ent/cmd/ent --help
atlas
```

测试同步修改：

```text
cmd/initra/main_test.go
```

### 19.2 new 命令

`initra new <app> --type api` 生成结果必须包含：

```text
internal/ent
internal/data
scripts/ent.ps1
```

且不得包含：

```text
internal/gen/jet
tools/jetgen
scripts/jet.ps1
```

### 19.3 CRUD/module 命令

如果当前 CLI 生成 CRUD 样例中有 Jet 代码，必须替换为 Ent repository 样例。

建议命令语义：

```text
initra module add <module>
```

只生成模块骨架。

```text
initra ent schema add <Entity> --table <table_name>
```

生成 Ent schema 骨架。

```text
initra crud add <module> --entity <Entity>
```

生成基于 Ent 的 CRUD repository 样例。

如果暂时不实现新命令，至少保证现有 CRUD 生成不再生成 Jet 代码。

---

## 20. 文档改造

需要替换所有 go-jet 表述。

### 20.1 根文档

修改：

```text
README.md
AGENTS.md
CLAUDE.md
```

将：

```text
go-jet 生成代码
Jet SQL Builder
```

替换为：

```text
Ent schema 与生成代码
Ent 类型安全持久化访问
```

### 20.2 示例文档

修改：

```text
examples/api/README.md
templates/api/README.md.tmpl
```

删除：

```text
.\scripts\jet.ps1 -Env local
```

新增：

```text
.\scripts\ent.ps1
.\scripts\atlas.ps1 migrate diff <name> --env local
```

明确：

```text
internal/ent/schema 是数据库结构主源
db/migrations 是版本化迁移历史
db/seeds 是种子数据
```

---

## 21. 测试改造

### 21.1 架构测试

修改：

```text
examples/api/test/architecture/architecture_test.go
templates/api/test/architecture/architecture_test.go.tmpl
internal/architecture/architecture_test.go
```

新增断言：

1. 不存在 `internal/gen/jet`。
2. 不存在 `tools/jetgen`。
3. 不存在 `scripts/jet.ps1`。
4. 存在 `internal/ent/schema`。
5. 存在 `internal/ent/client.go`。
6. 存在 `internal/data/tx.go`。
7. 模板项目不得 import 根仓库 `internal/`。
8. handler/service 不得 import `internal/ent`。
9. repository 可以 import `internal/ent`。
10. `pkg/*` 不得 import `examples/api/internal/ent` 或 `templates/api/internal/ent`。

### 21.2 集成测试

修改：

```text
examples/api/test/integration/auth_repository_test.go
examples/api/test/integration/user_repository_test.go
templates/api/test/integration/auth_repository_test.go.tmpl
templates/api/test/integration/user_repository_test.go.tmpl
```

验证：

1. 创建用户后自动生成 ID。
2. 创建用户后自动填充 `created_at`、`updated_at`。
3. 创建用户后自动填充 `created_by`、`updated_by`。
4. 更新用户后自动更新 `updated_at`、`updated_by`。
5. 删除用户后 `deleted_at` 不为空。
6. 删除用户后普通查询查不到。
7. 物理删除被拒绝。
8. auth repository 能按 username/id 查询用户。
9. user repository 分页正常。
10. 角色关系创建、更新、查询正常。

---

## 22. 软删除与唯一约束注意事项

当前关系表存在唯一约束：

```text
sys_user_role(user_id, role_id)
sys_role_menu(role_id, menu_id)
```

如果对关系表使用软删除，会出现以下问题：

1. 旧关系软删除后，重新添加相同关系会触发唯一约束冲突。
2. PostgreSQL 可以使用 partial unique index 解决：

   * `UNIQUE (user_id, role_id) WHERE deleted_at IS NULL`
3. Ent schema 中需要通过 annotation 或 migration 手工维护 partial index。
4. 如果暂时不处理 partial index，则关系表更新时应采用“恢复旧记录”策略：

   * 如果已存在软删除记录，更新 `deleted_at = NULL`。
   * 如果不存在，再创建新记录。

本次优先推荐：

```text
关系表使用恢复旧记录策略，避免立即改唯一索引。
```

后续可单独做 migration，将唯一约束改为 partial unique index。

---

## 23. 错误处理规则

保持现有错误体系：

```go
apperrors.New(...)
apperrors.Wrap(...)
bizerrors.UserNotFound(...)
```

Ent 错误映射规则：

1. `ent.IsNotFound(err)` -> 返回 nil 或业务 not found。
2. 唯一约束冲突 -> 转为业务冲突错误。
3. 其他 DB 错误 -> `apperrors.CodeDBError`。
4. Hook 自动填充失败 -> `apperrors.CodeInternalError` 或直接返回底层错误，由上层包装。
5. 不要把 Ent 原始错误直接暴露到 HTTP 响应。

---

## 24. 编码边界

### 24.1 允许 import Ent 的位置

允许：

```text
examples/api/internal/data
examples/api/internal/module/*/*.repo.go
examples/api/internal/boot/providers.go
examples/api/internal/ent
```

模板中对应路径同理。

### 24.2 不允许 import Ent 的位置

不允许：

```text
handler
service
dto
vo
pkg/*
cmd/initra
```

例外：

```text
pkg/entx 可以 import entgo.io/ent，但不能 import 具体项目的 internal/ent。
```

---

## 25. 分阶段执行顺序

### 阶段一：基础设施引入

1. 修改 `go.mod` / `go.sum`。
2. 新增 `pkg/entx`。
3. 新增 `examples/api/internal/ent/schema`。
4. 新增 `examples/api/internal/ent/generate.go`。
5. 执行 Ent generate。
6. 新增 `examples/api/internal/data/ent_client.go`。
7. 新增 `examples/api/internal/data/tx.go`。
8. 修改 `boot/providers.go` 注册 `*ent.Client`。

验收：

```bash
go test ./...
cd examples/api && go test ./...
```

### 阶段二：auth/user repository 迁移

1. 改造 `auth.repo.go`。
2. 改造 `user.repo.go`。
3. 修改 user/auth providers。
4. 修改 user service，删除手工自动填充。
5. 保证 auth/user 相关测试通过。

验收：

```bash
cd examples/api
go test ./internal/module/...
go test ./test/integration/...
```

### 阶段三：删除 Jet

1. 删除 `examples/api/internal/gen/jet`。
2. 删除 `examples/api/tools/jetgen`。
3. 删除 `examples/api/scripts/jet.ps1`。
4. 删除模板中的对应文件。
5. 删除 go-jet 依赖。
6. 修改 CLI doctor。
7. 修改测试断言。
8. 修改 README/AGENTS/CLAUDE。

验收：

```bash
grep -R "go-jet\|jet/v2\|internal/gen/jet\|tools/jetgen\|jet.ps1" . --exclude-dir=.git
```

结果应为空，文档中历史说明除外。如保留历史说明，必须明确标注为“旧版”。

### 阶段四：模板同步

1. 将 `examples/api` 中 Ent 相关结构同步到 `templates/api`。
2. 模板文件中的模块路径必须使用 `{{ .ModulePath }}`。
3. 运行 CLI 生成新 API 项目。
4. 在生成项目中运行测试。

验收：

```bash
go test ./...
initra new demo --type api
cd demo
go test ./...
```

### 阶段五：Atlas 源切换

1. 修改 `examples/api/db/atlas.hcl`。
2. 修改 `templates/api/db/atlas.hcl.tmpl`。
3. 将 schema source 改为 `ent://internal/ent/schema`。
4. 保留 `db/migrations`。
5. 保留 `db/seeds`。
6. 更新文档说明 `db/schema` 降级为参考。

验收：

```powershell
.\scripts\atlas.ps1 migrate status --env local
.\scripts\atlas.ps1 migrate diff ent_baseline_check --env local
```

如果 diff 生成空迁移或无结构差异，说明 Ent schema 与现有 migration 基本对齐。

---

## 26. 最终验收标准

必须满足：

1. 根模块测试通过：

```bash
go test ./...
```

2. 示例 API 测试通过：

```bash
cd examples/api
go test ./...
```

3. 模板生成项目测试通过：

```bash
initra new demo --type api
cd demo
go test ./...
```

4. 全仓库不再存在 Jet 运行时依赖：

```bash
grep -R "github.com/go-jet/jet" . --exclude-dir=.git
```

5. 全仓库不再存在 Jet 生成目录：

```bash
find . -path "*internal/gen/jet*"
```

6. 全仓库不再存在 Jet 工具目录：

```bash
find . -path "*tools/jetgen*"
```

7. 全仓库不再存在 Jet 脚本：

```bash
find . -name "jet.ps1"
```

8. `examples/api` 中存在：

```text
internal/ent/schema
internal/ent/client.go
internal/data/ent_client.go
internal/data/tx.go
scripts/ent.ps1
```

9. `templates/api` 中存在对应模板文件。

10. `handler/service/dto/vo` 不直接 import `internal/ent`。

---

## 27. 禁止事项

1. 禁止把 Ent entity 作为 HTTP 响应对象。
2. 禁止在 service 中直接调用 Ent Client。
3. 禁止在 handler 中直接调用 Ent Client。
4. 禁止在 `pkg/*` 中 import `examples/api/internal/ent`。
5. 禁止在 `cmd/initra` 中 import Ent 生成代码。
6. 禁止继续保留 `go-jet` 依赖。
7. 禁止继续保留 `tools/jetgen`。
8. 禁止继续保留 `internal/gen/jet`。
9. 禁止直接调用 Ent Delete 进行物理删除。
10. 禁止让 `db/schema/*.sql` 与 `internal/ent/schema` 同时作为 schema 主源。

---

## 28. 推荐最终说明

重构完成后，项目说明中应统一表述为：

```text
initra 的 API 标准模板使用 Ent 作为类型安全持久化访问层，使用 Ent schema 描述数据库结构，使用 Atlas 管理版本化迁移，使用 Ent Mixin + Runtime Hook 实现雪花 ID、审计字段、软删除等通用自动填充能力。

examples/api 是 templates/api 的可运行验证样例；templates/api 与 examples/api 必须保持同步。cmd/initra 只负责生成工程骨架，不承载运行时业务能力。pkg/* 提供可复用 Go package，其中 pkg/entx 仅提供 Ent 通用 Hook 和上下文工具，不依赖具体项目生成的 internal/ent。
```
