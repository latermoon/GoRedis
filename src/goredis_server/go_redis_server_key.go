package goredis_server

import (
	. "../goredis"
	"./storage"
)

func (server *GoRedisServer) OnDEL(cmd *Command) (reply *Reply) {
	keys := cmd.StringArgs()[1:]
	count := 0
	for _, key := range keys {
		switch server.Storages.KeyTypeStorage.GetType(key) {
		case storage.KeyTypeString:
			n, _ := server.Storages.StringStorage.Del([]string{key}...)
			count += n
		default:
		}
	}
	reply = IntegerReply(count)
	return
}
