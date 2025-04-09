/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import (
	"context"

	"formulago/biz/domain/admin"
	"formulago/data"
	"formulago/data/ent/role"

	"github.com/casbin/casbin/v2"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/pkg/errors"
)

type Authority struct {
	Cbs  *casbin.Enforcer
	Data *data.Data
}

func NewAuthority(data *data.Data, cbs *casbin.Enforcer) admin.Authority {
	return &Authority{
		Cbs:  cbs,
		Data: data,
	}
}

func (a *Authority) UpdateApiAuthority(ctx context.Context, roleIDStr string, infos []*admin.ApiAuthorityInfo) error {
	// clear old policies
	oldPolicies, err := a.Cbs.GetFilteredPolicy(0, roleIDStr)
	if err != nil {
		return err
	}
	if len(oldPolicies) != 0 {
		removeResult, err := a.Cbs.RemoveFilteredPolicy(0, roleIDStr)
		if err != nil {
			return err
		}
		if !removeResult {
			return errors.New("casbin policies remove failed")
		}
	}
	// add new policies
	var policies [][]string
	for _, v := range infos {
		policies = append(policies, []string{roleIDStr, v.Path, v.Method})
	}
	addResult, err := a.Cbs.AddPolicies(policies)
	if err != nil {
		return err
	}
	if !addResult {
		return errors.New("casbin policies add failed")
	}
	return nil
}

// ApiAuthority CompanyInfo Api authority policy by role id
func (a *Authority) ApiAuthority(ctx context.Context, roleIDStr string) (infos []*admin.ApiAuthorityInfo, err error) {
	policies, err := a.Cbs.GetFilteredPolicy(0, roleIDStr)
	if err != nil {
		return
	}
	for _, v := range policies {
		infos = append(infos, &admin.ApiAuthorityInfo{
			Path:   v[1],
			Method: v[2],
		})
	}
	return
}

func (a *Authority) UpdateMenuAuthority(ctx context.Context, roleID uint64, menuIDs []uint64) error {
	tx, err := a.Data.DBClient.Tx(ctx)
	if err != nil {
		return errors.Wrap(err, "starting a transaction err")
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				hlog.Error("UpdateMenuAuthority err:", err, "rollback err:", rollbackErr)
			}
		}
	}()

	err = tx.Role.UpdateOneID(roleID).ClearMenus().Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "delete role's menu failed, error")
	}

	err = tx.Role.UpdateOneID(roleID).AddMenuIDs(menuIDs...).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "add role's menu failed, error")
	}

	return tx.Commit()
}

func (a *Authority) MenuAuthority(ctx context.Context, roleID uint64) (menuIDs []uint64, err error) {
	menus, err := a.Data.DBClient.Role.Query().Where(role.IDEQ(roleID)).QueryMenus().All(ctx)
	for _, v := range menus {
		if v.ID != 1 {
			menuIDs = append(menuIDs, v.ID)
		}
	}
	return
}
