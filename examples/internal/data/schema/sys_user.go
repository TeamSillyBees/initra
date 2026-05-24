package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	schemamixin "github.com/teamsillybees/initra/examples/internal/data/schema/mixin"
)

// SysUser 系统后台用户表，用于后台登录、审计和权限归属。
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
		field.String("username").MaxLen(64).NotEmpty().Unique().
			Comment("登录用户名，全局唯一。"),
		field.Text("password_hash").NotEmpty().
			Comment("经过安全哈希后的密码密文。"),
		field.String("nickname").MaxLen(128).Optional().Nillable().
			Comment("用户昵称或显示名。"),
		field.String("phone").MaxLen(32).Optional().Nillable().Unique().
			Comment("手机号，可用于登录或通知。"),
		field.String("email").MaxLen(255).Optional().Nillable().Unique().
			Comment("邮箱地址，可用于找回密码和通知。"),
		field.Text("avatar_url").Optional().Nillable().
			Comment("头像资源地址。"),
		field.Bool("is_super_admin").Default(false).
			Comment("是否为超级管理员，超级管理员通常拥有全量权限。"),
		field.Bool("is_enable").Default(true).
			Comment("账号是否启用。"),
		field.Int("sort_id").Default(0).
			Comment("排序值，便于后台列表定制顺序。"),
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
		entsql.WithComments(true),
		schema.Comment("系统后台用户表，用于后台登录、审计和权限归属。"),
	}
}
