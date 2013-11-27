package main

import (
	"./tool"
	"fmt"
	"reflect"
	"testing"
)

var host = ":1602"

func TestSet(t *testing.T) {
	conn := tool.GetRedisPool(host).Get()
	defer conn.Close()

	cmds := [][]interface{}{
		[]interface{}{"set", "name", "latermoon"},
		[]interface{}{"set", "age", "27"},
		[]interface{}{"set", "sex", "M"},
	}

	for _, cmd := range cmds {
		fmt.Println(cmd...)
		reply, err := conn.Do(cmd[0].(string), cmd[1:]...)
		if err != nil {
			t.Error(err)
		}
		fmt.Println("->", reply)
	}
}

func TestGet(t *testing.T) {
	conn := tool.GetRedisPool(host).Get()
	defer conn.Close()

	names := []string{"name", "age", "sex"}
	for _, name := range names {
		fmt.Println("get", name)
		reply, err := conn.Do("get", name)
		if err != nil {
			t.Error(err)
		}
		fmt.Println("->", string(reply.([]byte)))
	}
}

func TestIncr(t *testing.T) {
	conn := tool.GetRedisPool(host).Get()
	defer conn.Close()

	fmt.Println("incr age")
	reply, err := conn.Do("incr", "age")
	if err != nil {
		t.Error(err)
	}
	if reflect.TypeOf(reply).Kind() != reflect.Int64 {
		t.Errorf("bad reply type %s", reflect.TypeOf(reply))
	}
	fmt.Println("->", reply.(int64))

	fmt.Println("incrby age 10")
	reply, err = conn.Do("incrby", "age", 10)
	if err != nil {
		t.Error(err)
	}
	if reflect.TypeOf(reply).Kind() != reflect.Int64 {
		t.Errorf("bad reply type %s", reflect.TypeOf(reply))
	}
	fmt.Println("->", reply.(int64))
}

func TestMSet(t *testing.T) {
	conn := tool.GetRedisPool(host).Get()
	defer conn.Close()

	fmt.Println("mset", "a", "1001", "b", "1002", "c", "1003")
	reply, err := conn.Do("mset", "a", "1001", "b", "1002", "c", "1003")
	if err != nil {
		t.Error(err)
	}
	if reflect.TypeOf(reply).Kind() != reflect.String {
		t.Errorf("bad reply type %s", reflect.TypeOf(reply))
	}
	fmt.Println("->", reply.(string))
}

func TestMGet(t *testing.T) {
	conn := tool.GetRedisPool(host).Get()
	defer conn.Close()

	fmt.Println("mget", "a", "b", "c", "d")
	reply, err := conn.Do("mget", "a", "b", "c", "d")
	if err != nil {
		t.Error(err)
	}
	if reflect.TypeOf(reply).Kind() != reflect.Slice {
		t.Errorf("bad reply type %s", reflect.TypeOf(reply))
	}
	for _, elem := range reply.([]interface{}) {
		switch elem.(type) {
		case []byte:
			fmt.Println("->", string(elem.([]byte)))
		default:
			fmt.Println("->", elem)
		}
	}
}
