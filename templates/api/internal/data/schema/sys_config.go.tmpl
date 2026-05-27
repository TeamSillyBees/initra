package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	entxmixin "github.com/teamsillybees/initra/pkg/entx/mixin"
)

// SysConfig 系统配置表，用于集中存放可在后台维护的运行时配置。
type SysConfig struct {
	ent.Schema
}

// Mixin 返回系统配置表通用字段。
func (SysConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		entxmixin.ID{},
		entxmixin.SoftDelete{},
		entxmixin.Audit{},
	}
}

// Fields 返回系统配置表字段定义。
func (SysConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("config_key").MaxLen(128).NotEmpty().Unique().
			Comment("配置键，程序通过该键读取配置。"),
		field.Text("config_value").Default("").
			Comment("配置值。"),
		field.Text("config_desc").Optional().Nillable().
			Comment("配置项描述。"),
		field.Bool("is_builtin").Default(false).
			Comment("是否为系统内置配置，内置配置通常不允许删除。"),
		field.Int("sort_id").Default(0).
			Comment("排序值。"),
	}
}

// Annotations 返回系统配置表数据库元信息。
func (SysConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_config"},
		entsql.WithComments(true),
		schema.Comment("系统配置表，用于集中存放可在后台维护的运行时配置。"),
	}
}
