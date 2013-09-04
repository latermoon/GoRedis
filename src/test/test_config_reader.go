package main

import (
	"../goredis/util"
	"fmt"
)

func main() {
	config, e1 := util.OpenConfig("test_config_reader.conf")
	if e1 != nil {
		panic(e1)
	}
	fmt.Println("port:", config.IntForKey("port", 1001))
	fmt.Println("cluster", config.StringArrayForKey("redis_profile_cluster"))
	fmt.Println("appendonly", config.BoolForKey("appendonly", false))
}
