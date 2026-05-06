package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	schemamixin "github.com/teamsillybees/initra/examples/api/internal/ent/schema/mixin"
)

// SysConfig 描述 sys_config 系统配置表。
type SysConfig struct {
	ent.Schema
}

// Mixin 返回系统配置表通用字段。
func (SysConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.ID{},
		schemamixin.SoftDelete{},
		schemamixin.Audit{},
	}
}

// Fields 返回系统配置表字段定义。
func (SysConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("config_key").MaxLen(128).NotEmpty().Unique(),
		field.Text("config_value").Default(""),
		field.Text("config_desc").Optional().Nillable(),
		field.Bool("is_builtin").Default(false),
		field.Int("sort_id").Default(0),
	}
}

// Annotations 返回系统配置表数据库元信息。
func (SysConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_config"},
	}
}
