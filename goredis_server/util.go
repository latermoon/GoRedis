package goredis_server

import (
	. "../goredis"
	. "./storage"
	"strconv"
	"strings"
)

// 数据类型描述
var entryTypeDesc = map[EntryType]string{
	EntryTypeUnknown:   "unknown",
	EntryTypeString:    "string",
	EntryTypeHash:      "hash",
	EntryTypeList:      "list",
	EntryTypeSet:       "set",
	EntryTypeSortedSet: "zset"}

// 指令集名称 CommandCategory
type CCate string

const (
	CCateKey         CCate = "key"
	CCateString      CCate = "string"
	CCateHash        CCate = "hash"
	CCateList        CCate = "list"
	CCateSet         CCate = "set"
	CCateSortedSet   CCate = "zset"
	CCateAof         CCate = "aof"
	CCatePubSub      CCate = "pubsub"
	CCateTransaction CCate = "trans"
	CCateScript      CCate = "script"
	CCateConnection  CCate = "conn"
	CCateServer      CCate = "server"
	CCateUnknown     CCate = "unknown"
)

var CommandCategoryList = []CCate{CCateKey, CCateString, CCateHash, CCateList, CCateSet, CCateSortedSet, CCateAof, CCatePubSub, CCateTransaction, CCateScript, CCateConnection, CCateServer, CCateUnknown}

// 指令所属类别
var ccatemap map[string]CCate

// 指令集命令列表
var ccatemaplist = map[CCate][]string{
	CCateKey:         []string{"DEL", "DUMP", "EXISTS", "EXPIRE", "EXPIREAT", "KEYS", "MIGRATE", "MOVE", "OBJECT", "PERSIST", "PEXPIRE", "PEXPIREAT", "PTTL", "RANDOMKEY", "RENAME", "RENAMENX", "RESTORE", "SORT", "TTL", "TYPE"},
	CCateString:      []string{"APPEND", "BITCOUNT", "BITOP", "DECR", "DECRBY", "GET", "GETBIT", "GETRANGE", "GETSET", "INCR", "INCRBY", "INCRBYFLOAT", "MGET", "MSET", "MSETNX", "PSETEX", "SET", "SETBIT", "SETEX", "SETNX", "SETRANGE", "STRLEN"},
	CCateHash:        []string{"HDEL", "HEXISTS", "HGET", "HGETALL", "HINCRBY", "HINCRBYFLOAT", "HKEYS", "HLEN", "HMGET", "HMSET", "HSET", "HSETNX", "HVALS"},
	CCateList:        []string{"BLPOP", "BRPOP", "BRPOPLPUSH", "LINDEX", "LINSERT", "LLEN", "LPOP", "LPUSH", "LPUSHX", "LRANGE", "LREM", "LSET", "LTRIM", "RPOP", "RPOPLRUSH", "RPUSH", "RPUSHX"},
	CCateSet:         []string{"SADD", "SCARD", "SDIFF", "SDIFFSTORE", "SINTER", "SINTERSTORE", "SISMEMBER", "SMEMBERS", "SMOVE", "SPOP", "SRANDMEMBER", "SREM", "SUNION", "SUNIONSTORE"},
	CCateSortedSet:   []string{"ZADD", "ZCARD", "ZCOUNT", "ZINCRBY", "ZINTERSTORE", "ZRANGE", "ZRANGEBYSCORE", "ZRANK", "ZREM", "ZREMRANGEBYRANK", "ZREMRANGEBYSCORE", "ZREVRANGE", "ZREVRANGEBYSCORE", "ZREVRANK", "ZSCORE", "ZUNIONSTORE"},
	CCateAof:         []string{"AOF_PUSH", "AOF_PUSH_ASYNC", "AOF_POP", "AOF_INDEX", "AOF_RANGE", "AOF_LEN"},
	CCatePubSub:      []string{"PSUBSCRIBE", "PUBSUB", "PUBLISH", "PUNSUBSCRIBE", "SUBSCRIBE", "UNSUBSCRIBE"},
	CCateTransaction: []string{"DISCARD", "EXEC", "MULTI", "UNWATCH", "WATCH"},
	CCateScript:      []string{"EVAL", "EVALSHA", "SCRIPT"},
	CCateConnection:  []string{"AUTH", "ECHO", "PING", "QUIT", "SELECT"},
	CCateServer:      []string{"BGREWRITEAOF", "BGSAVE", "CLIENT", "CONFIG", "DBSIZE", "DEBUG", "FLUSHALL", "FLUSHDB", "INFO", "LASTSAVE", "MONITOR", "SAVE", "SHUTDOWM", "SLAVEOF", "SLOWLOG", "SYNC", "TIME"},
}

// 需要同步到从库的命令
var needSyncCmds = []string{
	"SET", "INCR", "DECR", "INCRBY", "DECRBY", "MSET",
	"HDEL", "HSET", "HMSET", "HINCRBY",
	"LPOP", "LPUSH", "LREM", "RPOP", "RPUSH",
	"SADD", "SREM",
	"ZADD", "ZINCRBY", "ZREM",
	"DEL"}

func init() {
	// 填充
	ccatemap = make(map[string]CCate)
	for cate, cmds := range ccatemaplist {
		for _, cmd := range cmds {
			// ccatemap["GET"] = CCateString
			// ccatemap["SET"] = CCateString
			// ccatemap["LPUSH"] = CCateList
			ccatemap[cmd] = cate
		}
	}
}

// 返回一个指令所属的分类
func GetCommandCategory(cmd string) (cate CCate) {
	var exist bool
	cate, exist = ccatemap[strings.ToUpper(cmd)]
	if !exist {
		cate = CCateUnknown
	}
	return
}

func EntryTypeDescription(et EntryType) (s string) {
	var exist bool
	s, exist = entryTypeDesc[et]
	if !exist {
		s = entryTypeDesc[EntryTypeUnknown]
	}
	return
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'g', 12, 64)
}

func copyBytes(src []byte) (dst []byte) {
	dst = make([]byte, len(src))
	copy(dst, src)
	return
}

func BytesToInterfaceSlice(vals [][]byte) (result []interface{}) {
	result = make([]interface{}, len(vals))
	for i, val := range vals {
		result[i] = val
	}
	return
}

func StringToInterfaceSlice(vals []string) (result []interface{}) {
	result = make([]interface{}, len(vals))
	for i, val := range vals {
		result[i] = val
	}
	return
}

func entryToCommand(key []byte, entry Entry) (cmd *Command) {
	args := make([][]byte, 0, 10)

	switch entry.Type() {
	case EntryTypeString:
		args = append(args, []byte("SET"))
		args = append(args, key)
		args = append(args, entry.(*StringEntry).Value())
	case EntryTypeHash:
		table := entry.(*HashEntry).Map()
		args = append(args, []byte("HMSET"))
		args = append(args, key)
		for field, value := range table {
			args = append(args, []byte(field))
			args = append(args, value)
		}
	case EntryTypeSortedSet:
		args = append(args, []byte("ZADD"))
		args = append(args, key)
		iter := entry.(*SortedSetEntry).SortedSet().Iterator()
		for iter.Next() {
			score := iter.Key().(float64)
			arr := iter.Value().([]string)
			for _, member := range arr {
				args = append(args, []byte(formatFloat(score)))
				args = append(args, []byte(member))
			}
		}
	case EntryTypeSet:
		args = append(args, []byte("SADD"))
		args = append(args, key)
		keys := entry.(*SetEntry).Keys()
		for _, key := range keys {
			args = append(args, []byte(key.(string)))
		}
	case EntryTypeList:
		args = append(args, []byte("RPUSH"))
		args = append(args, key)
		sl := entry.(*ListEntry).List()
		for e := sl.Front(); e != nil; e = e.Next() {
			args = append(args, e.Value.([]byte))
		}
	default:
	}
	if len(args) > 0 {
		cmd = NewCommand(args...)
	}
	return
}
