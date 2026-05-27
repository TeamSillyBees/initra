package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	entschemamixin "entgo.io/ent/schema/mixin"

	"github.com/teamsillybees/initra/pkg/idgen"
)

// ID 为业务表提供统一的雪花 ID 主键字段。
type ID struct {
	entschemamixin.Schema
}

// Fields 返回 ID mixin 的字段定义。
func (ID) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			GoType(idgen.ID(0)).
			DefaultFunc(idgen.NextID).
			Immutable().
			Positive().
			Comment("雪花算法生成的主键 ID。"),
	}
}
