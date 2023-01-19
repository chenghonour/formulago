/*
 * Copyright 2022 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

// Package domain provide domain/model/entity and related interfaces

package domain

import "context"

type Api interface {
	Create(ctx context.Context, req ApiInfo) error
	Update(ctx context.Context, req ApiInfo) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, req ListApiReq) (resp []*ApiInfo, total int, err error)
}

type ApiInfo struct {
	ID          uint64
	CreatedAt   string
	UpdatedAt   string
	Path        string
	Description string
	Group       string
	Method      string
}

type ListApiReq struct {
	Page        uint64
	PageSize    uint64
	Path        string
	Description string
	Method      string
	Group       string
}
