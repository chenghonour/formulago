/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

// Package wecom provides a client for accessing the WeCom API.(企业微信)

package wecom

import (
	"context"

	wechatSDK "github.com/chenghonour/wechat-sdk"
	"github.com/chenghonour/wechat-sdk/corp"
	"github.com/chenghonour/wechat-sdk/corp/addrbook"
	"github.com/pkg/errors"
)

// GetUserIDByPhone get user id from wecom by phone
func (w *Wecom) GetUserIDByPhone(ctx context.Context, phone string) (userID string, err error) {
	cp := wechatSDK.NewCorp(w.Config.Wecom.CorpID)
	// set debug mode (support custom log)
	// cp.SetClient(wx.WithDebug(), wx.WithLogger(wx.DefaultLogger()))
	// get token
	token, err := cp.AccessToken(ctx, w.Config.Wecom.SecretID)
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
func (w *Wecom) GetUserByID(ctx context.Context, userID string) (userInfo *addrbook.User, err error) {
	cp := wechatSDK.NewCorp(w.Config.Wecom.CorpID)
	// set debug mode (support custom log)
	// cp.SetClient(wx.WithDebug(), wx.WithLogger(wx.DefaultLogger()))
	// get token
	token, err := cp.AccessToken(ctx, w.Config.Wecom.SecretID)
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
func (w *Wecom) GetOAuthUser(ctx context.Context, code string) (userInfo *corp.ResultOAuthUser, err error) {
	cp := wechatSDK.NewCorp(w.Config.Wecom.CorpID)
	// set debug mode (support custom log)
	// cp.SetClient(wx.WithDebug(), wx.WithLogger(wx.DefaultLogger()))
	// get token
	token, err := cp.AccessToken(ctx, w.Config.Wecom.SecretID)
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
