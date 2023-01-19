/*
 * Copyright 2022 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

// package middleware provides the middleware for the service http handler.

package middleware

import (
	"context"
	"formulago/biz/domain"
	"github.com/cockroachdb/errors"
	"strconv"
	"time"

	"formulago/data/ent"

	logic "formulago/biz/logic/admin"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/casbin/casbin/v2"

	"formulago/configs"
	Data "formulago/data"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/jwt"
)

type jwtLogin struct {
	Username  string `form:"username,required" json:"username,required"`   //lint:ignore SA5008 ignoreCheck
	Password  string `form:"password,required" json:"password,required"`   //lint:ignore SA5008 ignoreCheck
	Captcha   string `form:"captcha,required" json:"captcha,required"`     //lint:ignore SA5008 ignoreCheck
	CaptchaID string `form:"captchaID,required" json:"captchaId,required"` //lint:ignore SA5008 ignoreCheck
}

// jwt identityKey
var (
	identityKey   = "jwt-id"
	jwtMiddleware = new(jwt.HertzJWTMiddleware)
)

// init the jwt middleware
func init() {
	var err error
	jwtMiddleware, err = newJWT(configs.Data(), Data.Default(), Data.CasbinEnforcer())
	if err != nil {
		hlog.Fatal(err, "JWT Init Error")
	}
}

func GetJWTMiddleware() *jwt.HertzJWTMiddleware {
	return jwtMiddleware
}

func newJWT(config configs.Config, db *Data.Data, enforcer *casbin.Enforcer) (jwtMiddleware *jwt.HertzJWTMiddleware, err error) {
	// the jwt middleware
	jwtMiddleware, err = jwt.New(&jwt.HertzJWTMiddleware{
		Realm:       "formulago",
		Key:         []byte(config.Auth.AccessSecret),
		Timeout:     time.Duration(config.Auth.AccessExpire) * time.Second,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			// take map which have roleID, userID as Payload
			if v, ok := data.(map[string]interface{}); ok {
				return jwt.MapClaims{
					identityKey: v,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(ctx context.Context, c *app.RequestContext) interface{} {
			claims := jwt.ExtractClaims(ctx, c)
			payloadMap, ok := claims[identityKey].(map[string]interface{})
			if !ok {
				hlog.Error("get payloadMap error", "claims data:", claims[identityKey])
				return nil
			}
			// take roleID, userID from PayloadMap
			c.Set("roleID", payloadMap["roleID"])
			c.Set("userID", payloadMap["userID"])
			return payloadMap
		},
		Authenticator: func(ctx context.Context, c *app.RequestContext) (interface{}, error) {
			var oauthLogin = ctx.Value("OAuthKey") == config.Auth.OAuthKey
			var res = new(domain.LoginResp)
			if !oauthLogin {
				// normal jwtLogin
				var loginVal jwtLogin
				if err := c.BindAndValidate(&loginVal); err != nil {
					return "", err
				}
				// verify captcha
				valid := logic.CaptchaStore.Verify(loginVal.CaptchaID, loginVal.Captcha, true)
				if !valid {
					return nil, errors.New("invalid captcha")
				}

				// Login
				username := loginVal.Username
				password := loginVal.Password
				res, err = logic.NewLogin(db).Login(ctx, username, password)
				if err != nil {
					hlog.Error(err, "jwtLogin error")
					return nil, err
				}
			} else {
				// oauth2.0 jwtLogin
				providerAny, ok := c.Get("provider")
				if !ok {
					return nil, errors.New("invalid provider")
				}
				provider, ok := providerAny.(string)
				if !ok {
					return nil, errors.New("invalid provider")
				}

				credentialAny, ok := c.Get("credential")
				if !ok {
					return nil, errors.New("invalid credential")
				}
				credential, ok := credentialAny.(string)
				if !ok {
					return nil, errors.New("invalid credential")
				}
				res, err = logic.NewLogin(db).LoginByOAuth(ctx, provider, credential)
				if err != nil {
					hlog.Error(err, "oauth jwtLogin error")
					return nil, err
				}
			}

			// jwtLogin success
			// store token
			var tokenInfo domain.TokenInfo
			tokenInfo.UserID = res.UserID
			tokenInfo.UserName = res.Username
			tokenInfo.ExpiredAt = time.Now().Add(time.Duration(config.Auth.AccessExpire) * time.Second).Format("2006-01-02 15:04:05")
			err = logic.NewToken(db).Create(ctx, &tokenInfo)
			if err != nil {
				hlog.Error(err, "jwtLogin error, store token error")
				return nil, err
			}
			// store UserID-Username in cache
			Data.Default().Cache.Set("UserID2Username"+strconv.Itoa(int(res.UserID)), res.Username, time.Duration(config.Auth.AccessExpire)*time.Second)

			// return the payload
			// take str roleID, userID into PayloadMap
			payloadMap := make(map[string]interface{})
			payloadMap["roleID"] = strconv.Itoa(int(res.RoleID))
			payloadMap["userID"] = strconv.Itoa(int(res.UserID))
			return payloadMap, nil
		},
		Authorizator: func(data interface{}, ctx context.Context, c *app.RequestContext) bool {
			// get the path
			obj := string(c.URI().Path())
			// get the method
			act := string(c.Method())
			// get the roleID
			payloadMap, ok := data.(map[string]interface{})
			if !ok {
				hlog.Error("get payloadMap error", "claims data:", data)
				return false
			}
			roleID := payloadMap["roleID"].(string)
			userID := payloadMap["userID"].(string)

			// check token is valid
			userIDInt, err := strconv.Atoi(userID)
			if err != nil {
				hlog.Error("get payloadMap error", err)
				return false
			}
			existToken := logic.NewToken(db).IsExistByUserID(ctx, uint64(userIDInt))
			if !existToken {
				return false
			}

			// check the role status
			roleInterface, exist := db.Cache.Get("roleData" + roleID)
			// if the role cache is not exist or the role is not active, return false
			if !exist {
				hlog.Error("role cache is not exist")
				return false
			}
			role, ok := roleInterface.(*ent.Role)
			if !ok || role.Status != 1 {
				hlog.Error("role cache is not a valid *ent.Role or the role is not active")
				return false
			}

			sub := roleID
			// check the permission
			pass, err := enforcer.Enforce(sub, obj, act)
			if err != nil {
				hlog.Error("casbin err,  role id: ", roleID, " path: ", obj, " method: ", act, " pass: ", pass, " err: ", err.Error())
				return false
			}
			if !pass {
				hlog.Info("casbin forbid role id: ", roleID, " path: ", obj, " method: ", act, " pass: ", pass)
			}
			hlog.Info("casbin allow role id: ", roleID, " path: ", obj, " method: ", act, " pass: ", pass)
			return pass
		},
		Unauthorized: func(ctx context.Context, c *app.RequestContext, code int, message string) {
			c.JSON(code, map[string]interface{}{
				"code":    code,
				"message": message,
			})
		},
	})

	return
}
