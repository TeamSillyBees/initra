package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	entschemamixin "entgo.io/ent/schema/mixin"
)

// SoftDelete 为支持逻辑删除的表提供 deleted_at 字段。
type SoftDelete struct {
	entschemamixin.Schema
}

// Fields 返回逻辑删除字段定义。
func (SoftDelete) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_at").Optional().Nillable().
			Comment("逻辑删除时间，NULL 表示未删除。"),
	}
}

// Indexes 返回逻辑删除字段索引。
func (SoftDelete) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("deleted_at"),
	}
}
