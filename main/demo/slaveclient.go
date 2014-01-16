package main

import (
	. "../../goredis"
	. "../../goredis_server"
	"../../libs/stdlog"
	"bufio"
	"net"
)

func main() {
	// 10.80.101.193:7200
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		panic(err)
	}
	session := NewSession(conn)
	client := NewSlaveClient(session)
	client.SetCallback(&Callback{})
	err = client.Sync("")
	if err != nil {
		panic(err)
	}
}

type Callback struct {
	SlaveClientCallback
}

func (s *Callback) RdbSizeCallback(size int64) {
	stdlog.Println("recv rdb size:", size)
}

func (s *Callback) RdbRecvCallback(r *bufio.Reader) {
}

func (s *Callback) IdleCallback() {
	stdlog.Println("waiting ...")
}

func (s *Callback) CommandRecvCallback(cmd *Command) {
	stdlog.Println("recv:", cmd)
}
