package tool

import (
	"github.com/latermoon/redigo/redis"
	"time"
)

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
