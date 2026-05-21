/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import (
	"context"
	"fmt"
"errors"
	"formulago/biz/domain/admin"
	"formulago/pkg/times"
	"formulago/configs"
	"formulago/data"
	"formulago/data/ent/oauthprovider"
	"formulago/data/ent/predicate"
	"formulago/pkg/wecom"
	"github.com/cloudwego/hertz/pkg/common/json"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Oauth struct {
	Data   *data.Data
	Config configs.Config
}

func NewOauth(data *data.Data, config configs.Config) admin.Oauth {
	return &Oauth{
		Data:   data,
		Config: config,
	}
}

func (o *Oauth) Create(ctx context.Context, providerReq *admin.ProviderInfo) error {
	_, err := o.Data.DBClient.OauthProvider.Create().
		SetName(providerReq.Name).
		SetClientID(providerReq.ClientID).
		SetClientSecret(providerReq.ClientSecret).
		SetRedirectURL(providerReq.RedirectUrl).
		SetScopes(providerReq.Scopes).
		SetAuthURL(providerReq.AuthUrl).
		SetTokenURL(providerReq.TokenUrl).
		SetAuthStyle(providerReq.AuthStyle).
		SetInfoURL(providerReq.InfoUrl).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create oauth failed: %w", err)
	}
	return nil
}

func (o *Oauth) Update(ctx context.Context, providerReq *admin.ProviderInfo) error {
	_, err := o.Data.DBClient.OauthProvider.UpdateOneID(providerReq.ID).
		SetName(providerReq.Name).
		SetClientID(providerReq.ClientID).
		SetClientSecret(providerReq.ClientSecret).
		SetRedirectURL(providerReq.RedirectUrl).
		SetScopes(providerReq.Scopes).
		SetAuthURL(providerReq.AuthUrl).
		SetTokenURL(providerReq.TokenUrl).
		SetAuthStyle(providerReq.AuthStyle).
		SetInfoURL(providerReq.InfoUrl).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("update oauth failed: %w", err)
	}
	return nil
}

func (o *Oauth) Delete(ctx context.Context, providerID uint64) error {
	err := o.Data.DBClient.OauthProvider.DeleteOneID(providerID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete oauth failed: %w", err)
	}
	return nil
}

func (o *Oauth) List(ctx context.Context, req *admin.OauthListReq) (list []*admin.ProviderInfo, total int, err error) {
	var predicates []predicate.OauthProvider
	if req.Name != "" {
		predicates = append(predicates, oauthprovider.NameContains(req.Name))
	}
	providers, err := o.Data.DBClient.OauthProvider.Query().
		Where(predicates...).
		Offset(int((req.Page - 1) * req.PageSize)).
		Limit(int(req.PageSize)).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list oauth failed: %w", err)
	}

	for _, provider := range providers {
		list = append(list, &admin.ProviderInfo{
			ID:           provider.ID,
			Name:         provider.Name,
			ClientID:     provider.ClientID,
			ClientSecret: provider.ClientSecret,
			RedirectUrl:  provider.RedirectURL,
			Scopes:       provider.Scopes,
			AuthUrl:      provider.AuthURL,
			TokenUrl:     provider.TokenURL,
			AuthStyle:    provider.AuthStyle,
			InfoUrl:      provider.InfoURL,
			CreatedAt:    provider.CreatedAt.Format(times.TimeFormat),
			UpdatedAt:    provider.UpdatedAt.Format(times.TimeFormat),
		})
	}
	total, err = o.Data.DBClient.OauthProvider.Query().Where(predicates...).Count(ctx)
	if err != nil {
		err = fmt.Errorf("count oauth provider failed: %w", err)
		return list, total, err
	}
	return list, total, nil
}

func (o *Oauth) Login(ctx context.Context, req *admin.OauthLoginReq) (string, error) {
	provider, err := o.Data.DBClient.OauthProvider.Query().Where(oauthprovider.Name(req.Provider)).First(ctx)
	if err != nil {
		return "", fmt.Errorf("get oauth provider failed: %w", err)
	}

	var config oauth2.Config
	if v, found := o.Data.Cache.Get("oauthProviderConfig" + provider.Name); found {
		var ok bool
		config, ok = v.(oauth2.Config)
		if !ok {
			return "", errors.New("get cache provider config failed")
		}
	} else {
		config = oauth2.Config{
			ClientID:     provider.ClientID,
			ClientSecret: provider.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:   provider.AuthURL,
				TokenURL:  provider.TokenURL,
				AuthStyle: oauth2.AuthStyle(provider.AuthStyle),
			},
			RedirectURL: provider.RedirectURL,
			Scopes:      strings.Split(provider.Scopes, " "),
		}
		o.Data.Cache.Set("oauthProviderConfig"+provider.Name, config, 24*time.Hour)
	}

	if _, ok := o.Data.Cache.Get("oauthProviderUserInfoURL" + provider.Name); !ok {
		o.Data.Cache.Set("oauthProviderUserInfoURL"+provider.Name, provider.InfoURL, 24*time.Hour)
	}

	var oauthURL string
	switch provider.Name {
	case "wecom":
		if req.LoginType == "Inside" {
			oauthURL = fmt.Sprintf("https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_base&state=%s&agentid=%s#wechat_redirect",
				provider.AppID, url.QueryEscape(provider.RedirectURL), req.State, provider.ClientID)
		} else {
			oauthURL = config.AuthCodeURL(req.State, oauth2.SetAuthURLParam("login_type", "CorpApp"),
				oauth2.SetAuthURLParam("appid", provider.AppID),
				oauth2.SetAuthURLParam("agentid", provider.ClientID))
		}
	default:
		oauthURL = config.AuthCodeURL(req.State)
	}

	return oauthURL, nil
}

func (o *Oauth) Callback(ctx context.Context, req *admin.OauthCallbackReq) (*admin.OauthUserInfo, error) {
	if _, found := o.Data.Cache.Get("oauthProviderConfig" + req.ProviderName); !found {
		provider, err := o.Data.DBClient.OauthProvider.Query().Where(oauthprovider.Name(req.ProviderName)).First(ctx)
		if err != nil {
			return nil, fmt.Errorf("get oauth provider failed: %w", err)
		}
		config := oauth2.Config{
			ClientID:     provider.ClientID,
			ClientSecret: provider.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:   provider.AuthURL,
				TokenURL:  provider.TokenURL,
				AuthStyle: oauth2.AuthStyle(provider.AuthStyle),
			},
			RedirectURL: provider.RedirectURL,
			Scopes:      strings.Split(provider.Scopes, " "),
		}
		o.Data.Cache.Set("oauthProviderConfig"+provider.Name, config, 24*time.Hour)

		if _, ok := o.Data.Cache.Get("oauthProviderUserInfoURL" + provider.Name); !ok {
			o.Data.Cache.Set("oauthProviderUserInfoURL"+provider.Name, provider.InfoURL, 24*time.Hour)
		}
	}

	// get user information
	userInfo := new(admin.OauthUserInfo)
	switch req.ProviderName {
	case "wecom":
		wecom := wecom.New(o.Config, o.Data)
		u, err := wecom.GetOAuthUser(ctx, req.Code)
		if err != nil {
			return nil, fmt.Errorf("get wecom user info failed: %w", err)
		}
		wecomUser, err := wecom.GetUserByID(ctx, u.UserID)
		if err != nil {
			return nil, fmt.Errorf("get wecom user info failed: %w", err)
		}
		userInfo.Credential = wecomUser.UserID
		userInfo.Mobile = wecomUser.Mobile
		userInfo.Email = wecomUser.Email
	default:
		c, ok := o.Data.Cache.Get("oauthProviderConfig" + req.ProviderName)
		if !ok {
			return nil, errors.New("get cache provider config failed")
		}
		userInfoURL, ok := o.Data.Cache.Get("oauthProviderUserInfoURL" + req.ProviderName)
		if !ok {
			return nil, errors.New("get cache provider user info url failed")
		}
		var err error
		userInfo, err = getUserInfo(c.(oauth2.Config), userInfoURL.(string), req.Code)
		if err != nil {
			return nil, fmt.Errorf("get user info failed: %w", err)
		}
		if userInfo == nil {
			userInfo.Credential = userInfo.Username
		}
	}

	return userInfo, nil
}

func getUserInfo(c oauth2.Config, infoURL string, code string) (*admin.OauthUserInfo, error) {
	token, err := c.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %w", err)
	}

	var response *http.Response
	if c.Endpoint.AuthStyle == 1 {
		response, err = http.Get(infoURL + token.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("failed getting user info: %s", err.Error())
		}
	} else if c.Endpoint.AuthStyle == 2 {
		client := &http.Client{}
		request, err := http.NewRequest("GET", infoURL, nil)
		if err != nil {
			return nil, fmt.Errorf("Endpoint Request failed: %w", err)
		}

		request.Header.Set("Accept", "application/json")
		request.Header.Set("Authorization", "Bearer "+token.AccessToken)

		response, err = client.Do(request)
		if err != nil {
			return nil, fmt.Errorf("failed getting user info: %w", err)
		}
	}

	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
	}

	var u *admin.OauthUserInfo
	err = json.Unmarshal(contents, &u)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshaling response body: %w", err)
	}

	return u, nil
}
