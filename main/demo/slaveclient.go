package main

import (
	. "../../goredis"
	. "../../goredis_server"
	"../../libs/stdlog"
	"bufio"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		panic(err)
	}
	session := NewSession(conn)
	client := NewSlaveClient(session)
	client.RdbSizeCallback = func(size int64) {
		stdlog.Println("recv rdb size:", size)
	}
	client.RdbRecvCallback = func(r *bufio.Reader) {

	}
	client.CommandRecvCallback = func(cmd *Command) {
		stdlog.Println("recv:", cmd)
	}
	err = client.Sync("")
	if err != nil {
		panic(err)
	}
}
