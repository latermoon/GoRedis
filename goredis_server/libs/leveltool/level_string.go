package leveltool

/*
__key:name:string = latermoon
__key:age:string = 27
*/

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"strings"
)

type LevelString struct {
	db *leveldb.DB
	ro *opt.ReadOptions
	wo *opt.WriteOptions
}

func NewLevelString(db *leveldb.DB) (l *LevelString) {
	l = &LevelString{}
	l.db = db
	l.ro = &opt.ReadOptions{}
	l.wo = &opt.WriteOptions{}
	return
}

func (l *LevelString) stringKey(key []byte) []byte {
	return []byte(strings.Join([]string{KEY_PREFIX, SEP_LEFT, string(key), SEP_RIGHT, STRING_SUFFIX}, ""))
}

func (l *LevelString) Get(key []byte) (value []byte) {
	var err error
	value, err = l.db.Get(l.stringKey(key), l.ro)
	if err != nil {
		value = nil
	}
	return
}

func (l *LevelString) Delete(key []byte) (n int) {
	sk := l.stringKey(key)
	val := l.Get(sk)
	if val != nil {
		l.db.Delete(sk, l.wo)
		n = 1
	}
	return
}

func (l *LevelString) Set(key []byte, value []byte) (err error) {
	err = l.db.Put(l.stringKey(key), value, l.wo)
	return
}
