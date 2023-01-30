/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import (
	"context"
	"formulago/biz/domain"
	"formulago/data"
	"formulago/pkg/encrypt"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cockroachdb/errors"

	"formulago/data/ent"
	"formulago/data/ent/user"
)

type Login struct {
	Data *data.Data
}

func NewLogin(data *data.Data) domain.Login {
	return &Login{
		Data: data,
	}
}

func (l *Login) Login(ctx context.Context, username, password string) (res *domain.LoginResp, err error) {
	// check username
	result, err := l.Data.DBClient.User.Query().Where(user.UsernameEQ(username)).First(ctx)
	if err != nil {
		return nil, err
	}
	if result == nil {
		err = errors.New("login user not exist")
		return nil, err
	}
	// check password
	if ok := encrypt.BcryptCheck(password, result.Password); !ok {
		err = errors.New("wrong password")
		return nil, err
	}
	res = new(domain.LoginResp)
	res.Username = username
	res.UserID = result.ID
	// get role info
	res.RoleID = result.RoleID
	res.RoleName, res.RoleValue, err = l.getRoleInfo(ctx, result.RoleID)

	return
}

// LoginByOAuth only use for oauth2.0 login, should have been Authenticated by oauth
func (l *Login) LoginByOAuth(ctx context.Context, provider, credential string) (res *domain.LoginResp, err error) {
	// check credential
	var userInfo *ent.User
	switch provider {
	case "wecom":
		// check wecom
		var err error
		userInfo, err = l.Data.DBClient.User.Query().Where(user.Wecom(credential)).Only(ctx)
		if err != nil {
			return nil, err
		}
		if userInfo == nil {
			err = errors.New("login user not exist, find by wecom")
			return nil, err
		}
	default:
		// check username
		var err error
		userInfo, err = l.Data.DBClient.User.Query().Where(user.Username(credential)).Only(ctx)
		if err != nil {
			return nil, err
		}
		if userInfo == nil {
			err = errors.New("login user not exist, find by username")
			return nil, err
		}
	}

	res = new(domain.LoginResp)
	res.Username = userInfo.Username
	res.UserID = userInfo.ID
	// get role info
	res.RoleID = userInfo.RoleID
	res.RoleName, res.RoleValue, err = l.getRoleInfo(ctx, userInfo.RoleID)

	return
}

func (l *Login) getRoleInfo(ctx context.Context, roleID uint64) (roleName, roleValue string, err error) {
	// get role info from cache
	v, exist := l.Data.Cache.Get("roleData" + strconv.Itoa(int(roleID)))
	if exist {
		roleName = v.(*ent.Role).Name
		roleValue = v.(*ent.Role).Value
		return
	}

	// get role info from database
	roleData, err := l.Data.DBClient.Role.Query().All(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			err = errors.Wrap(err, "fail to find any roles")
			return "", "", err
		}
		hlog.Error(err, "database error")
		return "", "", err
	}

	for _, v := range roleData {
		l.Data.Cache.Set("roleData"+strconv.Itoa(int(v.ID)), v, 24*time.Hour)
		if v.ID == roleID {
			roleName = v.Name
			roleValue = v.Value
		}
	}

	if roleName == "" || roleValue == "" {
		err = errors.New("fail to find any role info")
	}

	return roleName, roleValue, err
}
