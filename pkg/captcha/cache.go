/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

// package captcha provide captcha cache store related methods

package captcha

import (
	"time"

	"github.com/patrickmn/go-cache"
)

func NewCacheStore(cache *cache.Cache) *CacheStore {
	return &CacheStore{
		Expiration: time.Minute * 5,
		PreKey:     "CAPTCHA_",
		Cache:      cache,
	}
}

type CacheStore struct {
	Expiration time.Duration
	PreKey     string
	Cache      *cache.Cache
}

func (c *CacheStore) Set(id string, value string) error {
	c.Cache.Set(c.PreKey+id, value, c.Expiration)
	return nil
}

func (c *CacheStore) Get(key string, clear bool) string {
	val, exist := c.Cache.Get(key)
	if !exist {
		return ""
	}
	if clear {
		c.Cache.Delete(key)
	}
	return val.(string)
}

func (c *CacheStore) Verify(id, answer string, clear bool) bool {
	key := c.PreKey + id
	v := c.Get(key, clear)
	return v == answer
}
