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
	prefix := []byte(strings.Join([]string{KEY_PREFIX, SEP_LEFT, string(key), SEP_RIGHT}, ""))
	PrefixEnumerate(iter, prefix, func(i int, iter iterator.Iterator, quit *bool) {
		// 直接从key中截取最后的部分，就是type
		right := bytes.Index(iter.Key(), []byte(SEP_RIGHT))
		t = string(iter.Key()[right+1:])
		*quit = true
	}, "next")
	return
}
