package leveltool

/**
基于leveldb实现的zset，用于海量存储，节约内存
__key[user_rank]zset = 2
__zset[user_rank]s:1378000907596:100422 = ""
__zset[user_rank]s:1378000907596:100428 = ""
...
__zset[user_rank]m:100422 = 1378000907596
__zset[user_rank]m:300000 = 1378000907596
...
*/

type LevelZSet struct {
	key string
}

func (l *LevelZSet) zsetKey() []byte {
	return joinStringBytes(KEY_PREFIX, SEP_LEFT, l.key, SEP_RIGHT, ZSET_SUFFIX)
}

func (l *LevelZSet) memberKey(member []byte) []byte {
	return joinStringBytes(ZSET_PREFIX, SEP_LEFT, l.key, SEP_RIGHT, "m", SEP, string(member))
}

func (l *LevelZSet) scoreKey(member []byte, score []byte) []byte {
	return joinStringBytes(ZSET_PREFIX, SEP_LEFT, l.key, SEP_RIGHT, "s", SEP, string(score), SEP, string(member))
}
