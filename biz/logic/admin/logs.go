/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import (
	"context"
	"formulago/biz/domain/admin"
	"formulago/data"
	"formulago/data/ent"
	"formulago/data/ent/logs"
	"formulago/data/ent/predicate"
	"formulago/pkg/types"
	"github.com/pkg/errors"
)

type Logs struct {
	Data *data.Data
}

func NewLogs(data *data.Data) admin.Logs {
	return &Logs{
		Data: data,
	}
}

func (l *Logs) Create(ctx context.Context, logsReq *admin.LogsInfo) error {
	logsReq.Api = types.SubStrByLen(logsReq.Api, 255)
	logsReq.ReqContent = types.SubStrByLen(logsReq.ReqContent, 512)
	logsReq.RespContent = types.SubStrByLen(logsReq.RespContent, 512)
	logsReq.UserAgent = types.SubStrByLen(logsReq.UserAgent, 255)
	err := l.Data.DBClient.Logs.Create().
		SetType(logsReq.Type).
		SetMethod(logsReq.Method).
		SetAPI(logsReq.Api).
		SetSuccess(logsReq.Success).
		SetReqContent(logsReq.ReqContent).
		SetRespContent(logsReq.RespContent).
		SetIP(logsReq.Ip).
		SetUserAgent(logsReq.UserAgent).
		SetOperator(logsReq.Operator).
		SetTime(int(logsReq.Time)).
		Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, "create logs failed")
		return err
	}
	return nil
}

func (l *Logs) List(ctx context.Context, req *admin.LogsListReq) (list []*admin.LogsInfo, total int, err error) {
	var predicates []predicate.Logs
	if req.Type != "" {
		predicates = append(predicates, logs.TypeEQ(req.Type))
	}
	if req.Method != "" {
		predicates = append(predicates, logs.MethodEQ(req.Method))
	}
	if req.Api != "" {
		predicates = append(predicates, logs.APIContains(req.Api))
	}
	if req.Operator != "" {
		predicates = append(predicates, logs.OperatorContains(req.Operator))
	}
	if req.Success != nil {
		predicates = append(predicates, logs.SuccessEQ(*req.Success))
	}
	logsData, err := l.Data.DBClient.Logs.Query().Where(predicates...).
		Offset(int((req.Page - 1) * req.PageSize)).
		Limit(int(req.PageSize)).
		Order(ent.Desc(logs.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err, "query logsData list failed")
	}
	for _, v := range logsData {
		list = append(list, &admin.LogsInfo{
			Type:        v.Type,
			Method:      v.Method,
			Api:         v.API,
			Success:     v.Success,
			ReqContent:  v.ReqContent,
			RespContent: v.RespContent,
			Ip:          v.IP,
			UserAgent:   v.UserAgent,
			Operator:    v.Operator,
			Time:        uint32(v.Time),
			CreatedAt:   v.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   v.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	total, err = l.Data.DBClient.Logs.Query().Where(predicates...).Count(ctx)
	if err != nil {
		return nil, 0, errors.Wrap(err, "query logsData count failed")
	}
	return
}

func (l *Logs) DeleteAll(ctx context.Context) error {
	_, err := l.Data.DBClient.Logs.Delete().Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "delete logsData failed")
	}
	return nil
}
