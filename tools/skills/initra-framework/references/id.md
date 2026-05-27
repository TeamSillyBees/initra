# 业务 ID 与 Ent Mixin

业务 ID 使用 `github.com/teamsillybees/initra/pkg/idgen.ID`。该类型底层为 int64，JSON、Text 和 Huma OpenAPI 统一表现为字符串，避免浏览器在安全整数范围外丢失精度。

## 使用规则

- Ent 主键、外键、auth user ID、REST path params、service/repo 入参、领域模型和 JSON VO 都使用 `idgen.ID`。
- HTTP/OpenAPI ID schema 应为 `type: string`，pattern 为 `^[1-9][0-9]{0,18}$`，example 为 `"1771234567890123456"`。
- 新建 Ent schema 时复用 `pkg/entx/mixin`，不要在业务项目中复制本地 mixin。
- `idgen.Register(injector, cfg.IDGen.Node)` 在 Ent client 创建前调用，用于配置默认雪花 ID 生成器。

## Ent Schema 示例

```go
import (
	entxmixin "github.com/teamsillybees/initra/pkg/entx/mixin"
	"github.com/teamsillybees/initra/pkg/idgen"
)

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entxmixin.ID{},
		entxmixin.SoftDelete{},
		entxmixin.Audit{},
	}
}

func (UserRole) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id").GoType(idgen.ID(0)).Positive(),
		field.Int64("role_id").GoType(idgen.ID(0)).Positive(),
	}
}
```

## 允许 Int64 的场景

只在以下场景调用 `Int64()`：

- 对接雪花 ID 生成器或底层 ID 基础设施。
- 对接只接受 `int64` 的第三方库。
- 手写 SQL 参数必须传入 `int64`。
- 极少数底层基础设施需要原始数值。

业务 service、handler、repository 签名不要退回 `int64` 或 `string`。
