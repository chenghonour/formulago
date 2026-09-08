/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package middleware

import (
	"context"
	"net/http"
	"time"

	"formulago/pkg/ratelimit"

	"github.com/cloudwego/hertz/pkg/app"
)

// loginLimiter is shared by every login-ish route so that the budgets of
// /api/login and /api/refresh_token add up per client IP.
var loginLimiter = ratelimit.New(ratelimit.Config{
	MaxRequests: 30,               // per Window
	Window:      time.Minute,      // counting window
	MaxFailures: 10,               // consecutive-window failures before lockout
	Lockout:     30 * time.Minute, // applied once the budget is exhausted
})

// LoginRateLimit throttles credential endpoints: bursty requests are cut off
// and repeated failures (HTTP 401) lock the client IP out for Lockout.
func LoginRateLimit() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		ip := c.ClientIP()
		if !loginLimiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, map[string]any{
				"code":    http.StatusTooManyRequests,
				"message": "too many requests, please retry later",
			})
			c.Abort()
			return
		}
		c.Next(ctx)
		// any outcome other than 200 counts as a failure; a 200 (or a lockout
		// bypass elsewhere) clears the streak
		loginLimiter.Report(ip, c.Response.StatusCode() == http.StatusOK)
	}
}
