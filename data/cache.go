/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package data

import (
	"time"

	"github.com/patrickmn/go-cache"
)

// CacheMap .
var cacheMap = make(map[string]*cache.Cache)

// newCacheDB . Create a new cache db instance
func newCacheDB(DBName string, defaultExpiration, cleanupInterval time.Duration) *cache.Cache {
	db := cache.New(defaultExpiration, cleanupInterval)
	cacheMap[DBName] = db
	return db
}

// CacheDB . Get a cache db instance
func CacheDB(cacheDB string) *cache.Cache {
	if db, ok := cacheMap[cacheDB]; ok {
		return db
	}

	// default expiration time is 10 minutes, cleanup interval is 15 minute
	db := newCacheDB(cacheDB, 10*time.Minute, 15*time.Minute)
	return db
}
