package goredis_server

/*
自定义geo指令集，用于实现附近位置搜索
geo_insert map lat lng value time <StatusReply: OK>
geo_nearby map lat lng start end
geo_del map value

*/

import (
	. "../goredis"
)

func (server *GoRedisServer) OnGEO_INSERT(cmd *Command) (reply *Reply) {
	return StatusReply("OK")
}

func (server *GoRedisServer) OnGEO_NEARBY(cmd *Command) (reply *Reply) {
	return StatusReply("OK")
}

func (server *GoRedisServer) OnGEO_DEL(cmd *Command) (reply *Reply) {
	return StatusReply("OK")
}
