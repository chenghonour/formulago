/*
 * Copyright 2023 FormulaGo Authors
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

type DictionaryDetail struct {
	ent.Schema
}

func (DictionaryDetail) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").Comment("the title shown in the ui | 展示名称 （建议配合i18n）"),
		field.String("key").Comment("key | 键"),
		field.String("value").Comment("value | 值"),
		field.Uint64("dictionary_id").Optional().Comment("Dictionary ID | 字典ID"),
	}
}

func (DictionaryDetail) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.StatusMixin{},
	}
}

func (DictionaryDetail) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("dictionary", Dictionary.Type).Field("dictionary_id").Ref("dictionary_details").Unique(),
	}
}

func (DictionaryDetail) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("key", "dictionary_id").Unique(),
	}
}

func (DictionaryDetail) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_dictionary_details"},
		entsql.WithComments(true),
	}
}
