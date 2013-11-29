package leveltool

/**
基于leveldb实现的zset，用于海量存储，节约内存
__key[user_rank]zset = 2
__zset[user_rank]score:1378000907596:100422 = ""
__zset[user_rank]score:1378000907596:100428 = ""
...
__zset[user_rank]key:100422 = 1378000907596
__zset[user_rank]key:300000 = 1378000907596
...
*/

import (
	"bytes"
)

type LevelZSet struct {
}

func (l *LevelZSet) encodeMemberKey(member []byte) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 100))
	return buf.Bytes()
}
