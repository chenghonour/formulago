/*
 * Copyright 2022 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

// package data provide data model and related methods

package data

import (
	"formulago/configs"
	"formulago/data/ent"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/patrickmn/go-cache"
)

var data *Data

func init() {
	var err error
	data, err = NewData(configs.Data())
	if err != nil {
		hlog.Fatal(err)
	}
}

// Default Get a default database and cache instance
func Default() *Data {
	return data
}

// Data .
type Data struct {
	DBClient *ent.Client
	Cache    *cache.Cache
}

// NewData .
func NewData(config configs.Config) (data *Data, err error) {
	client, err := NewClient(config)
	if err != nil {
		return
	}
	cacheDB := CacheDB("default")
	return &Data{
		DBClient: client,
		Cache:    cacheDB,
	}, nil
}
