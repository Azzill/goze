/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package cache

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"time"
)

type RedisClient struct {
	conn redis.Conn
}

func NewRedisClient(network string, address string, password string, writeTimeout time.Duration, readTimeout time.Duration,
	connectTimeout time.Duration, db int,
) *RedisClient {
	client := &RedisClient{}
	if address == "" {
		return client
	}
	c, err := redis.Dial(network, address, redis.DialPassword(password),
		redis.DialConnectTimeout(connectTimeout), redis.DialWriteTimeout(writeTimeout),
		redis.DialReadTimeout(readTimeout), redis.DialDatabase(db))
	if err != nil {
		log.Panic(err)
	}
	client.conn = c
	return client
}

func (c *RedisClient) OpsValueSet(key string, value interface{}, ttl time.Duration) bool {
	panic("not implemented")
}

func (c *RedisClient) OpsValueGet(key string) interface{} {
	panic("not implemented")
}

func (c *RedisClient) DeleteKey(key string) bool {
	panic("not implemented")
}

func (c *RedisClient) OpsMapPut(key string, k string, value interface{}) bool {
	panic("not implemented")
}

func (c *RedisClient) OpsMapDel(key string, k string) bool {
	panic("not implemented")
}

func (c *RedisClient) OpsMapGet(key string, k string) interface{} {
	panic("not implemented")
}

func (c *RedisClient) Expire(key string, ttl time.Duration) bool {
	panic("not implemented")
}
