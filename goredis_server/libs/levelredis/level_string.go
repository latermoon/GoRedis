package levelredis

import (
// "github.com/latermoon/levigo"
)

type LevelString struct {
	redis *LevelRedis
}

func NewLevelString(redis *LevelRedis) (l *LevelString) {
	l = &LevelString{}
	l.redis = redis
	return
}

func (l *LevelString) stringKey(key []byte) []byte {
	return joinStringBytes(KEY_PREFIX, SEP_LEFT, string(key), SEP_RIGHT, STRING_SUFFIX)
}

func (l *LevelString) Get(key []byte) (value []byte) {
	var err error
	value, err = l.redis.db.Get(l.redis.ro, l.stringKey(key))
	if err != nil {
		value = nil
	}
	return
}

func (l *LevelString) Delete(keys ...[]byte) (n int) {
	n = 0
	for _, key := range keys {
		val := l.Get(key)
		if val != nil {
			l.redis.db.Delete(l.redis.wo, l.stringKey(key))
			n++
		}
	}
	return
}

func (l *LevelString) Set(key []byte, value []byte) (err error) {
	err = l.redis.db.Put(l.redis.wo, l.stringKey(key), value)
	return
}
