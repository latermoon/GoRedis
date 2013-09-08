package main

import (
	"../goredis"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"runtime"
	"strconv"
	"strings"
)

var (
	profile_backend []string = []string{"10.80.101.185:7100", "10.80.101.185:7101", "10.80.101.185:7102", "10.80.101.185:7103", "10.80.101.185:7104", "10.80.101.186:7105", "10.80.101.186:7106", "10.80.101.186:7107", "10.80.101.186:7108", "10.80.101.186:7109"}
)
var redisConns []redis.Conn = make([]redis.Conn, 10)

func main() {
	fmt.Println("GoRedis 0.1 by latermoon")
	runtime.GOMAXPROCS(2)

	initRedisConns()

	server := goredis.NewRedisServer()

	server.On("GET", func(cmd *goredis.Command) (reply *goredis.Reply) {
		key := cmd.StringAtIndex(1)
		valueData, e1 := getRedis(key).Do("GET", key)
		fmt.Println(redis.String(valueData, e1))
		if e1 == nil {
			reply = goredis.BulkReply(valueData)
		} else {
			reply = goredis.ErrorReply(e1.Error())
		}
		return
	})

	server.On("SET", func(cmd *goredis.Command) (reply *goredis.Reply) {
		key := cmd.StringAtIndex(1)
		value := cmd.StringAtIndex(2)
		_, e1 := getRedis(key).Do("SET", key, value)
		if e1 == nil {
			reply = goredis.StatusReply("OK")
		} else {
			reply = goredis.ErrorReply(e1.Error())
		}

		return
	})

	// 开始监听端口
	fmt.Println("Listen 8002")
	server.Listen(":8002")
}

func initRedisConns() {
	for i, host := range profile_backend {
		var e error
		redisConns[i], e = redis.Dial("tcp", host)
		if e != nil {
			fmt.Println(e)
		}
	}
}

func getRedis(key string) (conn redis.Conn) {
	parts := strings.Split(key, ":")
	if len(parts) <= 2 {
		panic("Bad Key: " + key)
	}
	momoid := parts[1]
	intMomoid, e1 := strconv.Atoi(momoid)
	if e1 != nil {
		panic(e1)
	}
	i := intMomoid % 10
	conn = redisConns[i]
	return
}
