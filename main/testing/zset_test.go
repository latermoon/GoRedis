package main

import (
	"./tool"
	"fmt"
	"reflect"
	"testing"
)

var host = ":1602"

func TestZAdd(t *testing.T) {
	conn := tool.GetRedisPool(host).Get()
	defer conn.Close()

	fmt.Println("zadd", "zz", "1", "a", "2", "b", "101", "c")
	reply, err := conn.Do("zadd", "zz", "1", "a", "2", "b", "101", "c")
	if err != nil {
		t.Error(err)
	}
	if reflect.TypeOf(reply).Kind() != reflect.Int64 {
		t.Errorf("bad reply type %s", reflect.TypeOf(reply))
	}
	fmt.Println("->", reply.(int64))
}
