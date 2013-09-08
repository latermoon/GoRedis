package main

import (
	//"../goredis_server"
	"fmt"
	"github.com/latermoon/GoRedis/goredis_server"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(2)

	fmt.Println("GoRedis 0.1 by latermoon")
	fmt.Println("Listen 1602")

	server := goredis_server.NewGoRedisServer()
	server.Listen(":1602")
}
