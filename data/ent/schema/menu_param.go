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

	"formulago/data/ent/schema/mixins"
)

type MenuParam struct {
	ent.Schema
}

func (MenuParam) Fields() []ent.Field {
	return []ent.Field{
		field.String("type").Comment("pass parameters via params or query | 参数类型"),
		field.String("key").Comment("the key of parameters | 参数键"),
		field.String("value").Comment("the value of parameters | 参数值"),
	}
}

func (MenuParam) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

func (MenuParam) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("menus", Menu.Type).
			Ref("params").Unique(),
	}
}

func (MenuParam) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_menu_params"},
		entsql.WithComments(true),
	}
}
