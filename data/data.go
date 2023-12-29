/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

// package data provide data model and related methods

package data

import (
	"context"
	"time"

	"formulago/configs"
	"formulago/data/ent"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
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
	Redis    *redis.Client
	Cache    *cache.Cache
}

// NewData .
func NewData(config configs.Config) (data *Data, err error) {
	data = new(Data)

	client, err := NewClient(config)
	if err != nil {
		return
	}
	data.DBClient = client

	if config.Redis.Enable {
		rdb := newRedisDB(config)
		data.Redis = rdb
	}

	cacheDB := CacheDB("default")
	data.Cache = cacheDB

	return data, nil
}

// CacheSet . Set cache data to memory cache and redis(if enable)
func (d *Data) CacheSet(ctx context.Context, k, v string, exp time.Duration) (err error) {
	d.Cache.Set(k, v, exp)

	if d.Redis != nil {
		err = d.Redis.Set(ctx, k, v, exp).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// CacheGet . Get cache data from memory cache first and redis(if enable)
func (d *Data) CacheGet(ctx context.Context, k string) (v string, exist bool, err error) {
	value, ok := d.Cache.Get(k)
	if ok {
		if str, ok := value.(string); ok {
			return str, true, nil
		}
	}

	if d.Redis != nil {
		v, err = d.Redis.Get(ctx, k).Result()
		// redis.Nil is returned when the key does not exist
		if err != nil && err != redis.Nil {
			return "", false, err
		}
		if err == redis.Nil {
			return "", false, nil
		}
		return v, true, nil
	}

	return "", false, nil
}

// CacheDelete . Delete cache data from memory cache and redis(if enable)
func (d *Data) CacheDelete(ctx context.Context, k string) (err error) {
	d.Cache.Delete(k)

	if d.Redis != nil {
		err = d.Redis.Del(ctx, k).Err()
		if err != nil {
			return err
		}
	}
	return nil
}
