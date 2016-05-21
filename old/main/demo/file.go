package main

import (
	"GoRedis/libs/shardredis"
	"fmt"
	"github.com/latermoon/redigo/redis"
	"io/ioutil"
	"os"
)

func main() {
	pool := shardredis.RedisPool(":6379")
	conn := pool.Get()
	defer conn.Close()

	readFromRedis(conn)
}

func readFromRedis(conn redis.Conn) {
	reply, err := conn.Do("GET", "file1")
	if err != nil {
		panic(err)
	}

	data := reply.([]byte)
	if err := ioutil.WriteFile("/Volumes/Private MBP/2.jpg", data, os.ModePerm); err != nil {
		panic(err)
	}
}

func writeToRedis(conn redis.Conn) {
	data, err := ioutil.ReadFile("/Volumes/Private MBP/0tumblr_mz607.jpg")
	if err != nil {
		panic(err)
	}

	if reply, err := conn.Do("SET", "file1", data); err != nil {
		panic(err)
	} else {
		fmt.Println(reply)
	}
}
