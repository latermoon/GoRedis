package test

import (
	// "fmt"
	"testing"
)

func TestHash(t *testing.T) {
	conn, err := NewRedisConn(host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// clean
	if _, err := conn.Do("DEL", "user"); err != nil {
		t.Fatal(err)
	}

	if reply, err := conn.Do("HSET", "user", "name", "latermoon"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 1 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("HGET", "user", "name"); err != nil {
		t.Fatal(err)
	} else if string(reply.([]byte)) != "latermoon" {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("HMSET", "user", "age", "12", "sex", "male"); err != nil {
		t.Fatal(err)
	} else if reply.(string) != "OK" {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("HMGET", "user", "age", "sex"); err != nil {
		t.Fatal(err)
	} else {
		bulks := reply.([]interface{})
		if string(bulks[0].([]byte)) != "12" ||
			string(bulks[1].([]byte)) != "male" {
			t.Error("bad reply")
		}
	}

	// if reply, err := conn.Do("HINCRBY", "user", "age", "2"); err != nil {
	// 	t.Fatal(err)
	// } else if reply.(int64) != 14 {
	// 	t.Error("bad reply")
	// }

	if reply, err := conn.Do("HLEN", "user"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 3 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("HDEL", "user", "sex"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 1 {
		t.Error("bad reply")
	}

}
