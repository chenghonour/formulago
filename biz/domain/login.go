/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package domain

import "context"

type Login interface {
	Login(ctx context.Context, username, password string) (res *LoginResp, err error)
	LoginByOAuth(ctx context.Context, provider, credential string) (res *LoginResp, err error)
}

type LoginResp struct {
	UserID    uint64
	Username  string
	RoleName  string
	RoleValue string
	RoleID    uint64
}
