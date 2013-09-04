package goredis

import (
	"./storage"
)

type RedisStorages struct {
	StringStorage storage.StringStorage
	HashStorage   storage.HashStorage
	ListStorage   storage.ListStorage
	SetStorage    storage.SetStorage
}

// 主Server
type GoRedisServer struct {
	RedisServer
	Storages *RedisStorages
}

func NewGoRedisServer() (server *GoRedisServer) {
	server = &GoRedisServer{}
	server.init()
	server.Storages = &RedisStorages{}
	server.Storages.StringStorage = storage.NewMemoryStringStorage()
	return
}

func (server *GoRedisServer) OnGET(key string) (val interface{}) {
	val, _ = server.Storages.StringStorage.Get(key)
	return
}

func (server *GoRedisServer) OnSET(key string, val string) (err error) {
	err = server.Storages.StringStorage.Set(key, val)
	return
}

func (server *GoRedisServer) OnMGET(keys ...string) (values []interface{}) {
	values, _ = server.Storages.StringStorage.MGet(keys...)
	return
}

func (server *GoRedisServer) OnDEL(keys ...string) (count int) {
	count, _ = server.Storages.StringStorage.Del(keys...)
	return
}
