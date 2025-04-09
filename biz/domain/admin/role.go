/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import "context"

type Role interface {
	Create(ctx context.Context, req RoleInfo) error
	Update(ctx context.Context, req RoleInfo) error
	Delete(ctx context.Context, id uint64) error
	RoleInfoByID(ctx context.Context, ID uint64) (roleInfo *RoleInfo, err error)
	List(ctx context.Context, req *RoleListReq) (roleInfoList []*RoleInfo, total int, err error)
	UpdateStatus(ctx context.Context, ID uint64, status uint8) error
}

type RoleInfo struct {
	ID            uint64
	Name          string
	Value         string
	DefaultRouter string
	Status        uint64
	Remark        string
	OrderNo       uint32
	CreatedAt     string
	UpdatedAt     string
}

type RoleListReq struct {
	Page     uint64
	PageSize uint64
}
