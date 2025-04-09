/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package middleware

import (
	"context"
	admin2 "formulago/biz/domain/admin"
	"formulago/biz/logic/admin"
	"formulago/data"
	"formulago/pkg/types"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"strconv"
	"time"
)

func LogsMiddleware(d *data.Data) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// pre-handle
		start := time.Now()
		// ...
		c.Next(ctx)
		// post-handle
		var logs admin2.LogsInfo
		logs.Type = "Interface"
		logs.Method = string(c.Request.Method())
		logs.Api = string(c.Request.Path())
		logs.UserAgent = string(c.Request.Header.UserAgent())
		logs.Ip = c.ClientIP()
		// ReqContent
		reqBodyStr := string(c.Request.Body())
		reqBodyStr = types.SubStrByLen(reqBodyStr, 200)
		logs.ReqContent = reqBodyStr
		// RespContent
		respBodyStr := string(c.Response.Body())
		respBodyStr = types.SubStrByLen(respBodyStr, 200)
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
			v = "0"
		}
		var userIDStr string
		var username = "anonymous"
		var ok bool
		userIDStr, ok = v.(string)
		if !ok {
			userIDStr = "0"
		}
		userID, _ := strconv.Atoi(userIDStr)
		userInfo, _ := admin.NewUser(d).UserInfo(ctx, uint64(userID))
		if userInfo != nil {
			username = userInfo.Username
		}
		logs.Operator = username

		err := admin.NewLogs(data.Default()).Create(ctx, &logs)
		if err != nil {
			hlog.Error(err)
		}
	}
}
