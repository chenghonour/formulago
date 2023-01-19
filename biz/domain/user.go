/*
 * Copyright 2022 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package domain

import (
	"context"
	"time"
)

type User interface {
	Create(ctx context.Context, req CreateOrUpdateUserReq) error
	Update(ctx context.Context, req CreateOrUpdateUserReq) error
	ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error
	UserInfo(ctx context.Context, id uint64) (userInfo UserInfo, err error)
	List(ctx context.Context, req UserListReq) (userList []*UserInfo, total int, err error)
	UpdateUserStatus(ctx context.Context, id uint64, status uint64) error
	DeleteUser(ctx context.Context, id uint64) error
	UpdateProfile(ctx context.Context, req UpdateUserProfileReq) error
}

type CreateOrUpdateUserReq struct {
	ID       uint64
	Avatar   string
	RoleID   uint64
	Mobile   string
	Email    string
	Status   uint64
	Username string
	Password string
	Nickname string
}

type UserInfo struct {
	ID          uint64
	Status      uint8
	Username    string
	Password    string
	Nickname    string
	SideMode    string
	BaseColor   string
	ActiveColor string
	RoleID      uint64
	Mobile      string
	Email       string
	Wecom       string
	Avatar      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	RoleName    string
	RoleValue   string
}

type UserListReq struct {
	Page     uint64
	PageSize uint64
	Username string
	Nickname string
	Email    string
	Mobile   string
	RoleID   uint64
}

type UpdateUserProfileReq struct {
	ID       uint64
	Nickname string
	Email    string
	Mobile   string
	Avatar   string
}
