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
}
