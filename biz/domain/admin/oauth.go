/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import "context"

type Oauth interface {
	Create(ctx context.Context, providerReq *ProviderInfo) error
	Update(ctx context.Context, providerReq *ProviderInfo) error
	Delete(ctx context.Context, providerID uint64) error
	List(ctx context.Context, req *OauthListReq) (list []*ProviderInfo, total int, err error)
	Login(ctx context.Context, req *OauthLoginReq) (string, error)
	Callback(ctx context.Context, req *OauthCallbackReq) (*OauthUserInfo, error)
}

type ProviderInfo struct {
	ID           uint64
	Name         string
	ClientID     string
	ClientSecret string
	RedirectUrl  string
	Scopes       string
	AuthUrl      string
	TokenUrl     string
	AuthStyle    uint64
	InfoUrl      string
	CreatedAt    string
	UpdatedAt    string
}

type OauthListReq struct {
	Page     uint64
	PageSize uint64
	Name     string
}

type OauthLoginReq struct {
	State    string
	Provider string
	// Login type for wecom: QRCode, Inside, Quick | 企业微信登录类型：二维码，应用内部，快速 （其他应用的OAuth不需要传）
	LoginType string
}

type OauthCallbackReq struct {
	ProviderName string
	State        string
	Code         string
}

type OauthUserInfo struct {
	Credential string `json:"credential"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Mobile     string `json:"mobile"`
	NickName   string `json:"nickName"`
	Picture    string `json:"picture"`
}
