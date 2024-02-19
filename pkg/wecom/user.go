/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

// Package wecom provides a client for accessing the WeCom API.(企业微信)

package wecom

import (
	"context"
	"formulago/configs"
	"github.com/chenghonour/wechat-sdk"
	"github.com/chenghonour/wechat-sdk/corp"
	"github.com/chenghonour/wechat-sdk/corp/addrbook"
	"github.com/cockroachdb/errors"
)

// GetUserIDByPhone get user id from wecom by phone
func GetUserIDByPhone(ctx context.Context, config configs.Config, phone string) (userID string, err error) {
	cp := gochat.NewCorp(config.Wecom.CorpID)
	// set debug mode (support custom log)
	//cp.SetClient(wx.WithDebug(), wx.WithLogger(wx.DefaultLogger()))
	// get token
	token, err := cp.AccessToken(ctx, config.Wecom.SecretID)
	if err != nil {
		err = errors.Wrap(err, "get wecom token failed")
		return "", err
	}
	// execute
	res := new(addrbook.ResultUserID)
	apply := addrbook.GetUserID(phone, res)
	if err := cp.Do(ctx, token.Token, apply); err != nil {
		err = errors.Wrap(err, "get wecom user id failed")
		return "", err
	}
	userID = res.UserID
	return
}

// GetUserByID get user info from wecom by userID
func GetUserByID(ctx context.Context, config configs.Config, userID string) (userInfo *addrbook.User, err error) {
	cp := gochat.NewCorp(config.Wecom.CorpID)
	// set debug mode (support custom log)
	//cp.SetClient(wx.WithDebug(), wx.WithLogger(wx.DefaultLogger()))
	// get token
	token, err := cp.AccessToken(ctx, config.Wecom.SecretID)
	if err != nil {
		err = errors.Wrap(err, "get wecom token failed")
		return nil, err
	}
	// execute
	userInfo = new(addrbook.User)
	apply := addrbook.GetUser(userID, userInfo)
	if err := cp.Do(ctx, token.Token, apply); err != nil {
		err = errors.Wrap(err, "get wecom user id failed")
		return nil, err
	}
	return
}

// GetOAuthUser get user info from wecom by auth code
func GetOAuthUser(ctx context.Context, config configs.Config, code string) (userInfo *corp.ResultOAuthUser, err error) {
	cp := gochat.NewCorp(config.Wecom.CorpID)
	// set debug mode (support custom log)
	//cp.SetClient(wx.WithDebug(), wx.WithLogger(wx.DefaultLogger()))
	// get token
	token, err := cp.AccessToken(ctx, config.Wecom.SecretID)
	if err != nil {
		err = errors.Wrap(err, "get wecom token failed")
		return nil, err
	}

	// execute
	userInfo = new(corp.ResultOAuthUser)
	apply := corp.GetOAuthUser(code, userInfo)
	if err := cp.Do(ctx, token.Token, apply); err != nil {
		err = errors.Wrap(err, "get wecom user id failed")
		return nil, err
	}
	return
}
