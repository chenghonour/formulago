/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import (
	"context"
	"errors"
	"fmt"
	"formulago/biz/domain/admin"
	"formulago/pkg/times"
	"formulago/data"
	"formulago/data/ent"
	"formulago/data/ent/role"
	"formulago/data/ent/user"
	"strconv"
	"time"
)

type Role struct {
	Data *data.Data
}

func NewRole(data *data.Data) admin.Role {
	return &Role{
		Data: data,
	}
}

func (r *Role) Create(ctx context.Context, req admin.RoleInfo) error {
	roleEnt, err := r.Data.DBClient.Role.Create().
		SetName(req.Name).
		SetValue(req.Value).
		SetDefaultRouter(req.DefaultRouter).
		SetStatus(uint8(req.Status)).
		SetRemark(req.Remark).
		SetOrderNo(req.OrderNo).
		Save(ctx)
	if err != nil {
		err = fmt.Errorf("create Role failed: %w", err)
		return err
	}

	// set roleEnt to cache
	r.Data.Cache.Set("roleData"+strconv.Itoa(int(roleEnt.ID)), roleEnt, 24*time.Hour)
	return nil
}

func (r *Role) Update(ctx context.Context, req admin.RoleInfo) error {
	roleEnt, err := r.Data.DBClient.Role.UpdateOneID(req.ID).
		SetName(req.Name).
		SetValue(req.Value).
		SetDefaultRouter(req.DefaultRouter).
		SetStatus(uint8(req.Status)).
		SetRemark(req.Remark).
		SetOrderNo(req.OrderNo).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	if err != nil {
		err = fmt.Errorf("update Role failed: %w", err)
		return err
	}

	// set roleEnt to cache
	r.Data.Cache.Set("roleData"+strconv.Itoa(int(roleEnt.ID)), roleEnt, 24*time.Hour)
	return nil
}

func (r *Role) Delete(ctx context.Context, id uint64) error {
	// whether role is used by user
	exist, err := r.Data.DBClient.User.Query().Where(user.RoleIDEQ(id)).Exist(ctx)
	if err != nil {
		err = fmt.Errorf("query user - role failed: %w", err)
		return err
	}
	if exist {
		return errors.New("role is used by user")
	}
	// delete role from db
	err = r.Data.DBClient.Role.DeleteOneID(id).Exec(ctx)
	if err != nil {
		err = fmt.Errorf("delete Role failed: %w", err)
		return err
	}
	// delete role from cache
	r.Data.Cache.Delete("roleData" + strconv.Itoa(int(id)))
	return nil
}

func (r *Role) RoleInfoByID(ctx context.Context, ID uint64) (roleInfo *admin.RoleInfo, err error) {
	// get role from cache
	roleInterface, ok := r.Data.Cache.Get("roleData" + strconv.Itoa(int(ID)))
	if ok {
		if r, ok := roleInterface.(*ent.Role); ok {
			return &admin.RoleInfo{
				ID:            r.ID,
				Name:          r.Name,
				Value:         r.Value,
				DefaultRouter: r.DefaultRouter,
				Status:        uint64(r.Status),
				Remark:        r.Remark,
				OrderNo:       r.OrderNo,
				CreatedAt:     r.CreatedAt.Format(times.TimeFormat),
				UpdatedAt:     r.UpdatedAt.Format(times.TimeFormat),
			}, nil
		}
	}
	// get role from db
	roleEnt, err := r.Data.DBClient.Role.Query().Where(role.IDEQ(ID)).Only(ctx)
	if err != nil {
		err = fmt.Errorf("get Role failed: %w", err)
		return nil, err
	}
	// set role to cache
	r.Data.Cache.Set("roleData"+strconv.Itoa(int(ID)), roleEnt, 24*time.Hour)
	// convert to RoleInfo
	roleInfo = &admin.RoleInfo{
		ID:            roleEnt.ID,
		Name:          roleEnt.Name,
		Value:         roleEnt.Value,
		DefaultRouter: roleEnt.DefaultRouter,
		Status:        uint64(roleEnt.Status),
		Remark:        roleEnt.Remark,
		OrderNo:       roleEnt.OrderNo,
		CreatedAt:     roleEnt.CreatedAt.Format(times.TimeFormat),
		UpdatedAt:     roleEnt.UpdatedAt.Format(times.TimeFormat),
	}
	return
}

func (r *Role) List(ctx context.Context, req *admin.RoleListReq) (roleInfoList []*admin.RoleInfo, total int, err error) {
	roleEntList, err := r.Data.DBClient.Role.Query().Order(ent.Asc(role.FieldOrderNo)).
		Offset(int(req.Page-1) * int(req.PageSize)).
		Limit(int(req.PageSize)).All(ctx)
	if err != nil {
		err = fmt.Errorf("get RoleList failed: %w", err)
		return nil, 0, err
	}
	// convert to List
	for _, roleEnt := range roleEntList {
		roleInfoList = append(roleInfoList, &admin.RoleInfo{
			ID:            roleEnt.ID,
			Name:          roleEnt.Name,
			Value:         roleEnt.Value,
			DefaultRouter: roleEnt.DefaultRouter,
			Status:        uint64(roleEnt.Status),
			Remark:        roleEnt.Remark,
			OrderNo:       roleEnt.OrderNo,
			CreatedAt:     roleEnt.CreatedAt.Format(times.TimeFormat),
			UpdatedAt:     roleEnt.UpdatedAt.Format(times.TimeFormat),
		})
	}
	total, err = r.Data.DBClient.Role.Query().Count(ctx)
	if err != nil {
		err = fmt.Errorf("count role failed: %w", err)
		return
	}
	return
}

// UpdateStatus update role's status, 0: disable, 1: enable
func (r *Role) UpdateStatus(ctx context.Context, ID uint64, status uint8) error {
	roleEnt, err := r.Data.DBClient.Role.UpdateOneID(ID).SetStatus(status).Save(ctx)
	if err != nil {
		err = fmt.Errorf("update Role status failed: %w", err)
		return err
	}
	// set role to cache
	r.Data.Cache.Set("roleData"+strconv.Itoa(int(ID)), roleEnt, 24*time.Hour)

	return nil
}
