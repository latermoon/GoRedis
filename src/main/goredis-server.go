package main

import (
	"../goredis"
	grutil "../goredis/util"
	"flag"
	"fmt"
)

func main() {
	fmt.Println("GoRedis 0.1 by latermoon")
	portPtr := flag.Int("port", 1602, "Server port (default: 1602)")
	configPtr := flag.String("config", "goredis.conf", "/path/to/goredis.conf")
	flag.Parse()

	_, e1 := grutil.OpenConfig(*configPtr)
	if e1 != nil {
		panic(e1)
	}

	server := goredis.NewGoRedisServer()

	fmt.Printf("Listen :%d\n", *portPtr)
	server.Listen(fmt.Sprintf(":%d", *portPtr))
}
