package test

import (
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
}
