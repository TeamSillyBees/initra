# 业务 ID 与 Ent Schema

## ID 规则

- Ent 主键、外键、auth user ID、REST path、service 入参和 JSON VO 使用 `github.com/teamsillybees/initra/pkg/idgen.ID`。
- 对外 JSON/OpenAPI ID 表示为字符串；Huma 示例使用 `"1771234567890123456"`。
- 只有雪花生成器、第三方库、手写 SQL 参数或底层基础设施才调用 `Int64()`。
- 在入口处拒绝 `idgen.ID <= 0`。
- `pkg/idgen` 不提供固定的包级默认节点；应用启动时必须调用 `idgen.Register`，并为每个并行实例显式配置唯一的 0–1023 节点号。缺少配置应启动失败。

## Schema 字段

当前 schema 直接组合 `fieldx` 字段助手，不使用 Ent mixin：

```go
func (SysUser) Fields() []ent.Field {
	fields := []ent.Field{
		fieldx.ID(),
		field.String("username").NotEmpty().Unique(),
	}
	fields = append(fields, fieldx.SoftDelete()...)
	fields = append(fields, fieldx.Audit()...)
	return fields
}
```

- `fieldx.ID()` 使用 `idgen.ID`，并通过 `DefaultFunc(idgen.NextID)` 生成主键。
- `fieldx.Audit()` 提供 `created_at`、`updated_at`、`created_by`、`updated_by` 字段；时间由 schema default 维护。
- `fieldx.SoftDelete()` 只提供 `deleted_at` 字段；查询过滤和软删除更新由 service 显式实现。
- 需要常用软删除索引时使用 `indexx.SoftDelete()`。
- `fieldx.OptimisticLock()` 目前只提供 version 字段，不包含自动 compare-and-swap、递增或冲突检测。

外键字段显式保留 `idgen.ID` Go 类型：

```go
field.Int64("role_id").GoType(idgen.ID(0)).Positive()
```

## Runtime Hook

项目 Ent Client 注册：

```go
client.Use(entx.AuditHook(entx.AuditHookOptions{Operator: operatorFunc}))
client.Use(entx.RejectDeleteHook())
```

`AuditHook` 只填充操作人字段；ID 和时间由 schema default 生成。`RejectDeleteHook` 拦截物理删除，不会自动把查询或删除改写成软删除。
