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

// SysRole 描述 sys_role 系统角色表。
type SysRole struct {
	ent.Schema
}

// Mixin 返回系统角色表通用字段。
func (SysRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.ID{},
		schemamixin.SoftDelete{},
		schemamixin.Audit{},
	}
}

// Fields 返回系统角色表字段定义。
func (SysRole) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").MaxLen(64).NotEmpty().Unique(),
		field.String("name").MaxLen(128).NotEmpty(),
		field.Text("remark").Optional().Nillable(),
		field.Bool("is_builtin").Default(false),
		field.Bool("is_enable").Default(true),
		field.Int("sort_id").Default(0),
	}
}

// Edges 返回系统角色表关系定义。
func (SysRole) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user_roles", SysUserRole.Type),
		edge.To("role_menus", SysRoleMenu.Type),
	}
}

// Indexes 返回系统角色表索引定义。
func (SysRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_enable"),
	}
}

// Annotations 返回系统角色表数据库元信息。
func (SysRole) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_role"},
	}
}
