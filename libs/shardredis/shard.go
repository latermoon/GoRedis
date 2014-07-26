package shardredis

import (
	"GoRedis/libs/redigo/redis"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Shard struct {
	Start int64  // 0
	End   int64  // 1000000000
	Host  string // redis_cluster_profile_c0_{0}_s1.momo.com
	Port  int    // 7600
	Count int    // 10
	pools []*redis.Pool
}

func (s *Shard) init() (err error) {
	s.pools = make([]*redis.Pool, s.Count)
	for i := 0; i < s.Count; i++ {
		func(idx int) {
			s.pools[idx] = &redis.Pool{
				MaxIdle:     100,
				IdleTimeout: 240 * time.Second,
				Dial: func() (redis.Conn, error) {
					host := strings.Replace(s.Host, "{0}", strconv.Itoa(idx), -1)
					addr := fmt.Sprintf("%s:%d", host, s.Port+idx)
					return redis.Dial("tcp", addr)
				},
			}
		}(i)
	}
	return
}

func (s *Shard) ConnWith(hash string) redis.Conn {
	num, _ := strconv.Atoi(hash)
	i := num % s.Count
	return s.pools[i].Get()
}
