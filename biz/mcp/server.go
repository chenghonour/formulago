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

// Package mcp exposes FormulaGo admin capabilities over the Model Context
// Protocol (streamable HTTP transport, hosted on the main Hertz engine at
// /api/mcp). Authentication reuses the HTTP API's JWT session store: a client
// may call tools only with a bearer token handed out through the OAuth
// authorization-code (PKCE) flow, which upgrades the underlying browser-login
// session's source label to "web-mcp" (the same JWT serves both channels,
// never a newly issued one). Authorization reuses the same casbin policies,
// keyed by the original API route each tool wraps.
package mcp

import (
	"context"
	"net/http"

	"formulago/configs"
	"formulago/data"

	"github.com/casbin/casbin/v3"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server hosts the MCP endpoint and its OAuth endpoints on top of the
// process-wide data layer and casbin enforcer shared with the HTTP API.
type Server struct {
	cfg     configs.Config
	data    *data.Data
	enforce *casbin.Enforcer
	tools   []toolDef
}

// New builds the MCP server. Register its handlers on the Hertz engine via
// biz/router/mcp; see MCPHandler and the OAuth*Handler methods.
func New(cfg configs.Config, d *data.Data, e *casbin.Enforcer) *Server {
	s := &Server{cfg: cfg, data: d, enforce: e}
	s.tools = []toolDef{userListDef(s)}
	return s
}

// MCPHandler returns the authenticated streamable-HTTP handler for /api/mcp.
// Every request must carry a bearer token whose session was upgraded through
// the MCP OAuth flow (see requireAuth); per-tool casbin checks run on call.
func (s *Server) MCPHandler() http.Handler {
	m := server.NewMCPServer(
		"formulago", "0.1.0",
		server.WithToolCapabilities(true),
		server.WithInstructions("FormulaGo 管理后台 MCP 工具。当前提供用户列表查询（user_list）。"),
	)
	for _, def := range s.tools {
		m.AddTool(def.tool, s.toolGuard(def))
	}
	return s.requireAuth(server.NewStreamableHTTPServer(m))
}

// toolDef binds an MCP tool to its underlying API route so the casbin
// permission check is identical to calling that route over HTTP.
type toolDef struct {
	tool               mcp.Tool
	apiPath, apiMethod string
	run                func(ctx context.Context, u *authUser, req mcp.CallToolRequest) (*mcp.CallToolResult, error)
}
