package shardredis

import (
	"GoRedis/libs/redigo/redis"
	"sync"
	"time"
)

/*
Usage 1:
pool := shardredis.RedisPool("localhost:6379")
conn := pool.Get()
defer conn.Close()

Usage 2:
cluster := shardredis.Cluster("redis-profile-b")
rd := cluster.Get("100422")
reply, err := rd.Do("SET", "name", "latermoon")
rd.Close()
*/

var mu sync.Mutex
var pools = map[string]*redis.Pool{}

func RedisPool(host string) (pool *redis.Pool) {
	var ok bool
	if pool, ok = pools[host]; !ok {
		mu.Lock()
		defer mu.Unlock()
		if pool, ok = pools[host]; !ok {
			pool = &redis.Pool{
				MaxIdle:     100,
				IdleTimeout: 240 * time.Second,
				Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", host) },
			}
			pools[host] = pool
		}
	}
	return
}
