package goredis_server

import (
	. "../goredis"
	. "./storage"
	"fmt"
	"strconv"
)

// 命令类型集合
var cmdCategory = map[string][]string{
	"string": []string{"GET", "SET", "INCR", "DECR", "INCRBY", "DECRBY", "MSET", "MGET"},
	"hash":   []string{"HDEL", "HGET", "HSET", "HMGET", "HMSET", "HGETALL", "HINCRBY", "HKEYS", "HLEN"},
	"list":   []string{"LINDEX", "LLEN", "LPOP", "LPUSH", "LRANGE", "LREM", "RPOP", "RPUSH"},
	"set":    []string{"SADD", "SCARD", "SISMEMBER", "SMEMBERS", "SREM"},
	"zset":   []string{"ZADD", "ZCARD", "ZINCRBY", "ZRANGE", "ZRANGEBYSCORE", "ZREM", "ZREMRANGEBYRANK", "ZREMRANGEBYSCORE", "ZREVRANGE", "ZREVRANGEBYSCORE", "ZSCORE"},
}

// 需要同步到从库的命令
var needSyncCmds = []string{
	"SET", "INCR", "DECR", "INCRBY", "DECRBY", "MSET",
	"HDEL", "HSET", "HMSET", "HINCRBY",
	"LPOP", "LPUSH", "LREM", "RPOP", "RPUSH",
	"SADD", "SREM",
	"ZADD", "ZINCRBY", "ZREM",
	"DEL"}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'g', 12, 64)
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
			switch value.(type) {
			case string:
				args = append(args, []byte(field))
				args = append(args, []byte(value.(string)))
			case []byte:
				args = append(args, []byte(field))
				args = append(args, value.([]byte))
			default:
				fmt.Println("bad hset type", field, value)
			}
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
