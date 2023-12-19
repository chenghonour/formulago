/*
 * Copyright 2023 HaiCheng Authors
 *
 * Created by hua
 */

package data

import (
	"fmt"
	"formulago/configs"
	"github.com/redis/go-redis/v9"
)

// Official tutorial  https://redis.uptrace.dev/

func newRedisDB(config configs.Config) (rdb *redis.Client) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	return rdb
}
