/*
 * Copyright 2022 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package main

import (
	"formulago/biz/handler/middleware"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// customizeRegister registers customize routers.
func customizedRegister(r *server.Hertz) {
	// your code ...
	// login and refresh_token power by jwt Auth middleware
	r.POST("/api/login", middleware.GetJWTMiddleware().LoginHandler)
	r.POST("/api/logout", middleware.GetJWTMiddleware().LogoutHandler)
	r.POST("/api/refresh_token", middleware.GetJWTMiddleware().RefreshHandler)
}
