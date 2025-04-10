/*
 * Copyright 2023 FormulaGo Authors
 *
 * Created by hua
 */
// Package wecom provides a client for accessing the WeCom API.(企业微信)

package wecom

import (
	"formulago/configs"
	"formulago/data"
	"github.com/imroc/req/v3"
)

func New(config configs.Config, data *data.Data) *Wecom {
	return &Wecom{
		Config: config,
		Data:   data,
	}
}

// Wecom 企业微信操作，包含企业配置参数
type Wecom struct {
	Config configs.Config
	Data   *data.Data
}

// http client
var client = req.C()
