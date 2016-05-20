package main

import (
	. "GoRedis/libs/shardredis"
	"fmt"
)

func main() {
	fmt.Println("...")
	pool := RedisPool("localhost:6379")
	conn := pool.Get()
	defer conn.Close()
	reply, err := conn.Do("SET", "name", "ko")
	fmt.Println(reply, err)

}

func main2() {
	fmt.Println("start")

	shard0 := &Shard{}
	shard0.Host = "redis_cluster_profile_a0_{0}_s1"
	shard0.Port = 7100
	shard0.Count = 10
	shard0.Start = 0
	shard0.End = 18000000

	shard1 := &Shard{}
	shard1.Host = "redis_cluster_profile_a1_{0}_s1"
	shard1.Port = 7100
	shard1.Count = 10
	shard1.Start = 18000000
	shard1.End = 36000000

	cluster, _ := NewCluster(shard0, shard1)
	rd, _ := cluster.ConnWith("100422")
	defer rd.Close()
	reply, _ := rd.Do("GET", "user:100422:unread")
	fmt.Println(reply)
}
