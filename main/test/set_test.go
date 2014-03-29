package test

import (
	"testing"
)

func TestSet(t *testing.T) {
	conn, err := NewRedisConn(host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// clean
	if _, err := conn.Do("DEL", "members"); err != nil {
		t.Fatal(err)
	}

	if reply, err := conn.Do("SADD", "members", "A", "B"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 2 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("SADD", "members", "B", "C", "D"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 3 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("SISMEMBER", "members", "A"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 1 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("SREM", "members", "A", "D"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 2 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("SCARD", "members"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 2 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("SMEMBERS", "members"); err != nil {
		t.Fatal(err)
	} else {
		bulks := reply.([]interface{})
		if len(bulks) != 2 || string(bulks[0].([]byte)) != "B" || string(bulks[1].([]byte)) != "C" {
			t.Error("bad reply")
		}
	}
}
