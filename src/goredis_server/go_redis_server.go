package goredis_server

import (
	"./storage"
	. "github.com/latermoon/GoRedis/src/goredis"
)

// GoRedisServer
type GoRedisServer struct {
	CommandHandler
	RedisServer
	Storages storage.RedisStorages
}

func NewGoRedisServer() (server *GoRedisServer) {
	server = &GoRedisServer{}
	server.SetHandler(server) // set as itself
	server.Storages = storage.RedisStorages{}
	server.Storages.StringStorage = storage.NewMemoryStringStorage()
	server.Storages.KeyTypeStorage = storage.NewMemoryKeyTypeStorage()
	return
}

// for CommandHandler
func (server *GoRedisServer) On(name string, cmd *Command) (reply *Reply) {
	return ErrorReply("Not Supported: " + cmd.String())
}
