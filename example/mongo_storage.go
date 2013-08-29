package main

import (
	"../goredis"
	"../goredis/storage"
	"fmt"
	"runtime"
)

func main() {
	fmt.Println("GoRedis 0.1 by latermoon")
	runtime.GOMAXPROCS(2)

	server := goredis.NewSimpleRedisServer()

	mgs := storage.NewMongoStorage()
	if e1 := mgs.Connect("mongodb://172.16.9.14:27017/goredis"); e1 != nil {
		panic(e1)
	}
	defer mgs.Close()

	// 将MongoDB存储buffer起来提高性能，一般mongo的set操作5k/2，用bufferSize=100，可以变为1w/s，
	// 如果bufferSize大于瞬间要操作的指令数，可以达到cache版本的10w/s
	storage := storage.NewBufferedStorage(mgs, 100)

	server.OnSET = func(key string, value string) (err error) {
		err = storage.Set(key, value)
		return
	}

	server.OnGET = func(key string) (value interface{}) {
		value, _ = storage.Get(key)
		return
	}

	fmt.Println("Listen 8002")
	server.Listen(":8002")
}
