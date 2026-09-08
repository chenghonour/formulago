/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package main

import (
	"formulago/biz/handler/middleware"
	mcprouter "formulago/biz/router/mcp"
	"formulago/configs"
	"formulago/data"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// customizeRegister registers customize routers.
func customizedRegister(r *server.Hertz) {
	// your code ...
	// login and refresh_token power by jwt Auth middleware.
	// logout & refresh_token use server-side session revocation: only a token
	// whose digest still lives in sys_tokens can refresh or keep access.
	r.POST("/api/login", middleware.LoginRateLimit(),
		middleware.GetJWTMiddleware(configs.Data(), data.Default(), data.CasbinEnforcer()).LoginHandler)
	r.POST("/api/logout", middleware.LogoutHandler)
	r.POST("/api/refresh_token", middleware.RefreshHandler)

	// MCP endpoint (/api/mcp) and its OAuth browser flow
	mcprouter.Register(r)
}
