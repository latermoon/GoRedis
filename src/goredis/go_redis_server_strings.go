package goredis

import (
	"./storage"
)

func (server *GoRedisServer) OnGET(cmd *Command, key string) (val interface{}, err error) {
	val, err = server.Storages.StringStorage.Get(key)
	return
}

func (server *GoRedisServer) OnSET(cmd *Command, key string, val string) (err error) {
	server.Storages.KeyTypeStorage.SetType(key, storage.KeyTypeString)
	err = server.Storages.StringStorage.Set(key, val)
	return
}

func (server *GoRedisServer) OnMGET(cmd *Command, keys ...string) (values []interface{}, err error) {
	values, err = server.Storages.StringStorage.MGet(keys...)
	return
}

func (server *GoRedisServer) OnMSET(cmd *Command, keyvals ...string) (err error) {
	for i := 0; i < len(keyvals); i += 2 {
		server.Storages.KeyTypeStorage.SetType(keyvals[i], storage.KeyTypeString)
	}
	err = server.Storages.StringStorage.MSet(keyvals...)
	return
}
