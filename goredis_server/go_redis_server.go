package goredis_server

import (
	"./storage"
	"container/list"
	. "github.com/latermoon/GoRedis/goredis"
)

// GoRedisServer
type GoRedisServer struct {
	CommandHandler
	RedisServer
	// 存储支持
	Storages storage.RedisStorages
	// 从库
	Slaves *list.List
}

func NewGoRedisServer() (server *GoRedisServer) {
	server = &GoRedisServer{}
	// set as itself
	server.SetHandler(server)
	// default storages
	server.Storages = storage.RedisStorages{}
	leveldbStorage, _ := storage.NewLevelDBStorage("/tmp/goredis.ldb")
	server.Storages.KeyTypeStorage = storage.NewMemoryKeyTypeStorage()
	server.Storages.StringStorage = leveldbStorage
	server.Storages.ListStorage = storage.NewMemoryListStorage()
	server.Storages.HashStorage = storage.NewMemoryHashStorage()
	// slave
	server.Slaves = list.New()

	return
}

// for CommandHandler
func (server *GoRedisServer) On(name string, cmd *Command) (reply *Reply) {
	return ErrorReply("Not Supported: " + cmd.String())
}
