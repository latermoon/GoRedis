package goredis_server

/**
自定义enum指令集
enum_next key count 找出下一个key的内容
enum_prev key count 找出上一个key的内容
*/
import (
	. "../goredis"
	// . "./storage"
	"bytes"
	// "fmt"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func (server *GoRedisServer) OnENUM_NEXT(cmd *Command) (reply *Reply) {
	seekkey, err1 := cmd.ArgAtIndex(1)
	count, err2 := cmd.IntAtIndex(2)
	if err1 != nil {
		return ErrorReply(err1)
	} else if err2 != nil {
		return ErrorReply(err2)
	}
	if count < 1 || count > 10000 {
		return ErrorReply("count range: 1 < count < 10000")
	}
	db := server.datasource.DB()
	ro := &opt.ReadOptions{}
	// seek
	iter := db.NewIterator(ro)
	defer iter.Release()
	iter.Seek([]byte(seekkey))
	// result
	bulks := make([]interface{}, 0, count)
	if bytes.Compare(iter.Key(), seekkey) != 0 {
		bulks = append(bulks, copyBytes(iter.Key()))
	}
	for len(bulks) < count {
		if !iter.Next() {
			break
		}
		bulks = append(bulks, copyBytes(iter.Key()))
	}

	return MultiBulksReply(bulks)
}

func copyBytes(src []byte) (dst []byte) {
	dst = make([]byte, len(src))
	copy(dst, src)
	return
}
