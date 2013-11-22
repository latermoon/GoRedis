package main

import (
	"fmt"
	"github.com/latermoon/redigo/redis"
)

func main() {
	allkeys()
}

func allkeys() {
	conn, err := redis.Dial("tcp", "goredis-profile-c001:17600")
	if err != nil {
		panic(err)
	}
	lastkey := ""
	lasttype := ""
	idx := 1
	for {
		reply, _ := conn.Do("key_next", lastkey, 24)
		if reply == nil {
			break
		}
		lst := reply.([]interface{})
		count := len(lst)
		for i := 0; i < count; i += 2 {
			lastkey = string(lst[i].([]byte))
			lasttype = string(lst[i+1].([]byte))
			fmt.Println(idx, lastkey, lasttype)
			idx++
		}
		if count == 0 {
			break
		}
	}
}
