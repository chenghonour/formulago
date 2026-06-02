/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import "context"

type Menu interface {
	Create(ctx context.Context, menuReq *MenuInfo) error
	Update(ctx context.Context, menuReq *MenuInfo) error
	Delete(ctx context.Context, id uint64) error
	ListByRole(ctx context.Context, roleID uint64) (list []*MenuInfoTree, total uint64, err error)
	List(ctx context.Context, req *MenuListReq) (list []*MenuInfoTree, total int, err error)
}

type MenuInfo struct {
	Level     uint32
	ParentID  uint64
	Path      string
	Name      string
	Redirect  string
	Component string
	OrderNo   uint32
	Disabled  bool
	Meta      *MenuMeta
	ID        uint64
	MenuType  uint32
}

type MenuMeta struct {
	Title              string
	Icon               string
	HideMenu           bool
	HideBreadcrumb     bool
	CurrentActiveMenu  string
	IgnoreKeepAlive    bool
	HideTab            bool
	FrameSrc           string
	CarryParam         bool
	HideChildrenInMenu bool
	Affix              bool
	DynamicLevel       uint32
	RealPath           string
}

type MenuListReq struct {
	Name     string
	Title    string
	Page     uint64
	PageSize uint64
}

type MenuInfoTree struct {
	MenuInfo
	CreatedAt string
	UpdatedAt string
	Children  []*MenuInfoTree
}
