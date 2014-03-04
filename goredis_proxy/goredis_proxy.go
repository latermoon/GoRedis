package goredis_proxy

import (
	. "GoRedis/goredis"
)

// Redis高可用+海量存储方案
// GoRedisProxy背后使用Redis实现LRU，GoRedis实现海量存储，对外提供一个实现了高可用和海量存储的的RedisServer端口
// 特性：
// 1、背后的Redis和GoRedis均配置了主从，GoRedisProxy只访问主库，当主库断开，会自动切换到从库，并一直停留在从库，等待人工修正主库，并作为从库添加到集群里
// 2、当Redis里无法找到，会从GoRedis里查找并填充到Redis
// 3、使用SetFillCache fasle指令，使当前连接不访问Redis，全部访问GoRedis，此时扫描全量数据并不会导致LRU清空
type GoRedisProxy struct {
	ServerHandler
	RedisServer
}
