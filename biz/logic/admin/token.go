/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import (
	"context"
	"fmt"
	"formulago/biz/domain"
	"formulago/data"
	"formulago/data/ent"
	"formulago/data/ent/predicate"
	"formulago/data/ent/token"
	"formulago/data/ent/user"
	"github.com/cockroachdb/errors"
	"time"
)

type Token struct {
	Data *data.Data
}

func NewToken(data *data.Data) domain.Token {
	return &Token{
		Data: data,
	}
}

func (t *Token) Create(ctx context.Context, req *domain.TokenInfo) error {
	expiredAt, _ := time.ParseInLocation("2006-01-02 15:04:05", req.ExpiredAt, time.Local)
	if expiredAt.Sub(time.Now()).Seconds() < 5 {
		return errors.New("expired time must be greater than now, more than 5s")
	}
	// whether user token exists
	tokenExist, err := t.Data.DBClient.Token.Query().Where(token.UserID(req.UserID)).Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return errors.Wrap(err, "create Token failed")
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
		return errors.Wrap(err, "get userinfo failed")
	}

	// create token
	_, err = t.Data.DBClient.Token.Create().
		SetOwner(userInfo).
		SetUserID(req.UserID).
		SetToken(req.Token).
		SetSource(req.Source).
		SetExpiredAt(expiredAt).
		Save(ctx)
	if err != nil {
		return errors.Wrap(err, "create Token failed")
	}

	// add to cache
	t.Data.Cache.Set(fmt.Sprintf("token_%d", req.UserID), req.UserID, expiredAt.Sub(time.Now()))

	return nil
}

func (t *Token) Update(ctx context.Context, req *domain.TokenInfo) error {
	expiredAt, _ := time.ParseInLocation("2006-01-02 15:04:05", req.ExpiredAt, time.Local)
	if expiredAt.Sub(time.Now()).Seconds() < 5 {
		return errors.New("expired time must be greater than now, more than 5s")
	}
	_, err := t.Data.DBClient.Token.UpdateOneID(req.ID).
		SetUserID(req.UserID).
		SetToken(req.Token).
		SetSource(req.Source).
		SetUpdatedAt(time.Now()).
		SetExpiredAt(expiredAt).
		Save(ctx)
	if err != nil {
		return errors.Wrap(err, "update Token failed")
	}

	// add to cache
	t.Data.Cache.Set(fmt.Sprintf("token_%d", req.UserID), req.UserID, expiredAt.Sub(time.Now()))
	return nil
}

func (t *Token) IsExistByUserID(ctx context.Context, userID uint64) bool {
	// get token from cache
	_, exist := t.Data.Cache.Get(fmt.Sprintf("token_%d", userID))
	if exist {
		return true
	}

	// get token from db
	exist, _ = t.Data.DBClient.Token.Query().Where(token.UserID(userID)).Exist(ctx)
	return exist
}

func (t *Token) Delete(ctx context.Context, userID uint64) error {
	_, err := t.Data.DBClient.Token.Delete().Where(token.UserID(userID)).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "delete Token failed")
	}

	// delete from cache
	t.Data.Cache.Delete(fmt.Sprintf("token_%d", userID))
	return nil
}

func (t *Token) List(ctx context.Context, req *domain.TokenListReq) (res []*domain.TokenInfo, total int, err error) {
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
		return res, total, errors.Wrap(err, "get User - Token list failed")
	}

	for _, userEnt := range UserTokens {
		res = append(res, &domain.TokenInfo{
			ID:        userEnt.Edges.Token.ID,
			UserID:    userEnt.ID,
			UserName:  userEnt.Username,
			Token:     userEnt.Edges.Token.Token,
			Source:    userEnt.Edges.Token.Source,
			CreatedAt: userEnt.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: userEnt.UpdatedAt.Format("2006-01-02 15:04:05"),
			ExpiredAt: userEnt.Edges.Token.ExpiredAt.Format("2006-01-02 15:04:05"),
		})

		// delete expired token from db
		if userEnt.Edges.Token.ExpiredAt.Sub(time.Now()).Seconds() < 0 {
			_ = t.Delete(ctx, userEnt.ID)
		}
	}
	total, _ = t.Data.DBClient.User.Query().Where(userPredicates...).Count(ctx)
	return
}
