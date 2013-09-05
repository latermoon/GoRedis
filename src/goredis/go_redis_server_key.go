package goredis

import (
	"./storage"
)

func (server *GoRedisServer) OnDEL(cmd *Command, keys ...string) (count int, err error) {
	count = 0
	for _, key := range keys {
		switch server.Storages.KeyTypeStorage.GetType(key) {
		case storage.KeyTypeString:
			n, _ := server.Storages.StringStorage.Del([]string{key}...)
			count += n
		default:
		}
	}
	return
}
