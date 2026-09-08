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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// RFC 7636 Appendix B test vectors.
const (
	rfcVerifier  = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	rfcChallenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
)

func TestPKCEValid(t *testing.T) {
	if !pkceValid(rfcChallenge, rfcVerifier) {
		t.Fatal("pkceValid: RFC 7636 vector must verify")
	}
	if pkceValid(rfcChallenge, rfcVerifier+"x") {
		t.Fatal("pkceValid: tampered verifier must be rejected")
	}
	if pkceValid(rfcChallenge+"x", rfcVerifier) {
		t.Fatal("pkceValid: tampered challenge must be rejected")
	}
	if pkceValid("", "") {
		t.Fatal("pkceValid: empty values must be rejected")
	}
}

func TestRandomHex(t *testing.T) {
	a, err := randomHex(24)
	if err != nil {
		t.Fatalf("randomHex failed: %v", err)
	}
	b, err := randomHex(24)
	if err != nil {
		t.Fatalf("randomHex failed: %v", err)
	}
	if len(a) != 48 || a == b {
		t.Fatalf("randomHex: want 48-char unique values, got %q and %q", a, b)
	}
}

func TestBearerToken(t *testing.T) {
	cases := []struct {
		name   string
		header string
		want   string
		ok     bool
	}{
		{"plain bearer", "Bearer abc.def", "abc.def", true},
		{"lowercase scheme", "bearer abc", "abc", true},
		{"whitespace trimmed", "Bearer   abc  ", "abc", true},
		{"no token", "Bearer", "", false},
		{"missing header", "", "", false},
		{"wrong scheme", "Basic abc", "", false},
		{"garbage", "Bearer a b", "a b", true}, // token is whatever follows, non-empty
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", nil)
			if tc.header != "" {
				r.Header.Set("Authorization", tc.header)
			}
			got, ok := bearerToken(r)
			if ok != tc.ok || got != tc.want {
				t.Fatalf("bearerToken(%q) = (%q, %v), want (%q, %v)", tc.header, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestIsAbsoluteURL(t *testing.T) {
	for _, ok := range []string{"https://a.example/x", "claude://oauth/cb", "http://127.0.0.1:8080/cb"} {
		if !isAbsoluteURL(ok) {
			t.Fatalf("isAbsoluteURL(%q) = false, want true", ok)
		}
	}
	for _, bad := range []string{"", "/relative/path", "javascript:alert(1)", "a b c"} {
		if isAbsoluteURL(bad) {
			t.Fatalf("isAbsoluteURL(%q) = true, want false", bad)
		}
	}
}

func TestContains(t *testing.T) {
	if !contains([]string{"a", "b"}, "a") || contains([]string{"a", "b"}, "c") {
		t.Fatal("contains: wrong result")
	}
}

func TestUnauthorizedBodyIsJSONRPCAndChallengesBearer(t *testing.T) {
	w := httptest.NewRecorder()
	unauthorized(w, "missing bearer token")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
	if got := w.Header().Get("WWW-Authenticate"); !strings.Contains(got, "Bearer") {
		t.Fatalf("WWW-Authenticate = %q, want Bearer challenge", got)
	}
	if body := w.Body.String(); !strings.Contains(body, `"jsonrpc":"2.0"`) || !strings.Contains(body, "-32001") {
		t.Fatalf("unexpected body: %s", body)
	}
}
