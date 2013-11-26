package goredis_server

import (
	. "../goredis"
	"./libs/leveltool"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// 在数据量大的情况下，keys基本没有意义
// 取消keys，使用key_next或者key_prev来分段扫描全部key
func (server *GoRedisServer) OnKEYS(cmd *Command) (reply *Reply) {
	return ErrorReply("keys is not supported, use key_next/key_prev instead")
}

// 找出下一个key
// @return ["user:100422:name", "string", "user:100428:name", "string", "user:100422:setting", "hash", ...]
func (server *GoRedisServer) OnKEY_NEXT(cmd *Command) (reply *Reply) {
	seekkey, err := cmd.ArgAtIndex(1)
	if err != nil {
		return ErrorReply(err)
	}
	count := 1
	if len(cmd.Args) > 2 {
		count, err = cmd.IntAtIndex(2)
		if err != nil {
			return ErrorReply(err)
		}
		if count < 1 || count > 10000 {
			return ErrorReply("count range: 1 < count < 10000")
		}
	}
	withtype := false
	if len(cmd.Args) > 3 {
		withtype = cmd.StringAtIndex(3) == "withtype"
	}
	// search
	bulks := server.keySearch(seekkey, "next", count, withtype)
	return MultiBulksReply(bulks)
}

func (server *GoRedisServer) OnKEY_PREV(cmd *Command) (reply *Reply) {
	seekkey, err := cmd.ArgAtIndex(1)
	if err != nil {
		return ErrorReply(err)
	}
	count := 1
	if len(cmd.Args) > 2 {
		count, err = cmd.IntAtIndex(2)
		if err != nil {
			return ErrorReply(err)
		}
		if count < 1 || count > 10000 {
			return ErrorReply("count range: 1 < count < 10000")
		}
	}
	withtype := false
	if len(cmd.Args) > 3 {
		withtype = cmd.StringAtIndex(3) == "withtype"
	}
	// search
	bulks := server.keySearch(seekkey, "prev", count, withtype)
	return MultiBulksReply(bulks)
}

// 搜索并返回key和类型
// @param direction "prev" or else for "next"
// @return bulks bulks[0]=key, bulks[1]=type, bulks[2]=key2, ...
func (server *GoRedisServer) keySearch(prefix []byte, direction string, count int, withtype bool) (bulks []interface{}) {
	db := server.DB()
	ro := &opt.ReadOptions{}
	iter := db.NewIterator(ro)
	defer iter.Release()
	// buffer
	bufsize := count
	if withtype {
		bufsize = bufsize * 2
	}
	// enumerate
	bulks = make([]interface{}, 0, bufsize)
	leveltool.PrefixEnumerate(iter, prefix, func(i int, iter iterator.Iterator, quit *bool) {
		bulks = append(bulks, copyBytes(iter.Key()))
		if withtype {
			bs := iter.Value()[0] // 第一个字节
			bulks = append(bulks, EntryTypeDescription(EntryType(bs)))
		}
	}, direction)
	return
}

func (server *GoRedisServer) OnDEL(cmd *Command) (reply *Reply) {
	keys := cmd.Args[1:]
	n := 0
	// TODO
	for _, key := range keys {
		n++
		t := server.keyManager.levelKey().TypeOf(key)
		switch t {
		case "string":
			server.keyManager.levelString().Delete(key)
		case "hash":
			server.keyManager.hashByKey(string(key)).Drop()
		case "set":
			server.keyManager.setByKey(string(key)).Drop()
		case "list":
			server.keyManager.listByKey(string(key)).Drop()
		case "zset":
			server.keyManager.zsetByKey(string(key)).Drop()
		default:
			n--
		}
	}
	reply = IntegerReply(n)
	return
}

func (server *GoRedisServer) OnTYPE(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	t := server.keyManager.levelKey().TypeOf(key)
	if len(t) > 0 {
		reply = StatusReply(t)
	} else {
		reply = StatusReply("none")
	}
	return
}
