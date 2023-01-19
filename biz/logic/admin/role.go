/*
 * Copyright 2022 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import (
	"context"
	"formulago/biz/domain"
	"formulago/data"
	"formulago/data/ent"
	"formulago/data/ent/role"
	"formulago/data/ent/user"
	"github.com/cockroachdb/errors"
	"strconv"
	"time"
)

type Role struct {
	Data *data.Data
}

func NewRole(data *data.Data) domain.Role {
	return &Role{
		Data: data,
	}
}

func (r *Role) Create(ctx context.Context, req domain.RoleInfo) error {
	roleEnt, err := r.Data.DBClient.Role.Create().
		SetName(req.Name).
		SetValue(req.Value).
		SetDefaultRouter(req.DefaultRouter).
		SetStatus(uint8(req.Status)).
		SetRemark(req.Remark).
		SetOrderNo(req.OrderNo).
		Save(ctx)
	if err != nil {
		err = errors.Wrap(err, "create Role failed")
		return err
	}

	// set roleEnt to cache
	r.Data.Cache.Set("roleData"+strconv.Itoa(int(roleEnt.ID)), roleEnt, 24*time.Hour)
	return nil
}

func (r *Role) Update(ctx context.Context, req domain.RoleInfo) error {
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
		err = errors.Wrap(err, "update Role failed")
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
		err = errors.Wrap(err, "query user - role failed")
		return err
	}
	if exist {
		return errors.New("role is used by user")
	}
	// delete role from db
	err = r.Data.DBClient.Role.DeleteOneID(id).Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, "delete Role failed")
		return err
	}
	// delete role from cache
	r.Data.Cache.Delete("roleData" + strconv.Itoa(int(id)))
	return nil
}

func (r *Role) RoleInfoByID(ctx context.Context, ID uint64) (roleInfo *domain.RoleInfo, err error) {
	// get role from cache
	roleInterface, ok := r.Data.Cache.Get("roleData" + strconv.Itoa(int(ID)))
	if ok {
		if r, ok := roleInterface.(*ent.Role); ok {
			return &domain.RoleInfo{
				ID:            r.ID,
				Name:          r.Name,
				Value:         r.Value,
				DefaultRouter: r.DefaultRouter,
				Status:        uint64(r.Status),
				Remark:        r.Remark,
				OrderNo:       r.OrderNo,
				CreatedAt:     r.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt:     r.UpdatedAt.Format("2006-01-02 15:04:05"),
			}, nil
		}
	}
	// get role from db
	roleEnt, err := r.Data.DBClient.Role.Query().Where(role.IDEQ(ID)).Only(ctx)
	if err != nil {
		err = errors.Wrap(err, "get Role failed")
		return nil, err
	}
	// set role to cache
	r.Data.Cache.Set("roleData"+strconv.Itoa(int(ID)), roleEnt, 24*time.Hour)
	// convert to RoleInfo
	roleInfo = &domain.RoleInfo{
		ID:            roleEnt.ID,
		Name:          roleEnt.Name,
		Value:         roleEnt.Value,
		DefaultRouter: roleEnt.DefaultRouter,
		Status:        uint64(roleEnt.Status),
		Remark:        roleEnt.Remark,
		OrderNo:       roleEnt.OrderNo,
		CreatedAt:     roleEnt.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     roleEnt.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	return
}

func (r *Role) List(ctx context.Context, req *domain.RoleListReq) (roleInfoList []*domain.RoleInfo, total int, err error) {
	roleEntList, err := r.Data.DBClient.Role.Query().Order(ent.Asc(role.FieldOrderNo)).
		Offset(int(req.Page-1) * int(req.PageSize)).
		Limit(int(req.PageSize)).All(ctx)
	if err != nil {
		err = errors.Wrap(err, "get RoleList failed")
		return nil, 0, err
	}
	// convert to List
	for _, roleEnt := range roleEntList {
		roleInfoList = append(roleInfoList, &domain.RoleInfo{
			ID:            roleEnt.ID,
			Name:          roleEnt.Name,
			Value:         roleEnt.Value,
			DefaultRouter: roleEnt.DefaultRouter,
			Status:        uint64(roleEnt.Status),
			Remark:        roleEnt.Remark,
			OrderNo:       roleEnt.OrderNo,
			CreatedAt:     roleEnt.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     roleEnt.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	total, _ = r.Data.DBClient.Role.Query().Count(ctx)
	return
}

// UpdateStatus update role's status, 0: disable, 1: enable
func (r *Role) UpdateStatus(ctx context.Context, ID uint64, status uint8) error {
	roleEnt, err := r.Data.DBClient.Role.UpdateOneID(ID).SetStatus(status).Save(ctx)
	if err != nil {
		err = errors.Wrap(err, "update Role status failed")
		return err
	}
	// set role to cache
	r.Data.Cache.Set("roleData"+strconv.Itoa(int(ID)), roleEnt, 24*time.Hour)

	return nil
}
