package shardredis

import (
	"GoRedis/libs/redigo/redis"
	"errors"
	"strconv"
)

// 一个节点集群：维护了一组节点的分片
type Cluster struct {
	shards []*Shard
}

func NewCluster(shards ...*Shard) (c *Cluster, err error) {
	c = &Cluster{
		shards: shards,
	}
	for _, info := range c.shards {
		info.init()
	}
	return
}

func (c *Cluster) ConnWith(hash string) (redis.Conn, error) {
	num, err := strconv.ParseInt(hash, 10, 64)
	if err != nil {
		return nil, err
	}
	var shard *Shard
	for i := 0; i < len(c.shards); i++ {
		shard = c.shards[i]
		if num >= shard.Start && num <= shard.End {
			break
		}
	}
	if shard == nil {
		return nil, errors.New("no match shard")
	}
	return shard.ConnWith(hash), nil
}
