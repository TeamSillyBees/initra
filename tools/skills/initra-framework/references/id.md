# 业务 ID 与 Ent Mixin

## 规则

- 业务 ID 统一使用 `github.com/teamsillybees/initra/pkg/idgen.ID`。
- Ent 主键、外键、auth user ID、REST path 参数、service 入参和 JSON VO 都使用 `idgen.ID`。
- 对外 JSON/OpenAPI ID 是字符串；Huma 示例使用 `"1771234567890123456"`。
- 只有雪花生成器、第三方库、手写 SQL 参数或底层基础设施才调用 `Int64()`。

## Ent Schema

复用 `pkg/entx/mixin`：

```go
func (SysUser) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.ID{},
		mixin.Audit{},
		mixin.SoftDelete{},
		mixin.OptimisticLock{},
	}
}
```

外键字段使用 Ent 的 `GoType(idgen.ID(0))`：

```go
field.Int64("role_id").GoType(idgen.ID(0)).Positive()
```

## 禁止

- 不要在业务项目复制本地 schema mixin。
- 不要把 ID 在 service/handler 层转成 `string` 或 `int64` 传递。
- 不要把 `idgen.ID` 的零值当作有效 ID；入口处要校验 `<= 0`。
