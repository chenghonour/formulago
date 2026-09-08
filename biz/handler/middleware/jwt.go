/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

// package middleware provides the middleware for the service http handler.

package middleware

import (
	"context"
	"errors"
	"formulago/biz/domain/admin"
	"formulago/configs"
	Data "formulago/data"
	"formulago/pkg/encrypt"
	"formulago/pkg/times"
	"net/http"
	"strconv"
	"sync"
	"time"

	logic "formulago/biz/logic/admin"

	"github.com/casbin/casbin/v3"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/golang-jwt/jwt/v4"
	hzjwt "github.com/hertz-contrib/jwt"
)

type jwtLogin struct {
	Username  string `form:"username,required" json:"username,required"`   //lint:ignore SA5008 ignoreCheck
	Password  string `form:"password,required" json:"password,required"`   //lint:ignore SA5008 ignoreCheck
	Captcha   string `form:"captcha,required" json:"captcha,required"`     //lint:ignore SA5008 ignoreCheck
	CaptchaID string `form:"captchaID,required" json:"captchaId,required"` //lint:ignore SA5008 ignoreCheck
}

// jwt identityKey
const (
	identityKey  = "jwt-id"
	tokenCtxKey  = "JWT_TOKEN" // key the jwt library stores the raw token under
	defaultRealm = "formulago"
	signingAlgo  = "HS256"
)

// auth errors that may reach the client without leaking internals; everything
// else is logged server-side and replaced with a generic message
var (
	errInvalidCaptcha = errors.New("invalid captcha")
	errAuthFailed     = errors.New("invalid username or password")
)

var (
	jwtInitOnce sync.Once
	jwtMW       *hzjwt.HertzJWTMiddleware
	jwtDB       *Data.Data
	jwtEnforcer *casbin.Enforcer
)

// GetJWTMiddleware returns the process-wide JWT middleware singleton.
// It is safe to call from every route registration: only the first call
// builds the middleware, later calls return the same instance.
func GetJWTMiddleware(c configs.Config, d *Data.Data, e *casbin.Enforcer) *hzjwt.HertzJWTMiddleware {
	jwtInitOnce.Do(func() {
		jwtDB = d
		jwtEnforcer = e
		var err error
		jwtMW, err = newJWT(c)
		if err != nil {
			hlog.Fatal(err, "JWT Init Error")
		}
	})
	return jwtMW
}

func newJWT(config configs.Config) (jwtMiddleware *hzjwt.HertzJWTMiddleware, err error) {
	jwtMiddleware, err = hzjwt.New(&hzjwt.HertzJWTMiddleware{
		Realm:            defaultRealm,
		Key:              []byte(config.Auth.AccessSecret),
		SigningAlgorithm: signingAlgo,
		Timeout:          time.Duration(config.Auth.AccessExpire) * time.Second,
		// session lifetime is enforced by the sys_tokens row (absolute
		// AccessExpire deadline, never extended by RefreshHandler), so the
		// library refresh window MaxRefresh is intentionally not used
		IdentityKey: identityKey,
		// header-only lookup: tokens in query/cookie/form would widen the
		// theft surface (logs, referrers, link prefetch)
		TokenLookup: "header:Authorization",
		// only "Bearer" is accepted as the Authorization scheme prefix
		TokenHeadName:   "Bearer",
		PayloadFunc:     payloadFunc,
		IdentityHandler: identityHandler,
		Authenticator: func(ctx context.Context, c *app.RequestContext) (any, error) {
			return authenticator(ctx, c, config)
		},
		LoginResponse: func(ctx context.Context, c *app.RequestContext, _ int, tokenStr string, expire time.Time) {
			loginResponse(ctx, c, tokenStr, expire)
		},
		Authorizator: func(data any, ctx context.Context, c *app.RequestContext) bool {
			return authorizator(data, ctx, c)
		},
		HTTPStatusMessageFunc: func(e error, ctx context.Context, c *app.RequestContext) string {
			return statusMessage(e, ctx)
		},
		Unauthorized: func(ctx context.Context, c *app.RequestContext, code int, message string) {
			c.JSON(code, map[string]any{
				"code":    code,
				"message": message,
			})
		},
	})
	return
}

// payloadFunc puts roleID, userID into the token claims. No sensitive data.
func payloadFunc(data any) hzjwt.MapClaims {
	if v, ok := data.(map[string]any); ok {
		return hzjwt.MapClaims{
			identityKey: v,
		}
	}
	return hzjwt.MapClaims{}
}

// identityHandler exposes roleID/userID from the verified claims on ctx.
func identityHandler(ctx context.Context, c *app.RequestContext) any {
	claims := hzjwt.ExtractClaims(ctx, c)
	payloadMap, ok := claims[identityKey].(map[string]any)
	if !ok {
		hlog.CtxErrorf(ctx, "identityHandler: unexpected claims payload %v", claims[identityKey])
		return nil
	}
	c.Set("roleID", payloadMap["roleID"])
	c.Set("userID", payloadMap["userID"])
	return payloadMap
}

func authenticator(ctx context.Context, c *app.RequestContext, config configs.Config) (any, error) {
	oauthLogin := ctx.Value("OAuthKey") == config.Auth.OAuthKey
	res := new(admin.LoginResp)
	if !oauthLogin {
		// normal jwtLogin
		var loginVal jwtLogin
		if err := c.BindAndValidate(&loginVal); err != nil {
			return "", err
		}
		// verify captcha while IsProd is true
		if config.IsProd {
			if !logic.CaptchaStore.Verify(loginVal.CaptchaID, loginVal.Captcha, true) {
				return nil, errInvalidCaptcha
			}
		}
		// Login
		var err error
		res, err = logic.NewLogin(jwtDB).Login(ctx, loginVal.Username, loginVal.Password)
		if err != nil {
			// full detail to the log only; the client gets a generic message
			// so DB errors never leak and accounts cannot be enumerated
			hlog.CtxErrorf(ctx, "jwt login failed: %v", err)
			return nil, errAuthFailed
		}
	} else {
		// oauth2.0 jwtLogin
		provider, ok := c.Get("provider")
		if !ok {
			return nil, errAuthFailed
		}
		providerStr, ok := provider.(string)
		if !ok {
			return nil, errAuthFailed
		}
		credential, ok := c.Get("credential")
		if !ok {
			return nil, errAuthFailed
		}
		credentialStr, ok := credential.(string)
		if !ok {
			return nil, errAuthFailed
		}
		var err error
		res, err = logic.NewLogin(jwtDB).LoginByOAuth(ctx, providerStr, credentialStr)
		if err != nil {
			hlog.CtxErrorf(ctx, "jwt oauth login failed: %v", err)
			return nil, errAuthFailed
		}
	}

	// the payload; roleID/userID are strings because JSON claims decode
	// numeric values as float64, strings keep the type assertion stable
	payloadMap := make(map[string]any)
	payloadMap["roleID"] = strconv.FormatUint(res.RoleID, 10)
	payloadMap["userID"] = strconv.FormatUint(res.UserID, 10)
	return payloadMap, nil
}

// loginResponse persists the session server-side and then answers with the
// same JSON shape as the library default LoginResponse.
func loginResponse(ctx context.Context, c *app.RequestContext, tokenStr string, expire time.Time) {
	// the session row must be written before the token is handed out, because
	// ActiveSession (Authorizator) only accepts the digest stored here
	uid, ok := userIDOfToken(ctx, tokenStr)
	if !ok {
		hlog.CtxErrorf(ctx, "loginResponse: cannot resolve userID from issued token")
		unauthorized(ctx, c, http.StatusInternalServerError, "internal error")
		return
	}
	err := logic.NewToken(jwtDB).Create(ctx, &admin.TokenInfo{
		UserID:    uid,
		Token:     tokenStr,
		ExpiredAt: expire.Format(times.TimeFormat),
	})
	if err != nil {
		hlog.CtxErrorf(ctx, "loginResponse: store token failed: %v", err)
		unauthorized(ctx, c, http.StatusInternalServerError, "internal error")
		return
	}
	c.JSON(http.StatusOK, map[string]any{
		"code":   http.StatusOK,
		"token":  tokenStr,
		"expire": expire.Format(time.RFC3339),
	})
}

// authorizator runs after signature verification for every protected route:
//  1. the presented token must match the server-side session digest,
//  2. the user must exist and be enabled,
//  3. the permission subject is the user's CURRENT role from the DB, so role
//     changes take effect immediately instead of at the next login.
func authorizator(data any, ctx context.Context, c *app.RequestContext) bool {
	payloadMap, ok := data.(map[string]any)
	if !ok {
		hlog.CtxErrorf(ctx, "authorizator: unexpected payload type %T", data)
		return false
	}
	userIDStr, ok := payloadMap["userID"].(string)
	if !ok || userIDStr == "" {
		hlog.CtxErrorf(ctx, "authorizator: missing userID in claims")
		return false
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		hlog.CtxErrorf(ctx, "authorizator: invalid userID %q", userIDStr)
		return false
	}

	// 1. exact session credential check against sys_tokens
	tokenVal, exists := c.Get(tokenCtxKey)
	if !exists {
		hlog.CtxErrorf(ctx, "authorizator: missing raw token")
		return false
	}
	tokenRaw, ok := tokenVal.(string)
	if !ok || tokenRaw == "" {
		hlog.CtxErrorf(ctx, "authorizator: unexpected raw token type %T", tokenVal)
		return false
	}
	if _, err := logic.NewToken(jwtDB).ActiveSession(ctx, userID, encrypt.Sha256Hex(tokenRaw)); err != nil {
		if !errors.Is(err, admin.ErrSessionInvalid) {
			hlog.CtxErrorf(ctx, "authorizator: session check failed: %v", err)
		} else {
			hlog.CtxDebugf(ctx, "authorizator: no active session for user %d", userID)
		}
		return false
	}

	// 2. user must exist and be enabled (disable takes effect immediately)
	userInfo, err := logic.NewUser(jwtDB).UserInfo(ctx, userID)
	if err != nil {
		hlog.CtxErrorf(ctx, "authorizator: get user %d failed: %v", userID, err)
		return false
	}
	if userInfo.Status != 1 {
		hlog.CtxDebugf(ctx, "authorizator: user %d is disabled", userID)
		return false
	}

	// 3. the role must exist and be active
	roleID := userInfo.RoleID
	roleInfo, err := logic.NewRole(jwtDB).RoleInfoByID(ctx, roleID)
	if err != nil {
		hlog.CtxErrorf(ctx, "authorizator: get role %d failed: %v", roleID, err)
		return false
	}
	if roleInfo.Status != 1 {
		hlog.CtxDebugf(ctx, "authorizator: role %d is not active", roleID)
		return false
	}

	// 4. authorize the path with the CURRENT role, then refresh the claims
	//    held on the request context so downstream handlers see the same role
	sub := strconv.FormatUint(roleID, 10)
	obj := string(c.URI().Path())
	act := string(c.Method())
	pass, err := jwtEnforcer.Enforce(sub, obj, act)
	if err != nil {
		hlog.CtxErrorf(ctx, "authorizator: casbin enforce err, role: %s path: %s method: %s: %v", sub, obj, act, err)
		return false
	}
	if !pass {
		hlog.CtxDebugf(ctx, "authorizator: casbin forbid role: %s path: %s method: %s", sub, obj, act)
		return false
	}
	c.Set("roleID", sub)
	c.Set("userID", userIDStr)
	return true
}

// LogoutHandler revokes the server-side session of the presented token.
// Idempotent: answering 200 is fine even when the token is missing/expired,
// because an expired session can no longer pass ActiveSession anyway.
func LogoutHandler(ctx context.Context, c *app.RequestContext) {
	GetJWTMiddleware(configs.Data(), Data.Default(), Data.CasbinEnforcer())
	if uid, ok := userIDOfRequest(ctx, c); ok {
		if err := logic.NewToken(jwtDB).Delete(ctx, uid); err != nil {
			hlog.CtxErrorf(ctx, "logout: delete session failed: %v", err)
		}
	}
	c.JSON(http.StatusOK, map[string]any{"code": http.StatusOK})
}

// RefreshHandler re-issues the JWT only when the presented token still maps to
// a live server-side session, then rotates the stored digest (compare-and-swap),
// so revoked or stolen sessions cannot be kept alive by refreshing. The new
// token expires at the session's absolute deadline, never later.
func RefreshHandler(ctx context.Context, c *app.RequestContext) {
	GetJWTMiddleware(configs.Data(), Data.Default(), Data.CasbinEnforcer())
	token, err := jwtMW.ParseToken(ctx, c) // verifies signature and exp
	if err != nil {
		unauthorized(ctx, c, http.StatusUnauthorized, statusMessage(err, ctx))
		return
	}
	tokenVal, exists := c.Get(tokenCtxKey)
	if !exists {
		unauthorized(ctx, c, http.StatusUnauthorized, "missing token")
		return
	}
	tokenRaw, ok := tokenVal.(string)
	if !ok || tokenRaw == "" {
		unauthorized(ctx, c, http.StatusUnauthorized, "missing token")
		return
	}
	claims := hzjwt.ExtractClaimsFromToken(token)
	payloadMap, ok := claims[identityKey].(map[string]any)
	if !ok {
		unauthorized(ctx, c, http.StatusUnauthorized, "invalid token claims")
		return
	}
	userIDStr, ok := payloadMap["userID"].(string)
	if !ok {
		unauthorized(ctx, c, http.StatusUnauthorized, "invalid token claims")
		return
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		unauthorized(ctx, c, http.StatusUnauthorized, "invalid token claims")
		return
	}
	oldHash := encrypt.Sha256Hex(tokenRaw)
	session, err := logic.NewToken(jwtDB).ActiveSession(ctx, userID, oldHash)
	if err != nil {
		if errors.Is(err, admin.ErrSessionInvalid) {
			unauthorized(ctx, c, http.StatusUnauthorized, "session is invalid, please login again")
		} else {
			hlog.CtxErrorf(ctx, "refresh: session check failed: %v", err)
			unauthorized(ctx, c, http.StatusInternalServerError, "internal error")
		}
		return
	}
	// absolute session deadline: a refresh never extends the session
	newToken := jwt.NewWithClaims(jwt.GetSigningMethod(jwtMW.SigningAlgorithm), jwt.MapClaims{
		identityKey: payloadMap,
		"exp":       session.ExpiredAt.Unix(),
		"orig_iat":  time.Now().Unix(),
	})
	signed, err := newToken.SignedString(jwtMW.Key)
	if err != nil {
		hlog.CtxErrorf(ctx, "refresh: sign token failed: %v", err)
		unauthorized(ctx, c, http.StatusInternalServerError, "internal error")
		return
	}
	if err := logic.NewToken(jwtDB).Rotate(ctx, userID, oldHash, encrypt.Sha256Hex(signed)); err != nil {
		// session was revoked or rotated concurrently: do not hand out a token
		hlog.CtxDebugf(ctx, "refresh: session rotated or revoked concurrently")
		unauthorized(ctx, c, http.StatusUnauthorized, "session is invalid, please login again")
		return
	}
	c.JSON(http.StatusOK, map[string]any{
		"code":   http.StatusOK,
		"token":  signed,
		"expire": session.ExpiredAt.Format(time.RFC3339),
	})
}

// userIDOfRequest resolves the authenticated userID from the request token.
func userIDOfRequest(ctx context.Context, c *app.RequestContext) (uint64, bool) {
	token, err := jwtMW.ParseToken(ctx, c)
	if err != nil {
		hlog.CtxDebugf(ctx, "logout: parse token failed: %v", err)
		return 0, false
	}
	return userIDOfToken(ctx, token.Raw)
}

// userIDOfToken extracts the userID stored in our identityKey payload.
func userIDOfToken(ctx context.Context, tokenStr string) (uint64, bool) {
	token, err := jwtMW.ParseTokenString(tokenStr)
	if err != nil {
		hlog.CtxErrorf(ctx, "resolve userID from token failed: %v", err)
		return 0, false
	}
	claims := hzjwt.ExtractClaimsFromToken(token)
	payloadMap, ok := claims[identityKey].(map[string]any)
	if !ok {
		hlog.CtxErrorf(ctx, "unexpected token claims %v", claims[identityKey])
		return 0, false
	}
	userIDStr, ok := payloadMap["userID"].(string)
	if !ok {
		hlog.CtxErrorf(ctx, "missing userID in token claims")
		return 0, false
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		hlog.CtxErrorf(ctx, "invalid userID in token claims: %v", err)
		return 0, false
	}
	return userID, true
}

// statusMessage converts an error to a client-safe message. Everything is
// logged; only known sentinels and generic library errors are sent verbatim.
func statusMessage(e error, ctx context.Context) string {
	switch {
	case errors.Is(e, admin.ErrSessionInvalid):
		return "session is invalid, please login again"
	case errors.Is(e, errInvalidCaptcha), errors.Is(e, errAuthFailed):
		return e.Error()
	}
	hlog.CtxDebugf(ctx, "jwt middleware error: %v", e)
	return e.Error()
}

func unauthorized(ctx context.Context, c *app.RequestContext, code int, message string) {
	if jwtMW != nil {
		jwtMW.Unauthorized(ctx, c, code, message)
		return
	}
	c.JSON(code, map[string]any{"code": code, "message": message})
}
