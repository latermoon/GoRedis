package main

import (
	//"github.com/latermoon/GoRedis/goredis"
	"../goredis"
	"fmt"
)

func main() {
	fmt.Println("GoRedis 0.1 by latermoon")

	server := goredis.NewSimpleRedisServer()

	// KeyValue
	kvCache := make(map[string]interface{})
	// Set操作的写锁
	chanSet := make(chan int, 1)

	server.OnGET = func(key string) (value interface{}) {
		value = kvCache[key]
		return
	}

	server.OnSET = func(key string, value string) (err error) {
		err = nil
		chanSet <- 0
		kvCache[key] = value
		<-chanSet
		return
	}

	server.OnDEL = func(keys ...string) (count int) {
		chanSet <- 0
		for _, key := range keys {
			delete(kvCache, key)
		}
		<-chanSet
		count = len(keys)
		return
	}

	server.OnMGET = func(keys ...string) (bulks []interface{}) {
		bulks = make([]interface{}, len(keys))
		for i, key := range keys {
			bulks[i], _ = kvCache[key]
		}
		return
	}

	fmt.Println("Listen :8002")
	server.Listen(":8002")
}
