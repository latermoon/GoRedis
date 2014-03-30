package goredis_proxy

import (
	"GoRedis/goredis_server"
)

const (
	S_LAST_WRITE_KEY = "last_write_key"
)

func isWriteAction(cmd string) bool {
	return goredis_server.NeedSync(cmd)
}
