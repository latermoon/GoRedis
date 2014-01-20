package main

import (
	. "./tool"
	"fmt"
	"reflect"
	"testing"
)

var host = ":1602"

func TestSortedSet(t *testing.T) {
	conn := RedisPool(host).Get()
	defer conn.Close()

	key := "zz1"
	reply, err := conn.Do("del", key)
	CheckErr(t, err)
	reply, err = conn.Do("zadd", key, "1", "a", "2", "b", "3", "c")
	CheckErr(t, err)
	CheckType(t, reply, reflect.Int64)
	fmt.Println("->", reply.(int64))

	reply, err = conn.Do("zscore", key, "b")
	CheckErr(t, err)
	fmt.Println(reply, reflect.TypeOf(reply))
}
