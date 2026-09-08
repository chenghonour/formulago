/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import (
	"context"
	"errors"
	"fmt"
	"formulago/biz/domain/admin"
	"formulago/data"
	"formulago/data/ent"
	"formulago/data/ent/predicate"
	"formulago/data/ent/token"
	"formulago/data/ent/user"
	"formulago/pkg/encrypt"
	"formulago/pkg/times"
	"time"
)

type Token struct {
	Data *data.Data
}

func NewToken(data *data.Data) admin.Token {
	return &Token{
		Data: data,
	}
}

// parseExpiredAt parses and validates the ExpiredAt string of a TokenInfo.
func parseExpiredAt(req *admin.TokenInfo) (time.Time, error) {
	expiredAt, err := time.ParseInLocation(times.TimeFormat, req.ExpiredAt, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse expiredAt failed: %w", err)
	}
	if time.Until(expiredAt).Seconds() < 5 {
		return time.Time{}, errors.New("expired time must be greater than now, more than 5s")
	}
	return expiredAt, nil
}

// Create stores or refreshes the single login session of a user. req.Token
// is the raw JWT: its SHA-256 digest is stored, never the token itself.
func (t *Token) Create(ctx context.Context, req *admin.TokenInfo) error {
	expiredAt, err := parseExpiredAt(req)
	if err != nil {
		return err
	}
	// whether user token exists
	tokenExist, err := t.Data.DBClient.Token.Query().Where(token.UserID(req.UserID)).Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return fmt.Errorf("create Token failed: %w", err)
		}
	}
	// if exists, update token
	if tokenExist != nil {
		req.ID = tokenExist.ID
		return t.Update(ctx, req)
	}
	// userinfo
	userInfo, err := t.Data.DBClient.User.Query().Where(user.IDEQ(req.UserID)).Only(ctx)
	if err != nil {
		return fmt.Errorf("get userinfo failed: %w", err)
	}

	// create token
	_, err = t.Data.DBClient.Token.Create().
		SetOwner(userInfo).
		SetUserID(req.UserID).
		SetToken(encrypt.Sha256Hex(req.Token)).
		SetSource(req.Source).
		SetExpiredAt(expiredAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create Token failed: %w", err)
	}
	return nil
}

// Update refreshes an existing session row (admin edit path). req.Token, when
// non-empty, is the raw credential and is stored as its SHA-256 digest.
func (t *Token) Update(ctx context.Context, req *admin.TokenInfo) error {
	expiredAt, err := parseExpiredAt(req)
	if err != nil {
		return err
	}
	update := t.Data.DBClient.Token.UpdateOneID(req.ID).
		SetUserID(req.UserID).
		SetSource(req.Source).
		SetUpdatedAt(time.Now()).
		SetExpiredAt(expiredAt)
	// an empty token field means "keep the current credential", so admins can
	// extend/limit a session without knowing the issued JWT
	if req.Token != "" {
		update = update.SetToken(encrypt.Sha256Hex(req.Token))
	}
	_, err = update.Save(ctx)
	if err != nil {
		return fmt.Errorf("update Token failed: %w", err)
	}
	return nil
}

// ActiveSession validates that userID has a not-yet-expired session whose
// stored digest equals tokenHash. The session row is the source of truth
// (no cache) so a Delete is effective immediately on every instance.
func (t *Token) ActiveSession(ctx context.Context, userID uint64, tokenHash string) (*admin.TokenSession, error) {
	session, err := t.Data.DBClient.Token.Query().Where(token.UserIDEQ(userID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, admin.ErrSessionInvalid
		}
		return nil, fmt.Errorf("get Token failed: %w", err)
	}
	if !session.ExpiredAt.After(time.Now()) || session.Token != tokenHash {
		return nil, admin.ErrSessionInvalid
	}
	return &admin.TokenSession{ExpiredAt: session.ExpiredAt}, nil
}

// Rotate replaces the digest of the current session only when it still equals
// oldHash, so a refresh cannot resurrect a session that was revoked or already
// rotated away by a concurrent login.
func (t *Token) Rotate(ctx context.Context, userID uint64, oldHash, newHash string) error {
	affected, err := t.Data.DBClient.Token.Update().
		Where(token.UserIDEQ(userID), token.TokenEQ(oldHash)).
		SetToken(newHash).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("rotate Token failed: %w", err)
	}
	if affected == 0 {
		return admin.ErrSessionInvalid
	}
	return nil
}

func (t *Token) Delete(ctx context.Context, userID uint64) error {
	_, err := t.Data.DBClient.Token.Delete().Where(token.UserID(userID)).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete Token failed: %w", err)
	}
	// best-effort cleanup of legacy cache keys
	_ = t.Data.CacheDelete(ctx, fmt.Sprintf("token_%d", userID))
	return nil
}

func (t *Token) List(ctx context.Context, req *admin.TokenListReq) (res []*admin.TokenInfo, total int, err error) {
	// list token with user info
	var userPredicates = []predicate.User{user.HasToken()}
	if req.Username != "" {
		userPredicates = append(userPredicates, user.UsernameContainsFold(req.Username))
	}
	if req.UserID != 0 {
		userPredicates = append(userPredicates, user.IDEQ(req.UserID))
	}
	UserTokens, err := t.Data.DBClient.User.Query().Where(userPredicates...).
		WithToken(func(q *ent.TokenQuery) {
			// get token all fields default, or use q.Select() to get some fields
		}).Offset(int(req.Page-1) * int(req.PageSize)).
		Limit(int(req.PageSize)).All(ctx)
	if err != nil {
		return res, total, fmt.Errorf("get User - Token list failed: %w", err)
	}

	for _, userEnt := range UserTokens {
		res = append(res, &admin.TokenInfo{
			ID:        userEnt.Edges.Token.ID,
			UserID:    userEnt.ID,
			UserName:  userEnt.Username,
			Token:     userEnt.Edges.Token.Token,
			Source:    userEnt.Edges.Token.Source,
			CreatedAt: userEnt.CreatedAt.Format(times.TimeFormat),
			UpdatedAt: userEnt.UpdatedAt.Format(times.TimeFormat),
			ExpiredAt: userEnt.Edges.Token.ExpiredAt.Format(times.TimeFormat),
		})

		// delete expired token from db
		if time.Until(userEnt.Edges.Token.ExpiredAt).Seconds() < 0 {
			_ = t.Delete(ctx, userEnt.ID)
		}
	}
	total, err = t.Data.DBClient.User.Query().Where(userPredicates...).Count(ctx)
	if err != nil {
		err = fmt.Errorf("count token failed: %w", err)
		return
	}
	return
}
