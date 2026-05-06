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

// SysMenu 描述 sys_menu 系统菜单与权限资源表。
type SysMenu struct {
	ent.Schema
}

// Mixin 返回系统菜单表通用字段。
func (SysMenu) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.ID{},
		schemamixin.SoftDelete{},
		schemamixin.Audit{},
	}
}

// Fields 返回系统菜单表字段定义。
func (SysMenu) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("parent_id").Default(0),
		field.String("app_id").MaxLen(64).Optional().Nillable(),
		field.String("title").MaxLen(128).NotEmpty(),
		field.Int16("menu_type"),
		field.Text("route_path").Optional().Nillable(),
		field.Text("component_path").Optional().Nillable(),
		field.String("permission_code").MaxLen(128).Optional().Nillable().Unique(),
		field.String("icon").MaxLen(128).Optional().Nillable(),
		field.Bool("is_visible").Default(true),
		field.Bool("is_cached").Default(true),
		field.Int("sort_id").Default(0),
	}
}

// Edges 返回系统菜单表关系定义。
func (SysMenu) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("role_menus", SysRoleMenu.Type),
	}
}

// Indexes 返回系统菜单表索引定义。
func (SysMenu) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parent_id"),
	}
}

// Annotations 返回系统菜单表数据库元信息。
func (SysMenu) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_menu"},
	}
}
