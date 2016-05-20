package test

import (
	"testing"
)

func TestList(t *testing.T) {
	conn, err := NewRedisConn(host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// clean
	if _, err := conn.Do("DEL", "queue"); err != nil {
		t.Fatal(err)
	}

	if reply, err := conn.Do("RPUSH", "queue", "B", "C", "D", "E"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 4 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("LINDEX", "queue", "0"); err != nil {
		t.Fatal(err)
	} else if string(reply.([]byte)) != "B" {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("LPUSH", "queue", "A"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 5 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("LINDEX", "queue", "0"); err != nil {
		t.Fatal(err)
	} else if string(reply.([]byte)) != "A" {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("RPOP", "queue"); err != nil {
		t.Fatal(err)
	} else if string(reply.([]byte)) != "E" {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("LPOP", "queue"); err != nil {
		t.Fatal(err)
	} else if string(reply.([]byte)) != "A" {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("LLEN", "queue"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 3 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("LRANGE", "queue", "1", "-1"); err != nil {
		t.Fatal(err)
	} else {
		bulks := reply.([]interface{})
		if len(bulks) != 2 || string(bulks[0].([]byte)) != "C" || string(bulks[1].([]byte)) != "D" {
			t.Error("bad reply")
		}
	}

	if reply, err := conn.Do("LTRIM", "queue", "0", "1"); err != nil {
		t.Fatal(err)
	} else if reply.(string) != "OK" {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("LLEN", "queue"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 2 {
		t.Error("bad reply")
	}
}
