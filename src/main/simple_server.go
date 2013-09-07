package main

import (
	"../goredis"
)

type SimpleServer struct {
	goredis.RedisSever
}

func main() {
	server := &SimpleServer()
	server.RedisSever.
}
