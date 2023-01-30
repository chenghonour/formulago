/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import (
	"context"
	"fmt"
	"formulago/biz/domain"
	"formulago/data"
	"formulago/data/ent"
	"formulago/data/ent/menu"
	"formulago/data/ent/role"
	"github.com/cockroachdb/errors"
)

type Menu struct {
	Data *data.Data
}

func NewMenu(data *data.Data) domain.Menu {
	return &Menu{
		Data: data,
	}
}

func (m *Menu) Create(ctx context.Context, menuReq *domain.MenuInfo) error {
	// get menu level
	if menuReq.ParentID == 0 {
		// it is a first level menu
		menuReq.ParentID = 1
		menuReq.Level = 1
	} else {
		// it is a children level menu
		// get parent menu level
		parent, err := m.Data.DBClient.Menu.Query().Where(menu.IDEQ(menuReq.ParentID)).First(ctx)
		if err != nil {
			return errors.Wrap(err, "query menu failed")
		}
		// set menu level
		menuReq.Level = parent.MenuLevel + 1
	}

	// create menu
	err := m.Data.DBClient.Menu.Create().
		SetMenuLevel(menuReq.Level).
		SetMenuType(menuReq.MenuType).
		SetParentID(menuReq.ParentID).
		SetPath(menuReq.Path).
		SetName(menuReq.Name).
		SetRedirect(menuReq.Redirect).
		SetComponent(menuReq.Component).
		SetOrderNo(menuReq.OrderNo).
		SetDisabled(menuReq.Disabled).
		// meta
		SetTitle(menuReq.Meta.Title).
		SetIcon(menuReq.Meta.Icon).
		SetHideMenu(menuReq.Meta.HideMenu).
		SetHideBreadcrumb(menuReq.Meta.HideBreadcrumb).
		SetCurrentActiveMenu(menuReq.Meta.CurrentActiveMenu).
		SetIgnoreKeepAlive(menuReq.Meta.IgnoreKeepAlive).
		SetHideTab(menuReq.Meta.HideTab).
		SetFrameSrc(menuReq.Meta.FrameSrc).
		SetCarryParam(menuReq.Meta.CarryParam).
		SetHideChildrenInMenu(menuReq.Meta.HideChildrenInMenu).
		SetAffix(menuReq.Meta.Affix).
		SetDynamicLevel(menuReq.Meta.DynamicLevel).
		SetRealPath(menuReq.Meta.RealPath).
		Exec(ctx)

	if err != nil {
		return errors.Wrap(err, "create menu failed")
	}
	return nil
}

func (m *Menu) Update(ctx context.Context, menuReq *domain.MenuInfo) error {
	// get menu level
	if menuReq.ParentID == 0 {
		// it is a first level menu
		menuReq.ParentID = 1
		menuReq.Level = 1
	} else {
		// it is a children level menu
		// get parent menu level
		parent, err := m.Data.DBClient.Menu.Query().Where(menu.IDEQ(menuReq.ParentID)).First(ctx)
		if err != nil {
			return errors.Wrap(err, "query menu failed")
		}
		// set menu level
		menuReq.Level = parent.MenuLevel + 1
	}

	// update menu
	err := m.Data.DBClient.Menu.UpdateOneID(menuReq.ID).
		SetMenuLevel(menuReq.Level).
		SetMenuType(menuReq.MenuType).
		SetParentID(menuReq.ParentID).
		SetPath(menuReq.Path).
		SetName(menuReq.Name).
		SetRedirect(menuReq.Redirect).
		SetComponent(menuReq.Component).
		SetOrderNo(menuReq.OrderNo).
		SetDisabled(menuReq.Disabled).
		// meta
		SetTitle(menuReq.Meta.Title).
		SetIcon(menuReq.Meta.Icon).
		SetHideMenu(menuReq.Meta.HideMenu).
		SetHideBreadcrumb(menuReq.Meta.HideBreadcrumb).
		SetCurrentActiveMenu(menuReq.Meta.CurrentActiveMenu).
		SetIgnoreKeepAlive(menuReq.Meta.IgnoreKeepAlive).
		SetHideTab(menuReq.Meta.HideTab).
		SetFrameSrc(menuReq.Meta.FrameSrc).
		SetCarryParam(menuReq.Meta.CarryParam).
		SetHideChildrenInMenu(menuReq.Meta.HideChildrenInMenu).
		SetAffix(menuReq.Meta.Affix).
		SetDynamicLevel(menuReq.Meta.DynamicLevel).
		SetRealPath(menuReq.Meta.RealPath).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "update menu failed")
	}

	return nil
}

func (m *Menu) Delete(ctx context.Context, id uint64) error {
	// find out the menu whether it has children
	// if it has children, it can not be deleted
	exist, err := m.Data.DBClient.Menu.Query().Where(menu.ParentIDEQ(id)).Exist(ctx)
	if err != nil {
		return errors.Wrap(err, "query menu failed")
	}
	if exist {
		return errors.New("menu has children, can not be deleted")
	}

	// delete menu
	err = m.Data.DBClient.Menu.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "delete menu failed")
	}
	return nil
}

func (m *Menu) ListByRole(ctx context.Context, roleID uint64) (list []*domain.MenuInfoTree, total uint64, err error) {
	menus, err := m.Data.DBClient.Role.Query().Where(role.IDEQ(roleID)).
		QueryMenus().Order(ent.Asc(menu.FieldOrderNo)).All(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err, "query m by role failed")
	}

	list = findMenuChildren(menus, 1)
	total = uint64(len(list))
	return
}

func (m *Menu) List(ctx context.Context, req *domain.MenuListReq) (list []*domain.MenuInfoTree, total int, err error) {
	// query menu list
	menus, err := m.Data.DBClient.Menu.Query().Order(ent.Asc(menu.FieldOrderNo)).
		Offset(int(req.Page-1) * int(req.PageSize)).
		Limit(int(req.PageSize)).All(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err, "query menu list failed")
	}
	list = findMenuChildren(menus, 1)
	total, _ = m.Data.DBClient.Menu.Query().Count(ctx)
	return
}

func findMenuChildren(data []*ent.Menu, parentID uint64) []*domain.MenuInfoTree {
	if data == nil {
		return nil
	}
	var result []*domain.MenuInfoTree
	for _, v := range data {
		// discard the parent menu, only find the children menu
		if v.ParentID == parentID && v.ID != parentID {
			var m = new(domain.MenuInfoTree)
			m.ID = v.ID
			m.CreatedAt = v.CreatedAt.Format("2006-01-02 15:04:05")
			m.UpdatedAt = v.UpdatedAt.Format("2006-01-02 15:04:05")
			m.MenuType = v.MenuType
			m.Level = v.MenuLevel
			m.ParentID = v.ParentID
			m.Path = v.Path
			m.Name = v.Name
			m.Redirect = v.Redirect
			m.Component = v.Component
			m.OrderNo = v.OrderNo
			m.Meta = &domain.MenuMeta{
				Title:              v.Title,
				Icon:               v.Icon,
				HideMenu:           v.HideMenu,
				HideBreadcrumb:     v.HideBreadcrumb,
				CurrentActiveMenu:  v.CurrentActiveMenu,
				IgnoreKeepAlive:    v.IgnoreKeepAlive,
				HideTab:            v.HideTab,
				FrameSrc:           v.FrameSrc,
				CarryParam:         v.CarryParam,
				HideChildrenInMenu: v.HideChildrenInMenu,
				Affix:              v.Affix,
				DynamicLevel:       v.DynamicLevel,
				RealPath:           v.RealPath,
			}

			m.Children = findMenuChildren(data, v.ID)
			result = append(result, m)
		}
	}
	return result
}

func (m *Menu) CreateMenuParam(ctx context.Context, req *domain.MenuParam) error {
	// check menu whether exist
	exist, err := m.Data.DBClient.Menu.Query().Where(menu.IDEQ(req.MenuID)).Exist(ctx)
	if err != nil {
		return errors.Wrap(err, "query menu failed")
	}
	if !exist {
		return errors.New(fmt.Sprintf("menu not exist, menu id: %d", req.MenuID))
	}

	// create menu param
	err = m.Data.DBClient.MenuParam.Create().
		SetMenusID(req.MenuID).
		SetType(req.Type).
		SetKey(req.Key).
		SetValue(req.Value).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "create menu param failed")
	}
	return nil
}

func (m *Menu) UpdateMenuParam(ctx context.Context, req *domain.MenuParam) error {
	// check menu whether exist
	exist, err := m.Data.DBClient.Menu.Query().Where(menu.IDEQ(req.MenuID)).Exist(ctx)
	if err != nil {
		return errors.Wrap(err, "query menu failed")
	}
	if !exist {
		return errors.New(fmt.Sprintf("menu not exist, menu id: %d", req.MenuID))
	}

	// update menu param
	err = m.Data.DBClient.MenuParam.UpdateOneID(req.ID).
		SetMenusID(req.MenuID).
		SetType(req.Type).
		SetKey(req.Key).
		SetValue(req.Value).
		Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "update menu param failed")
	}
	return nil
}

func (m *Menu) DeleteMenuParam(ctx context.Context, menuParamID uint64) error {
	// delete menu param
	err := m.Data.DBClient.MenuParam.DeleteOneID(menuParamID).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "delete menu param failed")
	}
	return nil
}

func (m *Menu) MenuParamListByMenuID(ctx context.Context, menuID uint64) (list []domain.MenuParam, total uint64, err error) {
	// query menu param list
	params, err := m.Data.DBClient.Menu.Query().Where(menu.IDEQ(menuID)).QueryParams().All(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err, "query menu param list failed")
	}

	// convert to MenuParam
	for _, v := range params {
		var p domain.MenuParam
		p.ID = v.ID
		p.Type = v.Type
		p.Key = v.Key
		p.Value = v.Value
		p.CreatedAt = v.CreatedAt.Format("2006-01-02 15:04:05")
		p.UpdatedAt = v.UpdatedAt.Format("2006-01-02 15:04:05")
		list = append(list, p)
	}

	total = uint64(len(list))
	return
}
