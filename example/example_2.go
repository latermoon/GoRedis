package main

import (
	"../goredis"
	"fmt"
)

func main() {
	server := goredis.NewRedisServerEx()
	server.OnGet = func(s *goredis.Session, key string) (value interface{}, err error) {
		value = key
		err = nil
		return
	}

	server.OnSet = func(s *goredis.Session, key string, value string) (err error) {
		err = nil
		return
	}
	fmt.Println("listen :8002")
	server.Listen(":8002")

}
