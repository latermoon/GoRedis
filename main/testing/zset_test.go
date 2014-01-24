package main

import (
	. "./tool"
	"fmt"
	"reflect"
	"testing"
)

var host = ":1602"

var pool = RedisPool(host)

var lastT *testing.T

func TestSortedSet(t *testing.T) {
	lastT = t

	key := "zz1"
	checkCommand("del", key)
	checkCommand("zadd", key, "1", "a", "2", "b", "3", "c")
	checkCommand("zscore", key, "b")
	checkCommand("zcard", key)
	checkCommand("zrank", key, "c")
	checkCommand("zrevrank", key, "c")
	checkCommand("zrangebyscore", key, "0", "2")

}

// 发出指令，并打印返回值
func checkCommand(cmdName string, args ...interface{}) {
	conn := pool.Get()
	defer conn.Close()

	reply, err := conn.Do(cmdName, args...)
	if err != nil {
		lastT.Error(err)
	}
	fmt.Print("<--: ", cmdName)
	for _, arg := range args {
		fmt.Print(" ")
		fmt.Print(arg)
	}
	fmt.Println()
	fmt.Print("-->: ")
	switch reply.(type) {
	case []byte:
		fmt.Print("\"", string(reply.([]byte)), "\"")
	default:
		fmt.Print(reply)
	}
	fmt.Printf(" (%s)\n\n", reflect.TypeOf(reply))
}
