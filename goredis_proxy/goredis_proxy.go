package goredis_proxy

import (
	. "GoRedis/goredis"
)

// Redis高可用+海量存储方案
type GoRedisProxy struct {
	ServerHandler
	RedisServer
}
