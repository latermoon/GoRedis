package test

import (
	"testing"
)

func TestString(t *testing.T) {
	conn, err := NewRedisConn(host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// clean
	if _, err := conn.Do("DEL", "name", "A", "B", "C"); err != nil {
		t.Fatal(err)
	}

	if reply, err := conn.Do("SET", "name", "latermoon"); err != nil {
		t.Fatal(err)
	} else if reply.(string) != "OK" {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("GET", "name"); err != nil {
		t.Fatal(err)
	} else if string(reply.([]byte)) != "latermoon" {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("MSET", "A", "1", "B", "2"); err != nil {
		t.Fatal(err)
	} else if reply.(string) != "OK" {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("MGET", "A", "B"); err != nil {
		t.Fatal(err)
	} else {
		bulks := reply.([]interface{})
		if string(bulks[0].([]byte)) != "1" || string(bulks[1].([]byte)) != "2" {
			t.Error("bad reply")
		}
	}

	if reply, err := conn.Do("INCR", "C"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 1 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("INCRBY", "C", "2"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 3 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("DECR", "C"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 2 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("DECRBY", "C", 4); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != -2 {
		t.Error("bad reply")
	}
}
