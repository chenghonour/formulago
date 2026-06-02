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
	"formulago/data"
	"formulago/data/ent"
	"formulago/data/ent/predicate"
	"formulago/data/ent/user"
	"formulago/pkg/encrypt"
	"regexp"
	"strconv"
	"time"

	"github.com/jinzhu/copier"
)

var usernameRegex = regexp.MustCompile(`^[a-z]{5,}$`)

func validateUserInput(username, password string) error {
	if !usernameRegex.MatchString(username) {
		return errors.New("username must be at least 5 lowercase letters, no spaces or special characters")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

type User struct {
	Data *data.Data
}

func NewUser(data *data.Data) admin.User {
	return &User{
		Data: data,
	}
}

func (u *User) Create(ctx context.Context, req admin.CreateOrUpdateUserReq) error {
	if err := validateUserInput(req.Username, req.Password); err != nil {
		return err
	}
	password, _ := encrypt.BcryptEncrypt(req.Password)
	_, err := u.Data.DBClient.User.Create().
		SetAvatar(req.Avatar).
		SetRoleID(req.RoleID).
		SetMobile(req.Mobile).
		SetEmail(req.Email).
		SetStatus(uint8(req.Status)).
		SetUsername(req.Username).
		SetPassword(password).
		SetNickname(req.Nickname).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}

	return nil
}

func (u *User) Update(ctx context.Context, req admin.CreateOrUpdateUserReq) error {
	if err := validateUserInput(req.Username, req.Password); err != nil {
		return err
	}
	password, _ := encrypt.BcryptEncrypt(req.Password)
	_, err := u.Data.DBClient.User.Update().
		Where(user.IDEQ(req.ID)).
		SetAvatar(req.Avatar).
		SetRoleID(req.RoleID).
		SetMobile(req.Mobile).
		SetEmail(req.Email).
		SetStatus(uint8(req.Status)).
		SetUsername(req.Username).
		SetPassword(password).
		SetNickname(req.Nickname).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("update user failed: %w", err)
	}

	return nil
}

func (u *User) ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	// get user info
	targetUser, err := u.Data.DBClient.User.Query().Where(user.IDEQ(userID)).First(ctx)
	if err != nil {
		return fmt.Errorf("targetUser not found: %w", err)
	}
	// check old password
	if ok := encrypt.BcryptCheck(oldPassword, targetUser.Password); !ok {
		return errors.New("wrong old password")
	}
	// update password
	password, _ := encrypt.BcryptEncrypt(newPassword)
	_, err = u.Data.DBClient.User.Update().Where(user.IDEQ(userID)).SetPassword(password).Save(ctx)

	return err
}

func (u *User) UserInfo(ctx context.Context, id uint64) (userInfo *admin.UserInfo, err error) {
	userInfo = new(admin.UserInfo)
	// get user info from cache
	userInterface, exist := u.Data.Cache.Get("userInfo" + strconv.Itoa(int(id)))
	if exist {
		if u, ok := userInterface.(*admin.UserInfo); ok {
			return u, nil
		}
	}
	// get user info from db
	userEnt, err := u.Data.DBClient.User.Query().Where(user.IDEQ(id)).First(ctx)
	if err != nil {
		err = fmt.Errorf("get user failed: %w", err)
		return userInfo, err
	}
	// copy to UserInfo struct
	err = copier.Copy(&userInfo, &userEnt)
	if err != nil {
		err = fmt.Errorf("copy user info failed: %w", err)
		return userInfo, err
	}
	// role info from cache
	roleInterface, exist := u.Data.Cache.Get("roleData" + strconv.Itoa(int(userInfo.RoleID)))
	if exist {
		if role, ok := roleInterface.(*ent.Role); ok {
			userInfo.RoleName = role.Name
			userInfo.RoleValue = role.Value
			userInfo.DefaultRouter = role.DefaultRouter
		}
	}
	// set user to cache
	u.Data.Cache.Set("userInfo"+strconv.Itoa(int(userInfo.ID)), &userInfo, 72*time.Hour)
	return
}

func (u *User) List(ctx context.Context, req admin.UserListReq) (userList []*admin.UserInfo, total int, err error) {
	var predicates []predicate.User
	if req.Mobile != "" {
		predicates = append(predicates, user.MobileEQ(req.Mobile))
	}

	if req.Username != "" {
		predicates = append(predicates, user.UsernameContains(req.Username))
	}

	if req.Email != "" {
		predicates = append(predicates, user.EmailEQ(req.Email))
	}

	if req.Nickname != "" {
		predicates = append(predicates, user.NicknameContains(req.Nickname))
	}

	if req.RoleID != 0 {
		predicates = append(predicates, user.RoleIDEQ(req.RoleID))
	}

	users, err := u.Data.DBClient.User.Query().Where(predicates...).
		Offset(int(req.Page-1) * int(req.PageSize)).
		Limit(int(req.PageSize)).All(ctx)
	if err != nil {
		err = fmt.Errorf("get user list failed: %w", err)
		return userList, total, err
	}
	// copy to UserInfo struct
	err = copier.Copy(&userList, &users)
	if err != nil {
		err = fmt.Errorf("copy user info failed: %w", err)
		return userList, 0, err
	}
	total, err = u.Data.DBClient.User.Query().Where(predicates...).Count(ctx)
	if err != nil {
		err = fmt.Errorf("count user failed: %w", err)
		return
	}
	return
}

func (u *User) UpdateUserStatus(ctx context.Context, id uint64, status uint64) error {
	_, err := u.Data.DBClient.User.Update().Where(user.IDEQ(id)).SetStatus(uint8(status)).Save(ctx)
	return err
}

func (u *User) DeleteUser(ctx context.Context, id uint64) error {
	_, err := u.Data.DBClient.User.Delete().Where(user.IDEQ(id)).Exec(ctx)
	return err
}

func (u *User) UpdateProfile(ctx context.Context, req admin.UpdateUserProfileReq) error {
	_, err := u.Data.DBClient.User.Update().
		Where(user.IDEQ(req.ID)).
		SetNickname(req.Nickname).
		SetEmail(req.Email).
		SetMobile(req.Mobile).
		SetAvatar(req.Avatar).
		Save(ctx)
	return err
}
