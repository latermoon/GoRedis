package main

import (
	"./tool"
	"fmt"
	"reflect"
	"testing"
)

var host = ":1602"

func TestHSet(t *testing.T) {
	conn := tool.GetRedisPool(host).Get()
	defer conn.Close()

	fmt.Println("hset", "info", "name", "latermoon")
	reply, err := conn.Do("hset", "info", "name", "latermoon")
	if err != nil {
		t.Error(err)
	}
	if reflect.TypeOf(reply).Kind() != reflect.Int64 {
		t.Errorf("bad reply type %s", reflect.TypeOf(reply))
	}
	fmt.Println("->", reply.(int64))
}

func TestHGet(t *testing.T) {
	conn := tool.GetRedisPool(host).Get()
	defer conn.Close()

	fmt.Println("hget", "info", "name")
	reply, err := conn.Do("hget", "info", "name")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("->", string(reply.([]byte)))
}
