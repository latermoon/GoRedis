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
}
