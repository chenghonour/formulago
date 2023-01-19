/*
 * Copyright 2022 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"formulago/data/ent/schema/mixins"
)

// Logs holds the schema definition for the Logs entity.
type Logs struct {
	ent.Schema
}

// Fields of the Logs.
func (Logs) Fields() []ent.Field {
	return []ent.Field{
		field.String("type").Comment("type of log | 日志类型"),
		field.String("method").Comment("method of log | 日志请求方法"),
		field.String("api").Comment("api of log | 日志请求api"),
		field.Bool("success").Comment("success of log | 日志请求是否成功"),
		field.String("req_content").Optional().
			SchemaType(map[string]string{dialect.MySQL: "varchar(512)"}).Comment("content of request log | 日志请求内容"),
		field.String("resp_content").Optional().
			SchemaType(map[string]string{dialect.MySQL: "varchar(512)"}).Comment("content of response log | 日志返回内容"),
		field.String("ip").Optional().Comment("ip of log | 日志IP"),
		field.String("user_agent").Optional().Comment("user_agent of log | 日志用户客户端"),
		field.String("operator").Optional().Comment("operator of log | 日志操作者"),
		field.Int("time").Optional().Comment("time of log(millisecond) | 日志时间(毫秒)"),
	}
}

func (Logs) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

// Edges of the Logs.
func (Logs) Edges() []ent.Edge {
	return nil
}

func (Logs) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("api"),
	}
}

func (Logs) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_logs"},
		entsql.WithComments(true),
	}
}
