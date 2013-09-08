package goredis_server

import (
	"./storage"
	//. "github.com/latermoon/GoRedis/src/goredis"
	. "../goredis"
)

// GoRedisServer
type GoRedisServer struct {
	CommandHandler
	RedisServer
	// 存储支持
	Storages storage.RedisStorages
}

func NewGoRedisServer() (server *GoRedisServer) {
	server = &GoRedisServer{}
	// set as itself
	server.SetHandler(server)
	// default storages
	server.Storages = storage.RedisStorages{}
	server.Storages.KeyTypeStorage = storage.NewMemoryKeyTypeStorage()
	server.Storages.StringStorage, _ = storage.NewLevelDBStringStorage("/tmp/goredis.ldb")
	//server.Storages.StringStorage = storage.NewMemoryStringStorage()
	server.Storages.ListStorage = storage.NewMemoryListStorage()
	server.Storages.HashStorage = storage.NewMemoryHashStorage()
	return
}

// for CommandHandler
func (server *GoRedisServer) On(name string, cmd *Command) (reply *Reply) {
	return ErrorReply("Not Supported: " + cmd.String())
}
