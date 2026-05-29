package fieldx

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// OptimisticLock 返回乐观锁版本字段。
func OptimisticLock() ent.Field {
	return field.Int32("version").
		Default(1).
		Positive().
		Comment("乐观锁版本号。")
}
