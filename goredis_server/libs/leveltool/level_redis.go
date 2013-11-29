package leveltool

/*
基于leveldb实现的redis持久化层
*/

import (
	"bytes"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type IteratorDirection int

const (
	IteratorForward IteratorDirection = iota
	IteratorBackward
)

type LevelRedis struct {
	db *leveldb.DB
	ro *opt.ReadOptions
	wo *opt.WriteOptions
}

func NewLevelRedis(db *leveldb.DB) (l *LevelRedis) {
	l = &LevelRedis{}
	l.db = db
	l.ro = &opt.ReadOptions{}
	l.wo = &opt.WriteOptions{}
	return
}

func (l *LevelRedis) Strings() (s *LevelString) {
	s = NewLevelString(l.db)
	return
}

func (l *LevelRedis) GetList(key string) (lst *LevelList) {
	lst = NewLevelList(l.db, key)
	return
}

func (l *LevelRedis) GetHash(key string) (h *LevelHash) {
	h = NewLevelHash(l.db, key, false)
	return
}

func (l *LevelRedis) GetSet(key string) (s *LevelHash) {
	s = NewLevelHash(l.db, key, true)
	return
}

func (l *LevelRedis) GetSortedSet(key string) (z *LevelZSet) {
	z = NewLevelZSet(l, key)
	return
}

func (l *LevelRedis) Enumerate(min, max []byte, direction IteratorDirection, fn func(i int, key, value []byte, quit *bool)) {
	iter := l.db.NewIterator(l.ro)
	defer iter.Release()
	found := false
	if direction == IteratorBackward {
		found = iter.Seek(max)
	} else {
		found = iter.Seek(min)
	}
	i := -1
	if found {
		i++
		quit := false
		fn(i, iter.Key(), iter.Value(), &quit)
		if quit {
			return
		}
	}
	for {
		found = false
		if direction == IteratorBackward {
			found = iter.Prev()
			if !found || bytes.Compare(iter.Key(), min) < 0 {
				break
			}
		} else {
			found = iter.Next()
			if !found || bytes.Compare(iter.Key(), max) > 0 {
				break
			}
		}
		i++
		quit := false
		fn(i, iter.Key(), iter.Value(), &quit)
		if quit {
			return
		}
	}

	return
}
