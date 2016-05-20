package test

import (
	"github.com/latermoon/redigo/redis"
	"testing"
	"time"
)

// 硬编码全局地址
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

// 通用redis性能测试模版
// before用于初始化数据
// fn执行单条测试
func benchmark(b *testing.B, before func(conn redis.Conn), fn func(conn redis.Conn) error) {
	b.StopTimer()
	// init 1
	conn, err := NewRedisConn(host)
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		b.StopTimer()
		conn.Close()
	}()

	// init 2
	before(conn)
	b.StartTimer()

	// start
	for i := 0; i < b.N; i++ {
		err := fn(conn)
		if err != nil {
			b.Error(err)
		}
	}
}
