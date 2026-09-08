/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mcp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"formulago/biz/domain/admin"
	"formulago/biz/handler/middleware"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// authUser is the authenticated caller resolved from the bearer token and
// carried on the request context for the tool handlers.
type authUser struct {
	userID uint64
	roleID uint64
}

type ctxKey int

const authUserKey ctxKey = 0

// requireAuth gates the MCP endpoint: the presented bearer token must pass
// the exact same session/state chain as the HTTP middleware (middleware
// .ValidateTokenString) and its session must carry the "web-mcp" label. This
// closes the bypass of calling tools with an arbitrary browser-login JWT that
// never went through the MCP OAuth (PKCE) flow.
func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr, ok := bearerToken(r)
		if !ok {
			unauthorized(w, "missing bearer token")
			return
		}
		userID, roleID, session, err := middleware.ValidateTokenString(r.Context(), tokenStr)
		if err != nil || session == nil {
			if err != nil && !errors.Is(err, admin.ErrSessionInvalid) {
				hlog.CtxErrorf(r.Context(), "mcp auth: validate token failed: %v", err)
			}
			unauthorized(w, "invalid or expired token")
			return
		}
		if session.Source != middleware.TokenSourceWebMCP {
			hlog.CtxDebugf(r.Context(), "mcp auth: session source %q, want %q", session.Source, middleware.TokenSourceWebMCP)
			unauthorized(w, "token was not issued through the MCP OAuth flow")
			return
		}
		u := &authUser{userID: userID, roleID: roleID}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authUserKey, u)))
	})
}

// unauthorized answers 401 with a JSON-RPC style body and the WWW-Authenticate
// challenge so MCP clients that speak OAuth can (re)start the flow.
func unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("WWW-Authenticate", `Bearer`)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":null,"error":{"code":-32001,"message":"` + message + `"}}`))
}

// bearerToken extracts the raw bearer token from the Authorization header.
func bearerToken(r *http.Request) (string, bool) {
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(h[len(prefix):])
	return token, token != ""
}

// userFromContext returns the authenticated caller of a tool handler.
func userFromContext(ctx context.Context) (*authUser, bool) {
	u, ok := ctx.Value(authUserKey).(*authUser)
	return u, ok
}

// toolGuard enforces the casbin policy of the wrapped API route with the
// caller's CURRENT role (same subject rule as the HTTP authorizator), then
// runs the tool. Denials surface as a tool error so the agent can explain.
func (s *Server) toolGuard(def toolDef) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		u, ok := userFromContext(ctx)
		if !ok {
			return nil, admin.ErrSessionInvalid
		}
		sub := strconv.FormatUint(u.roleID, 10)
		pass, err := s.enforce.Enforce(sub, def.apiPath, def.apiMethod)
		if err != nil {
			hlog.CtxErrorf(ctx, "mcp casbin enforce failed, role: %s path: %s method: %s: %v", sub, def.apiPath, def.apiMethod, err)
			return mcp.NewToolResultError("internal error"), nil
		}
		if !pass {
			hlog.CtxDebugf(ctx, "mcp casbin forbid role: %s path: %s method: %s", sub, def.apiPath, def.apiMethod)
			return mcp.NewToolResultError("no permission: your role cannot access the underlying API " + def.apiPath), nil
		}
		return def.run(ctx, u, req)
	}
}
