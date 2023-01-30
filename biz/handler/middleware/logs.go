/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package middleware

import (
	"context"
	"formulago/biz/domain"
	"formulago/biz/logic/admin"
	"formulago/data"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"time"
)

func LogsMiddleware(d *data.Data) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// pre-handle
		start := time.Now()
		// ...
		c.Next(ctx)
		// post-handle
		var logs domain.LogsInfo
		logs.Type = "Interface"
		logs.Method = string(c.Request.Method())
		logs.Api = string(c.Request.Path())
		logs.UserAgent = string(c.Request.Header.UserAgent())
		logs.Ip = c.ClientIP()
		// ReqContent
		reqBodyStr := string(c.Request.Body())
		if len(reqBodyStr) > 200 {
			reqBodyStr = reqBodyStr[:200]
		}
		logs.ReqContent = reqBodyStr
		// RespContent
		respBodyStr := string(c.Response.Body())
		if len(respBodyStr) > 200 {
			respBodyStr = respBodyStr[:200]
		}
		logs.RespContent = respBodyStr
		// Success
		if c.Response.Header.StatusCode() == 200 {
			logs.Success = true
		}
		// Time
		costTime := time.Since(start).Milliseconds()
		logs.Time = uint32(costTime)

		// username from jwt login cache
		v, exist := c.Get("userID")
		if !exist || v == nil {
			v = "anonymous, no jwt login userinfo in cache"
		}
		var userIDStr string
		var username string
		var ok bool
		userIDStr, ok = v.(string)
		if !ok {
			userIDStr = "anonymous, no jwt login userinfo in cache"
		}
		u, exist := d.Cache.Get("UserID2Username" + userIDStr)
		if !exist || u == nil {
			username = "anonymous, no jwt login userinfo in cache"
		}
		username, ok = u.(string)
		if !ok {
			username = "anonymous, no jwt login userinfo in cache"
		}
		logs.Operator = username

		err := admin.NewLogs(data.Default()).Create(ctx, &logs)
		if err != nil {
			hlog.Error(err)
		}
	}
}
