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
	"entgo.io/ent/schema/field"

	"formulago/data/ent/schema/mixins"
)

type OauthProvider struct {
	ent.Schema
}

func (OauthProvider) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique().Comment("the provider's name | 提供商名称"),
		field.String("app_id").Optional().Comment("the app id | 应用id"),
		field.String("client_id").Comment("the client id | 客户端 id"),
		field.String("client_secret").Optional().Comment("the client secret | 客户端密钥"),
		field.String("redirect_url").Comment("the redirect url | 跳转地址"),
		field.String("scopes").Optional().Comment("the scopes | 权限范围"),
		field.String("auth_url").Comment("the auth url of the provider | 认证地址"),
		field.String("token_url").Optional().Comment("the token url of the provider | 获取 token地址"),
		field.Uint64("auth_style").Optional().Comment("the auth style, 0: auto detect; 1: third party login; 2: login with username and password"),
		field.String("info_url").Optional().Comment("the URL to request user information by token | 用户信息请求地址"),
	}
}

func (OauthProvider) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

func (OauthProvider) Edges() []ent.Edge {
	return nil
}

func (OauthProvider) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sys_oauth_providers"},
		entsql.WithComments(true),
	}
}
