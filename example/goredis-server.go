package main

import (
	//"../goredis_server"
	"../goredis_server"
	"flag"
	"fmt"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(2)
	//flag
	portPtr := flag.Int("p", 1602, "Server port")
	flag.Parse()

	fmt.Println("GoRedis 0.1 by latermoon")
	fmt.Println("Listen", *portPtr)

	server := goredis_server.NewGoRedisServer()
	host := fmt.Sprintf(":%d", *portPtr)
	server.Listen(host)
}
