package goredis

import ()

func (server *GoRedisServer) OnGET(cmd *Command, key string) (val interface{}, err error) {
	val, err = server.Storages.StringStorage.Get(key)
	return
}

func (server *GoRedisServer) OnSET(cmd *Command, key string, val string) (err error) {
	err = server.Storages.StringStorage.Set(key, val)
	return
}

func (server *GoRedisServer) OnMGET(cmd *Command, keys ...string) (values []interface{}, err error) {
	values, err = server.Storages.StringStorage.MGet(keys...)
	return
}

func (server *GoRedisServer) OnMSET(cmd *Command, keyvals ...string) (err error) {
	err = server.Storages.StringStorage.MSet(keyvals...)
	return
}

func (server *GoRedisServer) OnDEL(cmd *Command, keys ...string) (count int, err error) {
	count, err = server.Storages.StringStorage.Del(keys...)
	return
}
