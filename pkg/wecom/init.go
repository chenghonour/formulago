/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */
// Package wecom provides a client for accessing the WeCom API.(企业微信)

package wecom

import (
	"formulago/configs"
	"formulago/data"
	"net/http"
	"sync"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/work"
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
var client = http.DefaultClient

// oauthCallback is required by work.NewWork but unused by our flow:
// the authorize URL is hand-built in biz/logic/admin/oauth.go.
const oauthCallback = "http://localhost:3100/oauth/login/callback"

// appOnce guards the lazy init of the shared PowerWeChat Work app.
// configs.Config and data.Data are application-level singletons, so
// initializing from the first Wecom instance is safe; sharing the app
// lets PowerWeChat cache the access token across requests.
var (
	appOnce sync.Once
	app     *work.Work
	appErr  error
)

// app returns the shared PowerWeChat Work instance (lazy, initialized once).
func (w *Wecom) app() (*work.Work, error) {
	appOnce.Do(func() {
		app, appErr = initApp(w)
	})
	return app, appErr
}

func initApp(w *Wecom) (*work.Work, error) {
	config := &work.UserConfig{
		CorpID:  w.Config.Wecom.CorpID,
		AgentID: w.Config.Wecom.AgentID,
		Secret:  w.Config.Wecom.SecretID,
		Token:   w.Config.Wecom.Token,
		AESKey:  w.Config.Wecom.EncodingAESKey,
		OAuth: work.OAuth{
			Callback: oauthCallback,
			Scopes:   []string{"snsapi_base"},
		},
	}
	// share the project Redis for access token caching (falls back to
	// PowerWeChat's in-memory cache when Redis is disabled)
	if w.Data != nil && w.Data.Redis != nil {
		opts := w.Data.Redis.Options()
		config.Cache = kernel.NewRedisClient(&kernel.UniversalOptions{
			Addrs:    []string{opts.Addr},
			Password: opts.Password,
			DB:       opts.DB,
		})
	}
	return work.NewWork(config)
}
