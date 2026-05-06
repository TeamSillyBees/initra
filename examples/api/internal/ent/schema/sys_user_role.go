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

// SysUserRole 描述 sys_user_role 用户角色关系表。
type SysUserRole struct {
	ent.Schema
}

// Mixin 返回用户角色关系表通用字段。
func (SysUserRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.ID{},
		schemamixin.SoftDelete{},
	}
}

// Fields 返回用户角色关系表字段定义。
func (SysUserRole) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id").Positive(),
		field.Int64("role_id").Positive(),
		field.Int64("created_by").Optional().Nillable(),
		field.Time("created_at").Immutable(),
	}
}

// Edges 返回用户角色关系表关系定义。
func (SysUserRole) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", SysUser.Type).
			Ref("user_roles").
			Field("user_id").
			Required().
			Unique(),
		edge.From("role", SysRole.Type).
			Ref("user_roles").
			Field("role_id").
			Required().
			Unique(),
	}
}

// Indexes 返回用户角色关系表索引定义。
func (SysUserRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("role_id"),
		index.Fields("user_id", "role_id").Unique(),
	}
}

// Annotations 返回用户角色关系表数据库元信息。
func (SysUserRole) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_user_role"},
	}
}
