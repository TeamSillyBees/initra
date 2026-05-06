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

// SysRoleMenu 系统角色与菜单/按钮资源关系表，用于角色授权。
type SysRoleMenu struct {
	ent.Schema
}

// Mixin 返回角色菜单关系表通用字段。
func (SysRoleMenu) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.ID{},
		schemamixin.SoftDelete{},
	}
}

// Fields 返回角色菜单关系表字段定义。
func (SysRoleMenu) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("role_id").Positive().
			Comment("系统角色 ID，关联 sys_role.id。"),
		field.Int64("menu_id").Positive().
			Comment("系统菜单资源 ID，关联 sys_menu.id。"),
		field.Int64("created_by").Optional().Nillable().
			Comment("创建人用户 ID。"),
		field.Time("created_at").Immutable().
			Comment("创建时间。"),
	}
}

// Edges 返回角色菜单关系表关系定义。
func (SysRoleMenu) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("role", SysRole.Type).
			Ref("role_menus").
			Field("role_id").
			Required().
			Unique(),
		edge.From("menu", SysMenu.Type).
			Ref("role_menus").
			Field("menu_id").
			Required().
			Unique(),
	}
}

// Indexes 返回角色菜单关系表索引定义。
func (SysRoleMenu) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_id"),
		index.Fields("menu_id"),
		index.Fields("role_id", "menu_id").Unique(),
	}
}

// Annotations 返回角色菜单关系表数据库元信息。
func (SysRoleMenu) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_role_menu"},
	}
}
