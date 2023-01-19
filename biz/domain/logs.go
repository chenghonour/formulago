/*
 * Copyright 2022 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package domain

import "context"

type Logs interface {
	Create(ctx context.Context, logsReq *LogsInfo) error
	List(ctx context.Context, req *LogsListReq) (list []*LogsInfo, total int, err error)
	DeleteAll(ctx context.Context) error
}

type LogsInfo struct {
	Type        string
	Method      string
	Api         string
	Success     bool
	ReqContent  string
	RespContent string
	Ip          string
	UserAgent   string
	Operator    string
	Time        uint32
	CreatedAt   string
	UpdatedAt   string
}

type LogsListReq struct {
	Page     uint64
	PageSize uint64
	Type     string
	Method   string
	Api      string
	Success  *bool
	Operator string
}
