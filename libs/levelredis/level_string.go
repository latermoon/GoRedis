package levelredis

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
	value = l.redis.RawGet(l.stringKey(key))
	return
}

func (l *LevelString) Delete(keys ...[]byte) (n int) {
	n = 0
	for _, key := range keys {
		val := l.Get(key)
		if val != nil {
			l.redis.RawDel(l.stringKey(key))
			n++
		}
	}
	return
}

func (l *LevelString) Set(key []byte, value []byte) error {
	return l.redis.RawSet(l.stringKey(key), value)
}
