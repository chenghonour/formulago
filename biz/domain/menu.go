/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package domain

import "context"

type Menu interface {
	Create(ctx context.Context, menuReq *MenuInfo) error
	Update(ctx context.Context, menuReq *MenuInfo) error
	Delete(ctx context.Context, id uint64) error
	ListByRole(ctx context.Context, roleID uint64) (list []*MenuInfoTree, total uint64, err error)
	List(ctx context.Context, req *MenuListReq) (list []*MenuInfoTree, total int, err error)
	CreateMenuParam(ctx context.Context, req *MenuParam) error
	UpdateMenuParam(ctx context.Context, req *MenuParam) error
	DeleteMenuParam(ctx context.Context, menuParamID uint64) error
	MenuParamListByMenuID(ctx context.Context, menuID uint64) (list []MenuParam, total uint64, err error)
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
	Page     uint64
	PageSize uint64
}

type MenuInfoTree struct {
	MenuInfo
	CreatedAt string
	UpdatedAt string
	Children  []*MenuInfoTree
}

// MenuParam is the menu parameter structure.data stored at the table `sys_menu_params`
type MenuParam struct {
	ID        uint64
	MenuID    uint64
	Type      string
	Key       string
	Value     string
	CreatedAt string
	UpdatedAt string
}
