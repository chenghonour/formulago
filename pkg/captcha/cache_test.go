/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package captcha

import (
	"github.com/patrickmn/go-cache"
	"testing"
	"time"
)

func TestCacheStore_Get(t *testing.T) {
	cacheDB := cache.New(5*time.Minute, 5*time.Minute)
	cacheDB.Set("test", "test", 5*time.Minute)
	type fields struct {
		Expiration time.Duration
		PreKey     string
		Cache      *cache.Cache
	}
	field := fields{
		Expiration: 5 * time.Minute,
		PreKey:     "CAPTCHA_",
		Cache:      cacheDB,
	}
	type args struct {
		key   string
		clear bool
	}
	argWithNotClear := args{
		key: "test",
		// After the get, the cache key will not be cleared
		clear: false,
	}
	argWithClear := args{
		key: "test",
		// After the get, the cache key will be cleared
		clear: true,
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      string
		afterWant string
	}{
		{name: "cacheGet", fields: field, args: argWithNotClear, want: "test", afterWant: "test"},
		{name: "cacheGet", fields: field, args: argWithClear, want: "test", afterWant: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CacheStore{
				Expiration: tt.fields.Expiration,
				PreKey:     tt.fields.PreKey,
				Cache:      tt.fields.Cache,
			}
			if got := c.Get(tt.args.key, tt.args.clear); got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
			// After the get, the cache key should be cleared
			if got := c.Get(tt.args.key, tt.args.clear); got != tt.afterWant {
				t.Errorf("After get key, once again, Get() = %v, want %v", got, tt.afterWant)
			}
		})
	}
}

func TestCacheStore_Set(t *testing.T) {
	cacheDB := cache.New(5*time.Minute, 5*time.Minute)
	type fields struct {
		Expiration time.Duration
		PreKey     string
		Cache      *cache.Cache
	}
	field := fields{
		Expiration: 5 * time.Minute,
		PreKey:     "CAPTCHA_",
		Cache:      cacheDB,
	}
	type args struct {
		id    string
		value string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "cacheSet", fields: field, args: args{id: "test", value: "test"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CacheStore{
				Expiration: tt.fields.Expiration,
				PreKey:     tt.fields.PreKey,
				Cache:      tt.fields.Cache,
			}
			if err := c.Set(tt.args.id, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCacheStore_Verify(t *testing.T) {
	cacheDB := cache.New(5*time.Minute, 5*time.Minute)
	type fields struct {
		Expiration time.Duration
		PreKey     string
		Cache      *cache.Cache
	}
	field := fields{
		Expiration: 5 * time.Minute,
		PreKey:     "CAPTCHA_",
		Cache:      cacheDB,
	}
	cacheDB.Set(field.PreKey+"test", "test", 5*time.Minute)
	type args struct {
		id     string
		answer string
		clear  bool
	}
	arg := args{
		id:     "test",
		answer: "test",
		clear:  true,
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{name: "cacheVerify", fields: field, args: arg, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CacheStore{
				Expiration: tt.fields.Expiration,
				PreKey:     tt.fields.PreKey,
				Cache:      tt.fields.Cache,
			}
			if got := c.Verify(tt.args.id, tt.args.answer, tt.args.clear); got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}
