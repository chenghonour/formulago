/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import (
	"context"
	"errors"
	"time"
)

// ErrSessionInvalid is returned when no active session matches: the token
// row is missing, expired, or the credential hash does not match.
var ErrSessionInvalid = errors.New("token session is invalid")

// TokenSession is an active login session of a user.
type TokenSession struct {
	// ExpiredAt is the absolute session deadline. A refresh never extends it,
	// so a session lives at most AccessExpire after its first issue.
	ExpiredAt time.Time
	// Source is the session channel label: "web" for a normal web login;
	// "web-mcp" once the same token was also handed out through the MCP OAuth
	// flow (the browser-login token is dual-used, never re-issued).
	Source string
}

// Token manages the server-side login session of a user. The row stores the
// SHA-256 digest of the current JWT (never the raw token), so only the exact
// token issued by the latest login can pass ActiveSession.
type Token interface {
	Create(ctx context.Context, req *TokenInfo) error
	Update(ctx context.Context, req *TokenInfo) error
	// ActiveSession returns the session when a row exists whose digest equals
	// tokenHash and which is not yet expired, ErrSessionInvalid otherwise.
	ActiveSession(ctx context.Context, userID uint64, tokenHash string) (*TokenSession, error)
	// Rotate atomically replaces oldHash with newHash for the user's session.
	// It fails with ErrSessionInvalid when the session was revoked or already
	// rotated by a concurrent login/refresh (compare-and-swap on the digest).
	Rotate(ctx context.Context, userID uint64, oldHash, newHash string) error
	// MarkSource relabels the user's current session source (e.g. "web" →
	// "web-mcp" when the browser-login token is additionally handed out through
	// the MCP OAuth flow). CAS'd on the digest; fails with ErrSessionInvalid
	// when the session was revoked or rotated concurrently.
	MarkSource(ctx context.Context, userID uint64, tokenHash, source string) error
	Delete(ctx context.Context, userID uint64) error
	List(ctx context.Context, req *TokenListReq) (res []*TokenInfo, total int, err error)
}

type TokenInfo struct {
	ID        uint64
	CreatedAt string
	UpdatedAt string
	UserID    uint64
	UserName  string
	Token     string // SHA-256 digest of the JWT, never the raw token
	Source    string
	ExpiredAt string
}

type TokenListReq struct {
	Page     uint64
	PageSize uint64
	Username string
	UserID   uint64
}
