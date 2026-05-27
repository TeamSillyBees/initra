package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	entschemamixin "entgo.io/ent/schema/mixin"
)

// OptimisticLock 为高并发更新表提供乐观锁版本字段。
type OptimisticLock struct {
	entschemamixin.Schema
}

// Fields 返回乐观锁字段定义。
func (OptimisticLock) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("version").Default(1).
			Comment("乐观锁版本号。"),
	}
}
