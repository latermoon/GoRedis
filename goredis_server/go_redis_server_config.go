package goredis_server

import (
	. "GoRedis/libs/goredis"
	"strings"
)

/*

# > 100ms (microseconds)
slowlog-log-slower-than 100000


*/

func (server *GoRedisServer) OnCONFIG(cmd *Command) (reply *Reply) {
	action := cmd.StringAtIndex(1)
	switch strings.ToUpper(action) {
	case "GET":
		reply = server.configGet(cmd)
	case "SET":
		key := cmd.StringAtIndex(2)
		value := cmd.StringAtIndex(3)
		server.config.Set(key, []byte(value))
		reply = StatusReply("OK")
	default:
		reply = ErrorReply("bad config action: " + action)
	}
	return
}

func (server *GoRedisServer) configGet(cmd *Command) (reply *Reply) {
	patten := cmd.StringAtIndex(2)
	if patten == "*" {
		bulks := make([]interface{}, 0, 10)
		keys := server.config.Keys()
		for _, k := range keys {
			bulks = append(bulks, k)
			v := server.config.StringForKey(k)
			bulks = append(bulks, v)
		}
		reply = MultiBulksReply(bulks)
	} else {
		value := server.config.Get(patten)
		if value == nil {
			reply = BulkReply(nil)
		} else {
			reply = BulkReply(value)
		}
	}
	return
}
