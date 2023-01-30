/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"formulago/data/ent/schema/mixins"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").Unique().Comment("user's login name | 登录名"),
		field.String("password").Comment("password | 密码"),
		field.String("nickname").Unique().Comment("nickname | 昵称"),
		field.String("side_mode").Optional().Default("dark").Comment("template mode | 布局方式"),
		field.String("base_color").Optional().Default("#fff").Comment("base color of template | 后台页面色调"),
		field.String("active_color").Optional().Default("#1890ff").Comment("active color of template | 当前激活的颜色设定"),
		field.Uint64("role_id").Optional().Default(2).Comment("role id | 角色ID"),
		field.String("mobile").Unique().Comment("mobile number | 手机号"),
		field.String("email").Optional().Comment("email | 邮箱号"),
		field.String("wecom").Optional().Comment("wecom | 企业微信号"),
		field.String("avatar").
			SchemaType(map[string]string{dialect.MySQL: "varchar(512)"}).
			Optional().
			Default("").
			Comment("avatar | 头像路径"),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.StatusMixin{},
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("token", Token.Type).Unique(),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username", "email").
			Unique(),
	}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_users"},
		entsql.WithComments(true),
	}
}
