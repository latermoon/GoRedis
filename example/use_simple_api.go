package main

import (
	//"github.com/latermoon/GoRedis/goredis"
	"../goredis"
)

func main() {

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

	server.Listen(":8002")
}
