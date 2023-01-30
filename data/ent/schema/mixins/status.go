/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package mixins

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// StatusMixin implements the ent.Mixin for sharing
// status fields with package schemas.
type StatusMixin struct {
	mixin.Schema
}

func (StatusMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Uint8("status").
			Default(1).
			Optional().
			Comment("status 1 normal 0 ban | 状态 1 正常 0 禁用"),
	}
}
