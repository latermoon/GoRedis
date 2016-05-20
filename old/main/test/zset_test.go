package test

import (
	"fmt"
	"testing"
)

func TestSortedSet(t *testing.T) {
	conn, err := NewRedisConn(host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// clean
	if _, err := conn.Do("DEL", "visitor"); err != nil {
		t.Fatal(err)
	}

	if reply, err := conn.Do("ZADD", "visitor", "1", "A", "2", "B", "3", "C"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 3 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("ZRANK", "visitor", "A"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 0 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("ZREVRANK", "visitor", "A"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 2 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("ZRANGE", "visitor", "0", "-1"); err != nil {
		t.Fatal(err)
	} else {
		bulks := reply.([]interface{})
		if len(bulks) != 3 ||
			string(bulks[0].([]byte)) != "A" ||
			string(bulks[1].([]byte)) != "B" ||
			string(bulks[2].([]byte)) != "C" {
			t.Error("bad reply")
		}
	}

	if reply, err := conn.Do("ZREVRANGE", "visitor", "0", "-1", "WITHSCORES"); err != nil {
		t.Fatal(err)
	} else {
		bulks := reply.([]interface{})
		if len(bulks) != 6 ||
			string(bulks[0].([]byte)) != "C" ||
			string(bulks[1].([]byte)) != "3" ||
			string(bulks[2].([]byte)) != "B" ||
			string(bulks[3].([]byte)) != "2" ||
			string(bulks[4].([]byte)) != "A" ||
			string(bulks[5].([]byte)) != "1" {
			t.Error("bad reply")
		}
	}

	if reply, err := conn.Do("ZRANGEBYSCORE", "visitor", "1", "2"); err != nil {
		t.Fatal(err)
	} else {
		bulks := reply.([]interface{})
		if len(bulks) != 2 ||
			string(bulks[0].([]byte)) != "A" ||
			string(bulks[1].([]byte)) != "B" {
			t.Error("bad reply")
		}
	}

	if reply, err := conn.Do("ZREVRANGEBYSCORE", "visitor", "3", "2", "WITHSCORES"); err != nil {
		t.Fatal(err)
	} else {
		bulks := reply.([]interface{})
		if len(bulks) != 4 ||
			string(bulks[0].([]byte)) != "C" ||
			string(bulks[1].([]byte)) != "3" ||
			string(bulks[2].([]byte)) != "B" ||
			string(bulks[3].([]byte)) != "2" {
			t.Error("bad reply")
		}
	}

	if reply, err := conn.Do("ZREM", "visitor", "A", "C"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 2 {
		t.Error("bad reply")
	}

	// B
	if reply, err := conn.Do("ZCARD", "visitor"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 1 {
		t.Error("bad reply")
	}

	// ABCEFGH - B = 6
	if reply, err := conn.Do("ZADD", "visitor", "4", "E", "5", "F", "1", "A", "2", "B", "3", "C", "6", "G", "7", "H"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 6 {
		fmt.Println(reply)
		t.Error("bad reply")
	}

	if reply, err := conn.Do("ZREMRANGEBYSCORE", "visitor", "3", "4"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 2 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("ZREMRANGEBYRANK", "visitor", "0", "1"); err != nil {
		t.Fatal(err)
	} else if reply.(int64) != 2 {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("ZINCRBY", "visitor", "10", "F"); err != nil {
		t.Fatal(err)
	} else if string(reply.([]byte)) != "15" {
		t.Error("bad reply")
	}

	if reply, err := conn.Do("ZSCORE", "visitor", "F"); err != nil {
		t.Fatal(err)
	} else if string(reply.([]byte)) != "15" {
		t.Error("bad reply")
	}
}
