/*
 * Copyright 2022 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"formulago/data/ent/schema/mixins"
)

type Token struct {
	ent.Schema
}

func (Token) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id"),
		field.Uint64("user_id").Comment(" User's ID | 用户的ID").Unique(),
		field.String("token").Comment("Token string | Token 字符串"),
		field.String("source").Comment("Log in source such as GitHub | Token 来源 （本地为core, 第三方如github等）"),
		field.Time("expired_at").Comment(" Expire time | 过期时间"),
	}
}

func (Token) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

func (Token) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).Ref("token").Unique(),
	}
}

func (Token) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("expired_at"),
	}
}

func (Token) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_tokens"},
		entsql.WithComments(true),
	}
}
