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
	reply = BulkReply("GoRedis by latemroon\n")
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
