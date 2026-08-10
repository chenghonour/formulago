/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

// Package wecom provides a client for accessing the WeCom API.(企业微信)

package wecom

import (
	"context"
	"fmt"
)

// OAuthUser 企业微信网页授权登录解析出的用户身份
type OAuthUser struct {
	UserID string
}

// User 企业微信成员信息
type User struct {
	UserID string
	Mobile string
	Email  string
}

// GetUserIDByPhone get user id from wecom by phone
func (w *Wecom) GetUserIDByPhone(ctx context.Context, phone string) (userID string, err error) {
	app, err := w.app()
	if err != nil {
		return "", fmt.Errorf("init wecom app failed: %w", err)
	}
	res, err := app.User.MobileToUserID(ctx, phone)
	if err != nil {
		return "", fmt.Errorf("get wecom user id failed: %w", err)
	}
	if res.IsError() {
		return "", fmt.Errorf("get wecom user id failed: %w", res)
	}
	return res.UserID, nil
}

// GetUserByID get user info from wecom by userID
func (w *Wecom) GetUserByID(ctx context.Context, userID string) (*User, error) {
	app, err := w.app()
	if err != nil {
		return nil, fmt.Errorf("init wecom app failed: %w", err)
	}
	res, err := app.User.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get wecom user failed: %w", err)
	}
	if res.IsError() {
		return nil, fmt.Errorf("get wecom user failed: %w", res)
	}
	return &User{
		UserID: res.UserID,
		Mobile: res.Mobile,
		Email:  res.Email,
	}, nil
}

// GetOAuthUser get user info from wecom by auth code
func (w *Wecom) GetOAuthUser(ctx context.Context, code string) (*OAuthUser, error) {
	app, err := w.app()
	if err != nil {
		return nil, fmt.Errorf("init wecom app failed: %w", err)
	}
	res, err := app.OAuth.Provider.GetUserInfo(code)
	if err != nil {
		return nil, fmt.Errorf("get wecom oauth user failed: %w", err)
	}
	if res.ErrCode != 0 {
		return nil, fmt.Errorf("get wecom oauth user failed: errcode=%d errmsg=%s", res.ErrCode, res.ErrMSG)
	}
	return &OAuthUser{UserID: res.UserID}, nil
}
