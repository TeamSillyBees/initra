package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Audit 为常规业务表提供创建与更新审计字段。
type Audit struct {
	mixin.Schema
}

// Fields 返回审计字段定义。
func (Audit) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Immutable().
			Comment("创建时间。"),
		field.Time("updated_at").
			Comment("最后更新时间。"),
		field.Int64("created_by").Optional().Nillable().
			Comment("创建人用户 ID。"),
		field.Int64("updated_by").Optional().Nillable().
			Comment("最后更新人用户 ID。"),
	}
}
