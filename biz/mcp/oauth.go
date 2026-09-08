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
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"formulago/biz/handler/middleware"
	logic "formulago/biz/logic/admin"
	"formulago/pkg/encrypt"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// OAuth endpoint paths, all served under the main Hertz engine.
const (
	PathMCP       = "/api/mcp"
	PathAuthMeta  = PathMCP + "/oauth/metadata"
	PathAuthorize = PathMCP + "/oauth/authorize"
	PathRegister  = PathMCP + "/oauth/register"
	PathExchange  = PathMCP + "/oauth/exchange"
	PathComplete  = PathMCP + "/oauth/complete"
	PathToken     = PathMCP + "/oauth/token"
	// WellKnownMeta is the RFC 8414 discovery path; it mirrors PathAuthMeta.
	WellKnownMeta = "/.well-known/oauth-authorization-server"
)

// OAuth ephemeral store TTLs; records live in the shared memory/Redis cache.
const (
	oauthStateTTL  = 10 * time.Minute
	oauthCodeTTL   = 10 * time.Minute
	oauthClientTTL = 24 * time.Hour
)

// oauthState binds one authorization session initiated by a client.
type oauthState struct {
	ClientID      string `json:"clientID"`
	RedirectURI   string `json:"redirectURI"`
	CodeChallenge string `json:"codeChallenge"`
}

// oauthCode binds an issued code to the browser-login JWT it redeemed.
type oauthCode struct {
	State string `json:"state"`
	Token string `json:"token"`
}

// oauthClient is a minimal RFC 7591 registered client record.
type oauthClient struct {
	RedirectURIs []string `json:"redirectURIs"`
}

func authKey(state string) string { return "mcp:oauth:auth:" + state }
func codeKey(code string) string  { return "mcp:oauth:code:" + code }
func clientKey(id string) string  { return "mcp:oauth:client:" + id }

// readBody reads at most limit bytes of the request body.
func readBody(r *http.Request, limit int64) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r.Body, limit))
}

// randomHex returns n random bytes hex-encoded (crypto/rand).
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// cacheSet/Get/Delete reuse the shared memory + Redis cache of the data layer.
func (s *Server) cacheSet(ctx context.Context, key, val string, ttl time.Duration) error {
	return s.data.CacheSet(ctx, key, val, ttl)
}

func (s *Server) cacheGet(ctx context.Context, key string) (string, bool, error) {
	return s.data.CacheGet(ctx, key)
}

func (s *Server) cacheDel(ctx context.Context, key string) error {
	return s.data.CacheDelete(ctx, key)
}

func storeJSON(ctx context.Context, s *Server, key string, v any, ttl time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.cacheSet(ctx, key, string(b), ttl)
}

func loadJSON[T any](ctx context.Context, s *Server, key string) (v T, ok bool, err error) {
	raw, ok, err := s.cacheGet(ctx, key)
	if err != nil || !ok {
		return v, ok, err
	}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return v, false, err
	}
	return v, true, nil
}

// oauthBaseURL returns the externally visible base of this service: the
// configured ExternalURL (Caddy TLS) when set, otherwise request-derived.
func (s *Server) oauthBaseURL(r *http.Request) string {
	if base := strings.TrimRight(s.cfg.Mcp.ExternalURL, "/"); base != "" && base != "change-me" {
		return base
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func oauthError(w http.ResponseWriter, status int, code, desc string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code, "error_description": desc})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// pkceValid verifies an S256 code challenge against the client's verifier in
// constant time.
func pkceValid(challenge, verifier string) bool {
	sum := sha256.Sum256([]byte(verifier))
	got := base64.RawURLEncoding.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(got), []byte(challenge)) == 1
}

// OAuthMetadataHandler serves RFC 8414 discovery (also mounted at the standard
// /.well-known/oauth-authorization-server path).
func (s *Server) OAuthMetadataHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base := s.oauthBaseURL(r)
		writeJSON(w, http.StatusOK, map[string]any{
			"issuer":                                base,
			"authorization_endpoint":                base + PathAuthorize,
			"token_endpoint":                        base + PathToken,
			"registration_endpoint":                 base + PathRegister,
			"response_types_supported":              []string{"code"},
			"grant_types_supported":                 []string{"authorization_code"},
			"token_endpoint_auth_methods_supported": []string{"none"},
			"code_challenge_methods_supported":      []string{"S256"},
			"scopes_supported":                      []string{},
		})
	}
}

// OAuthRegisterHandler implements a minimal RFC 7591 dynamic registration:
// any public client is accepted (PKCE is mandatory), the record is kept so
// authorize can enforce its redirect URIs.
func (s *Server) OAuthRegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ClientName   string   `json:"client_name"`
			RedirectURIs []string `json:"redirect_uris"`
		}
		body, err := readBody(r, 1<<16)
		if err != nil {
			oauthError(w, http.StatusBadRequest, "invalid_request", "request body too large or unreadable")
			return
		}
		if len(strings.TrimSpace(string(body))) > 0 {
			if err := json.Unmarshal(body, &req); err != nil {
				oauthError(w, http.StatusBadRequest, "invalid_request", "malformed JSON body")
				return
			}
		}
		for _, u := range req.RedirectURIs {
			if !isAbsoluteURL(u) {
				oauthError(w, http.StatusBadRequest, "invalid_redirect_uri", "redirect_uris must be absolute URLs")
				return
			}
		}
		clientID, err := randomHex(16)
		if err != nil {
			hlog.CtxErrorf(r.Context(), "mcp oauth register: random failed: %v", err)
			oauthError(w, http.StatusInternalServerError, "server_error", "internal error")
			return
		}
		if err := storeJSON(r.Context(), s, clientKey(clientID), oauthClient{RedirectURIs: req.RedirectURIs}, oauthClientTTL); err != nil {
			hlog.CtxErrorf(r.Context(), "mcp oauth register: store client failed: %v", err)
			oauthError(w, http.StatusInternalServerError, "server_error", "internal error")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"client_id":                  clientID,
			"client_id_issued_at":        time.Now().Unix(),
			"client_name":                req.ClientName,
			"redirect_uris":              req.RedirectURIs,
			"token_endpoint_auth_method": "none",
			"grant_types":                []string{"authorization_code"},
			"response_types":             []string{"code"},
		})
	}
}

// OAuthAuthorizeHandler starts the authorization flow: it validates the PKCE
// request, stores the state binding, then bounces the browser to the frontend
// bridge route (which performs the actual web login if needed and redeems the
// existing session token at /oauth/exchange).
func (s *Server) OAuthAuthorizeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		state := q.Get("state")
		clientID := q.Get("client_id")
		redirectURI := q.Get("redirect_uri")
		challenge := q.Get("code_challenge")
		method := strings.ToUpper(q.Get("code_challenge_method"))
		fail := func(desc string) {
			hlog.CtxDebugf(r.Context(), "mcp oauth authorize rejected: %s", desc)
			oauthError(w, http.StatusBadRequest, "invalid_request", desc)
		}
		if state == "" || clientID == "" || redirectURI == "" || challenge == "" {
			fail("response_type, client_id, redirect_uri, state and code_challenge are required")
			return
		}
		if q.Get("response_type") != "code" {
			oauthError(w, http.StatusBadRequest, "unsupported_response_type", "only response_type=code is supported")
			return
		}
		if method != "" && method != "S256" {
			oauthError(w, http.StatusBadRequest, "invalid_request", "only code_challenge_method=S256 is supported")
			return
		}
		if !isAbsoluteURL(redirectURI) {
			fail("redirect_uri must be an absolute URI")
			return
		}
		// enforce registered redirect URIs when the client registered; unknown
		// public clients are tolerated (internal LAN trust, PKCE mandatory)
		if rec, ok, err := loadJSON[oauthClient](r.Context(), s, clientKey(clientID)); err != nil {
			hlog.CtxErrorf(r.Context(), "mcp oauth authorize: load client failed: %v", err)
			oauthError(w, http.StatusInternalServerError, "server_error", "internal error")
			return
		} else if ok && !contains(rec.RedirectURIs, redirectURI) {
			fail("redirect_uri does not match the registered client")
			return
		}
		if err := storeJSON(r.Context(), s, authKey(state), oauthState{
			ClientID:      clientID,
			RedirectURI:   redirectURI,
			CodeChallenge: challenge,
		}, oauthStateTTL); err != nil {
			hlog.CtxErrorf(r.Context(), "mcp oauth authorize: store state failed: %v", err)
			oauthError(w, http.StatusInternalServerError, "server_error", "internal error")
			return
		}
		target, err := s.frontendBridgeURL(s.oauthBaseURL(r), state)
		if err != nil {
			hlog.CtxErrorf(r.Context(), "mcp oauth authorize: bad FrontendAuthURL config: %v", err)
			oauthError(w, http.StatusInternalServerError, "server_error", "FrontendAuthURL is not configured")
			return
		}
		http.Redirect(w, r, target, http.StatusFound)
	}
}

// frontendBridgeURL appends state and the MCP base URL to the configured
// frontend bridge route (a hash-routed SPA, so the query goes before '#').
func (s *Server) frontendBridgeURL(base, state string) (string, error) {
	u, err := url.Parse(s.cfg.Mcp.FrontendAuthURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", errors.New("invalid FrontendAuthURL: " + s.cfg.Mcp.FrontendAuthURL)
	}
	fragment := u.Fragment
	u.Fragment = ""
	q := u.Query()
	q.Set("state", state)
	q.Set("mcpBase", base+PathMCP)
	u.RawQuery = q.Encode()
	if fragment != "" {
		u.Fragment = strings.TrimPrefix(fragment, "/")
	}
	return u.String(), nil
}

// OAuthExchangeHandler is the browser-side bridge leg: the frontend POSTs the
// web session's bearer JWT (never in the URL) with the state from the
// authorize redirect and receives a one-time authorization code back. CORS is
// restricted to the configured frontend origins.
func (s *Server) OAuthExchangeHandler() http.HandlerFunc {
	cors := s.corsHandler()
	return func(w http.ResponseWriter, r *http.Request) {
		if cors(w, r) {
			return // OPTIONS preflight answered
		}
		if err := r.ParseForm(); err != nil {
			oauthError(w, http.StatusBadRequest, "invalid_request", "malformed form body")
			return
		}
		state := r.Form.Get("state")
		tokenStr, ok := bearerToken(r)
		if !ok {
			oauthError(w, http.StatusUnauthorized, "invalid_grant", "missing bearer token")
			return
		}
		if _, ok, err := loadJSON[oauthState](r.Context(), s, authKey(state)); err != nil {
			hlog.CtxErrorf(r.Context(), "mcp oauth exchange: load state failed: %v", err)
			oauthError(w, http.StatusInternalServerError, "server_error", "internal error")
			return
		} else if !ok {
			oauthError(w, http.StatusBadRequest, "invalid_grant", "state not found or expired, start the authorization flow again")
			return
		}
		// the token must be a live API session; the session label is upgraded
		// to web-mcp when the client finally redeems the code
		if _, _, _, err := middleware.ValidateTokenString(r.Context(), tokenStr); err != nil {
			oauthError(w, http.StatusUnauthorized, "invalid_grant", "token is invalid or expired, please login on the web first")
			return
		}
		code, err := randomHex(24)
		if err != nil {
			hlog.CtxErrorf(r.Context(), "mcp oauth exchange: random failed: %v", err)
			oauthError(w, http.StatusInternalServerError, "server_error", "internal error")
			return
		}
		if err := storeJSON(r.Context(), s, codeKey(code), oauthCode{State: state, Token: tokenStr}, oauthCodeTTL); err != nil {
			hlog.CtxErrorf(r.Context(), "mcp oauth exchange: store code failed: %v", err)
			oauthError(w, http.StatusInternalServerError, "server_error", "internal error")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"code": code})
	}
}

// corsHandler answers OPTIONS preflight and decorates simple responses with
// Access-Control-Allow-Origin for the configured frontend origins. It reports
// whether the request was fully handled (OPTIONS).
func (s *Server) corsHandler() func(w http.ResponseWriter, r *http.Request) bool {
	origins := make(map[string]bool, len(s.cfg.Mcp.FrontendOrigins))
	for _, o := range s.cfg.Mcp.FrontendOrigins {
		origins[strings.TrimRight(o, "/")] = true
	}
	return func(w http.ResponseWriter, r *http.Request) bool {
		origin := r.Header.Get("Origin")
		allowed := origin == "" || origins[strings.TrimRight(origin, "/")]
		if r.Method == http.MethodOptions {
			if allowed && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Max-Age", "600")
				w.Header().Set("Vary", "Origin")
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
			return true
		}
		if allowed && origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		} else if origin != "" {
			oauthError(w, http.StatusForbidden, "access_denied", "origin is not allowed")
		}
		return false
	}
}

// OAuthCompleteHandler verifies the code/state pair and bounces the browser to
// the client's redirect_uri with the code, exactly once per issued code.
func (s *Server) OAuthCompleteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		code, state := q.Get("code"), q.Get("state")
		rec, ok, err := loadJSON[oauthCode](r.Context(), s, codeKey(code))
		if err != nil {
			hlog.CtxErrorf(r.Context(), "mcp oauth complete: load code failed: %v", err)
			oauthError(w, http.StatusInternalServerError, "server_error", "internal error")
			return
		}
		if !ok || rec.State != state {
			hlog.CtxDebugf(r.Context(), "mcp oauth complete: unknown code/state pair")
			oauthError(w, http.StatusBadRequest, "invalid_grant", "code not found or state mismatch")
			return
		}
		st, ok, err := loadJSON[oauthState](r.Context(), s, authKey(state))
		if err != nil {
			hlog.CtxErrorf(r.Context(), "mcp oauth complete: load state failed: %v", err)
			oauthError(w, http.StatusInternalServerError, "server_error", "internal error")
			return
		}
		if !ok {
			oauthError(w, http.StatusBadRequest, "invalid_grant", "state expired, start the authorization flow again")
			return
		}
		u, err := url.Parse(st.RedirectURI)
		if err != nil {
			hlog.CtxErrorf(r.Context(), "mcp oauth complete: bad stored redirect_uri: %v", err)
			oauthError(w, http.StatusInternalServerError, "server_error", "internal error")
			return
		}
		qq := u.Query()
		qq.Set("code", code)
		qq.Set("state", state)
		u.RawQuery = qq.Encode()
		http.Redirect(w, r, u.String(), http.StatusFound)
	}
}

// OAuthTokenHandler redeems the authorization code with its PKCE verifier and
// issues the client's access token: the very same web JWT that was exchanged
// in the browser bridge, whose session is then relabeled to "web-mcp" so the
// MCP endpoint accepts it. No refresh token is issued in v1.
func (s *Server) OAuthTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			oauthError(w, http.StatusBadRequest, "invalid_request", "malformed form body")
			return
		}
		code := r.Form.Get("code")
		verifier := r.Form.Get("code_verifier")
		clientID := r.Form.Get("client_id")
		if r.Form.Get("grant_type") != "authorization_code" {
			oauthError(w, http.StatusBadRequest, "unsupported_grant_type", "only authorization_code is supported")
			return
		}
		// single use: the code dies here regardless of the outcome below
		rec, ok, err := loadJSON[oauthCode](r.Context(), s, codeKey(code))
		if err != nil {
			hlog.CtxErrorf(r.Context(), "mcp oauth token: load code failed: %v", err)
			oauthError(w, http.StatusInternalServerError, "server_error", "internal error")
			return
		}
		if err := s.cacheDel(r.Context(), codeKey(code)); err != nil {
			hlog.CtxErrorf(r.Context(), "mcp oauth token: delete code failed: %v", err)
		}
		if !ok {
			oauthError(w, http.StatusBadRequest, "invalid_grant", "code not found, expired or already used")
			return
		}
		if verifier == "" {
			oauthError(w, http.StatusBadRequest, "invalid_grant", "code_verifier is required")
			return
		}
		st, ok, err := loadJSON[oauthState](r.Context(), s, authKey(rec.State))
		if err != nil {
			hlog.CtxErrorf(r.Context(), "mcp oauth token: load state failed: %v", err)
			oauthError(w, http.StatusInternalServerError, "server_error", "internal error")
			return
		}
		if !ok {
			oauthError(w, http.StatusBadRequest, "invalid_grant", "authorization state expired")
			return
		}
		if clientID != "" && st.ClientID != "" && clientID != st.ClientID {
			oauthError(w, http.StatusBadRequest, "invalid_grant", "client_id mismatch")
			return
		}
		if given := r.Form.Get("redirect_uri"); given != "" && given != st.RedirectURI {
			oauthError(w, http.StatusBadRequest, "invalid_grant", "redirect_uri mismatch")
			return
		}
		if !pkceValid(st.CodeChallenge, verifier) {
			hlog.CtxDebugf(r.Context(), "mcp oauth token: PKCE verifier mismatch")
			oauthError(w, http.StatusBadRequest, "invalid_grant", "code_verifier does not match")
			return
		}
		uid, _, session, err := middleware.ValidateTokenString(r.Context(), rec.Token)
		if err != nil || session == nil {
			oauthError(w, http.StatusUnauthorized, "invalid_grant", "token is no longer valid, restart the authorization flow")
			return
		}
		// dual-use the browser-login session: relabel web -> web-mcp so the MCP
		// endpoint accepts this exact JWT (idempotent for a repeated flow)
		if err := logic.NewToken(s.data).MarkSource(r.Context(), uid, encrypt.Sha256Hex(rec.Token), middleware.TokenSourceWebMCP); err != nil {
			hlog.CtxErrorf(r.Context(), "mcp oauth token: mark source failed: %v", err)
			oauthError(w, http.StatusInternalServerError, "server_error", "internal error")
			return
		}
		expiresIn := int(time.Until(session.ExpiredAt).Seconds())
		if expiresIn < 1 {
			expiresIn = 1
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"access_token": rec.Token,
			"token_type":   "Bearer",
			"expires_in":   expiresIn,
		})
	}
}

func isAbsoluteURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func contains(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}
