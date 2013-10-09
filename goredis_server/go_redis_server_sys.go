package goredis_server

import (
	. "../goredis"
	"bytes"
	"fmt"
)

func (server *GoRedisServer) OnPING(cmd *Command) (reply *Reply) {
	reply = StatusReply("PONG")
	return
}

func (server *GoRedisServer) OnINFO(cmd *Command) (reply *Reply) {
	// section := cmd.StringAtIndex(1)
	buf := bytes.Buffer{}
	buf.WriteString("# Server\n")
	buf.WriteString("goredis_version:0.1\n")
	buf.WriteString("\n")
	buf.WriteString("# Clients\n")
	buf.WriteString("connected_clients:n\n")
	reply = BulkReply(buf.String())
	return
}

func (server *GoRedisServer) OnCOUNTER(cmd *Command) (reply *Reply) {
	counters := server.cmdCounters.Counters()
	bulks := make([]interface{}, 0, len(counters))

	for name, counter := range counters {
		line := fmt.Sprintf("%s, %d", name, counter.Count())
		bulks = append(bulks, line)
	}
	reply = MultiBulksReply(bulks)
	return
}
