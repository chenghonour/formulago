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

// Package mcp mounts the MCP protocol endpoint and its OAuth endpoints on the
// main Hertz engine. The mcp-go streamable-HTTP handler is a plain
// http.Handler, bridged with adaptor.HertzHandler (its ResponseWriter
// implements Flush, so the SSE event stream passes through).
package mcp

import (
	bizmcp "formulago/biz/mcp"
	"formulago/configs"
	"formulago/data"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
)

// Register mounts the MCP routes on the Hertz engine:
//   - /api/mcp (GET/POST/DELETE): the MCP protocol endpoint, gated by the MCP
//     auth middleware inside biz/mcp (bearer token whose session carries the
//     "web-mcp" source label);
//   - /api/mcp/oauth/*: the OAuth endpoints used by the browser flow, public;
//   - /.well-known/oauth-authorization-server: RFC 8414 discovery mirror.
func Register(r *server.Hertz) {
	h := bizmcp.New(configs.Data(), data.Default(), data.CasbinEnforcer())
	toHertz := adaptor.HertzHandler

	// protocol endpoint: authenticated MCP streamable-HTTP server
	r.Any(bizmcp.PathMCP, toHertz(h.MCPHandler()))

	// OAuth endpoints (public; authentication happens inside the flow)
	r.Any(bizmcp.PathAuthMeta, toHertz(h.OAuthMetadataHandler()))
	r.Any(bizmcp.WellKnownMeta, toHertz(h.OAuthMetadataHandler()))
	r.Any(bizmcp.PathRegister, toHertz(h.OAuthRegisterHandler()))
	r.Any(bizmcp.PathAuthorize, toHertz(h.OAuthAuthorizeHandler()))
	r.Any(bizmcp.PathExchange, toHertz(h.OAuthExchangeHandler()))
	r.Any(bizmcp.PathComplete, toHertz(h.OAuthCompleteHandler()))
	r.Any(bizmcp.PathToken, toHertz(h.OAuthTokenHandler()))
}
