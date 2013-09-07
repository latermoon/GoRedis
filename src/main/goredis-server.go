package main

import (
	"../goredis"
	"fmt"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(2)

	fmt.Println("GoRedis 0.1 by latermoon")
	fmt.Println("Listen 1602")

	server := goredis.NewGoRedisServer()
	server.Listen(":1602")
}
