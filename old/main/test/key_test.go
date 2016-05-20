package test

import (
	"testing"
)

func TestKey(t *testing.T) {
	conn, err := NewRedisConn(host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

}
