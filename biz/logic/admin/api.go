/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

// package logic/admin provides the business logic for admin.

package admin

import (
	"context"
	"formulago/biz/domain"
	"formulago/data"
	"formulago/data/ent/api"
	"formulago/data/ent/predicate"
	"github.com/cockroachdb/errors"
)

type Api struct {
	Data *data.Data
}

func NewApi(data *data.Data) domain.Api {
	return &Api{
		Data: data,
	}
}

func (a *Api) Create(ctx context.Context, req domain.ApiInfo) error {
	_, err := a.Data.DBClient.API.Create().
		SetPath(req.Path).
		SetDescription(req.Description).
		SetAPIGroup(req.Group).
		SetMethod(req.Method).
		Save(ctx)
	if err != nil {
		err = errors.Wrap(err, "create Api failed")
		return err
	}
	return nil
}

func (a *Api) Update(ctx context.Context, req domain.ApiInfo) error {
	_, err := a.Data.DBClient.API.UpdateOneID(req.ID).
		SetPath(req.Path).
		SetDescription(req.Description).
		SetAPIGroup(req.Group).
		SetMethod(req.Method).
		Save(ctx)
	if err != nil {
		err = errors.Wrap(err, "update Api failed")
		return err
	}
	return nil
}

func (a *Api) Delete(ctx context.Context, id uint64) error {
	err := a.Data.DBClient.API.DeleteOneID(id).Exec(ctx)
	return err
}

func (a *Api) List(ctx context.Context, req domain.ListApiReq) (resp []*domain.ApiInfo, total int, err error) {
	var predicates []predicate.API
	if req.Path != "" {
		predicates = append(predicates, api.PathContains(req.Path))
	}
	if req.Description != "" {
		predicates = append(predicates, api.DescriptionContains(req.Description))
	}
	if req.Method != "" {
		predicates = append(predicates, api.MethodContains(req.Method))
	}
	if req.Group != "" {
		predicates = append(predicates, api.APIGroupContains(req.Group))
	}

	apis, err := a.Data.DBClient.API.Query().Where(predicates...).
		Offset(int(req.Page-1) * int(req.PageSize)).
		Limit(int(req.PageSize)).All(ctx)
	if err != nil {
		err = errors.Wrap(err, "get api list failed")
		return resp, total, err
	}
	for _, a := range apis {
		resp = append(resp, &domain.ApiInfo{
			ID:          a.ID,
			CreatedAt:   a.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   a.UpdatedAt.Format("2006-01-02 15:04:05"),
			Path:        a.Path,
			Description: a.Description,
			Group:       a.APIGroup,
			Method:      a.Method,
		})
	}
	total, _ = a.Data.DBClient.API.Query().Where(predicates...).Count(ctx)
	return resp, total, nil
}
