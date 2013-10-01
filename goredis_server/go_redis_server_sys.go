package goredis_server

import (
	. "../goredis"
	"fmt"
)

func (server *GoRedisServer) OnPING(cmd *Command) (reply *Reply) {
	reply = StatusReply("PONG")
	return
}

func (server *GoRedisServer) OnINFO(cmd *Command) (reply *Reply) {
	reply = BulkReply("GoRedis by latemroon")
	return
}

func (server *GoRedisServer) OnCOUNTER(cmd *Command) (reply *Reply) {
	bulks := make([]interface{}, 0, len(server.counters))

	for name, counter := range server.counters {
		line := fmt.Sprintf("%s, %d", name, counter.Count())
		bulks = append(bulks, line)
	}
	reply = MultiBulksReply(bulks)
	return
}
