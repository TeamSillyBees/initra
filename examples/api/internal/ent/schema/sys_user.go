package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	schemamixin "github.com/teamsillybees/initra/examples/api/internal/ent/schema/mixin"
)

// SysUser 描述 sys_user 系统用户表。
type SysUser struct {
	ent.Schema
}

// Mixin 返回系统用户表通用字段。
func (SysUser) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.ID{},
		schemamixin.SoftDelete{},
		schemamixin.Audit{},
	}
}

// Fields 返回系统用户表字段定义。
func (SysUser) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").MaxLen(64).NotEmpty().Unique(),
		field.Text("password_hash").NotEmpty(),
		field.String("nickname").MaxLen(128).Optional().Nillable(),
		field.String("phone").MaxLen(32).Optional().Nillable().Unique(),
		field.String("email").MaxLen(255).Optional().Nillable().Unique(),
		field.Text("avatar_url").Optional().Nillable(),
		field.Bool("is_super_admin").Default(false),
		field.Bool("is_enable").Default(true),
		field.Int("sort_id").Default(0),
	}
}

// Edges 返回系统用户表关系定义。
func (SysUser) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user_roles", SysUserRole.Type),
	}
}

// Indexes 返回系统用户表索引定义。
func (SysUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_enable"),
	}
}

// Annotations 返回系统用户表数据库元信息。
func (SysUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_user"},
	}
}
