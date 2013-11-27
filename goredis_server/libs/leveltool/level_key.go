package leveltool

import (
	"bytes"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"strings"
)

type LevelKey struct {
	db *leveldb.DB
	ro *opt.ReadOptions
	wo *opt.WriteOptions
}

func NewLevelKey(db *leveldb.DB) (l *LevelKey) {
	l = &LevelKey{}
	l.db = db
	l.ro = &opt.ReadOptions{}
	l.wo = &opt.WriteOptions{}
	return
}

func (l *LevelKey) TypeOf(key []byte) (t string) {
	iter := l.db.NewIterator(l.ro)
	defer iter.Release()
	// key在leveldb中存储格式如 __key[name]string，__key[user_rank]zset，需要用PrefixEnumerate找出来，然后截取后面类型部分
	prefix := []byte(strings.Join([]string{KEY_PREFIX, SEP_LEFT, string(key), SEP_RIGHT}, ""))
	PrefixEnumerate(iter, prefix, func(i int, iter iterator.Iterator, quit *bool) {
		// 直接从key中截取最后的部分，就是type
		right := bytes.Index(iter.Key(), []byte(SEP_RIGHT))
		t = string(iter.Key()[right+1:])
		*quit = true
	}, "next")
	return
}

// 搜索并返回key和类型
// @param direction "prev" or else for "next"
// @return bulks bulks[0]=key, bulks[1]=type, bulks[2]=key2, ...
func (l *LevelKey) Search(prefix []byte, direction string, count int, withtype bool) (bulks []interface{}) {
	ro := &opt.ReadOptions{}
	iter := l.db.NewIterator(ro)
	defer iter.Release()
	// buffer
	bufsize := count
	if withtype {
		bufsize = bufsize * 2
	}
	// enumerate
	innerPrefix := KEY_PREFIX + SEP_LEFT + string(prefix)
	bulks = make([]interface{}, 0, bufsize)
	PrefixEnumerate(iter, []byte(innerPrefix), func(i int, iter iterator.Iterator, quit *bool) {
		fullkey := copyBytes(iter.Key())
		sepLeftPos := bytes.Index(fullkey, []byte(SEP_LEFT))
		sepRightPos := bytes.Index(fullkey, []byte(SEP_RIGHT))
		key := fullkey[sepLeftPos+1 : sepRightPos]
		bulks = append(bulks, key)
		if withtype {
			t := fullkey[sepRightPos+1:]
			bulks = append(bulks, copyBytes(t))
		}
	}, direction)
	return
}
