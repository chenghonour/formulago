/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package domain

import "context"

type Token interface {
	Create(ctx context.Context, req *TokenInfo) error
	Update(ctx context.Context, req *TokenInfo) error
	IsExistByUserID(ctx context.Context, userID uint64) bool
	Delete(ctx context.Context, userID uint64) error
	List(ctx context.Context, req *TokenListReq) (res []*TokenInfo, total int, err error)
}

type TokenInfo struct {
	ID        uint64
	CreatedAt string
	UpdatedAt string
	UserID    uint64
	UserName  string
	Token     string
	Source    string
	ExpiredAt string
}

type TokenListReq struct {
	Page     uint64
	PageSize uint64
	Username string
	UserID   uint64
}
