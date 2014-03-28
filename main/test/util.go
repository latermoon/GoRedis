package test

import (
	"github.com/latermoon/redigo/redis"
	"time"
)

var host = "localhost:1602"

func NewRedisConn(host string) (redis.Conn, error) {
	return redis.Dial("tcp", host)
}

func RedisPool(host string) (pool *redis.Pool) {
	pool = &redis.Pool{
		MaxIdle:     100,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host)
			return c, err
		},
	}
	return
}
