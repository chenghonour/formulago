/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package middleware

import (
	"context"
	"formulago/api/model/admin"
	"formulago/configs"
	"github.com/cloudwego/hertz/pkg/app"
)

func ForbidOperation(config configs.Config) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// pre-handle
		if config.IsDemo && (string(c.Request.Method()) == "POST" || string(c.Request.Method()) == "PUT" || string(c.Request.Method()) == "DELETE") {
			resp := new(admin.BaseResp)
			resp.ErrCode = admin.ErrCode_Fail
			resp.ErrMsg = "forbidden operation in demo mode"
			c.JSON(403, resp)
			c.Abort()
			return
		}
		c.Next(ctx)
	}
}
