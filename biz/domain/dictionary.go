/*
 * Copyright 2022 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package domain

import "context"

type Dictionary interface {
	Create(ctx context.Context, req *DictionaryInfo) error
	Update(ctx context.Context, req *DictionaryInfo) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, req *DictListReq) (list []*DictionaryInfo, total int, err error)
	CreateDetail(ctx context.Context, req *DictionaryDetail) error
	UpdateDetail(ctx context.Context, req *DictionaryDetail) error
	DeleteDetail(ctx context.Context, id uint64) error
	DetailListByDictName(ctx context.Context, dictName string) (list []*DictionaryDetail, total uint64, err error)
	DetailByDictNameAndKey(ctx context.Context, dictName, key string) (detail *DictionaryDetail, err error)
}

type DictionaryInfo struct {
	ID          uint64
	Title       string
	Name        string
	Status      uint64
	Description string
	CreatedAt   string
	UpdatedAt   string
}

type DictListReq struct {
	Title    string
	Name     string
	Page     uint64
	PageSize uint64
}

type DictionaryDetail struct {
	ID        uint64
	Title     string
	Key       string
	Value     string
	Status    uint64
	CreatedAt string
	UpdatedAt string
	ParentID  uint64
}
