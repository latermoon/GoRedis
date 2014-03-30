package goredis_proxy

import (
	"GoRedis/goredis_server"
	"strings"
)

const (
	S_LAST_WRITE_KEY = "last_write_key"
)

// 跳过指令
var ignoreCmdList = "DUMP,KEYS,MIGRATE,MOVE,OBJECT,RESTORE,SCAN,EVAL,EVALSHA,SCRIPT,DISCARD,EXEC,MULTI,UNWATCH,WATCH,PSUBSCRIBE,PUBSUB,PUBLISH,PUNSUBSCRIBE,SUBSCRIBE,UNSUBSCRIBE,BGREWRITEAOF,BGSAVE,CLIENT,DBSIZE,DEBUG,FLUSHALL,FLUSHDB,LASTSAVE,SAVE,SHUTDOWM,SLAVEOF,SLOWLOG,SYNC,TIME"

var ignoreSync = map[string]bool{}

func init() {
	for _, cmd := range strings.Split(ignoreCmdList, ",") {
		ignoreSync[cmd] = true
	}
}

func isWriteAction(cmd string) bool {
	return goredis_server.NeedSync(cmd)
}
