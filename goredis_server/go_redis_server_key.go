package goredis_server

import (
	. "../goredis"
	"./libs/leveltool"
	. "./storage"
	"bytes"
	"fmt"
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
	db := server.datasource.DB()
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
	for _, key := range keys {
		entry := server.datasource.Get(key)
		if entry != nil {
			err := server.datasource.Remove(key)
			if err != nil {
				fmt.Println(err)
			}
			n++
		}
	}
	reply = IntegerReply(n)
	return
}

func (server *GoRedisServer) OnTYPE(cmd *Command) (reply *Reply) {
	key, _ := cmd.ArgAtIndex(1)
	entry := server.datasource.Get(key)
	if entry != nil {
		if desc, exist := entryTypeDesc[entry.Type()]; exist {
			return StatusReply(desc)
		}
	}
	return StatusReply("none")
}

// [Custom] 描述一个key
func (server *GoRedisServer) OnDESC(cmd *Command) (reply *Reply) {
	keys := cmd.Args[1:]
	bulks := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		entry := server.datasource.Get(key)
		if entry == nil {
			bulks = append(bulks, string(key)+" [nil]")
			continue
		}
		buf := bytes.Buffer{}
		buf.WriteString(string(key) + " [" + entryTypeDesc[entry.Type()] + "] ")

		switch entry.Type() {
		case EntryTypeString:
			buf.WriteString(entry.(*StringEntry).String())
		case EntryTypeHash:
			buf.WriteString(fmt.Sprintf("%s", entry.(*HashEntry).Map()))
		case EntryTypeSortedSet:
			iter := entry.(*SortedSetEntry).SortedSet().Iterator()
			for iter.Next() {
				members := iter.Value().([]string)
				score := iter.Key().(float64)
				for _, member := range members {
					buf.WriteString(fmt.Sprintf("%s(%s) ", member, server.formatFloat(score)))
				}
			}
		case EntryTypeSet:
			buf.WriteString(fmt.Sprintf("%s", entry.(*SetEntry).Keys()))
		case EntryTypeList:
			for e := entry.(*ListEntry).List().Front(); e != nil; e = e.Next() {
				buf.Write(e.Value.([]byte))
				buf.WriteString(" ")
			}
		default:

		}
		// append
		bulks = append(bulks, buf.String())
	}
	reply = MultiBulksReply(bulks)
	return
}
