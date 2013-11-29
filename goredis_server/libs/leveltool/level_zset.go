package leveltool

/**
基于leveldb实现的zset，用于海量存储，节约内存
+[user_rank]zset = 2
_z[user_rank]s#1378000907596#100422 = ""
_z[user_rank]s#1378000907596#100428 = ""
...
_z[user_rank]m#100422 = 1378000907596
_z[user_rank]m#300000 = 1378000907596
...
*/

import (
	"bytes"
	// "fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"strconv"
	"sync"
)

type LevelZSet struct {
	redis      *LevelRedis
	ro         *opt.ReadOptions
	wo         *opt.WriteOptions
	key        string
	totalCount int
	mu         sync.Mutex
}

func NewLevelZSet(redis *LevelRedis, key string) (l *LevelZSet) {
	l = &LevelZSet{}
	l.redis = redis
	l.key = key
	l.totalCount = -1
	return
}

func (l *LevelZSet) Size() int {
	return 0
}

func (l *LevelZSet) initOnce() {
	if l.totalCount == -1 {
		data, _ := l.redis.db.Get(l.zsetKey(), l.redis.ro)
		if data != nil {
			l.totalCount, _ = strconv.Atoi(string(data))
		} else {
			l.totalCount = 0
		}
	}
}

func (l *LevelZSet) zsetKey() []byte {
	return joinStringBytes(KEY_PREFIX, SEP_LEFT, l.key, SEP_RIGHT, ZSET_SUFFIX)
}

func (l *LevelZSet) zsetValue() []byte {
	s := strconv.Itoa(l.totalCount)
	return []byte(s)
}

func (l *LevelZSet) memberKey(member []byte) []byte {
	return joinStringBytes(ZSET_PREFIX, SEP_LEFT, l.key, SEP_RIGHT, "m", SEP, string(member))
}

func (l *LevelZSet) scoreKey(member []byte, score []byte) []byte {
	return joinStringBytes(ZSET_PREFIX, SEP_LEFT, l.key, SEP_RIGHT, "s", SEP, string(score), SEP, string(member))
}

func (l *LevelZSet) scoreKeyPrefix() []byte {
	return joinStringBytes(ZSET_PREFIX, SEP_LEFT, l.key, SEP_RIGHT, "s", SEP)
}

// __zset[user_rank]s#1378000907596#100428 = ""
func (l *LevelZSet) splitScoreKey(scorekey []byte) (score, member []byte) {
	pos2 := bytes.LastIndex(scorekey, []byte(SEP))
	pos1 := bytes.LastIndex(scorekey[:pos2], []byte(SEP))
	member = copyBytes(scorekey[pos2+1:])
	score = copyBytes(scorekey[pos1+1 : pos2])
	return
}

func (l *LevelZSet) Add(scoreMembers ...[]byte) (n int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.initOnce()
	batch := new(leveldb.Batch)
	count := len(scoreMembers)
	for i := 0; i < count; i += 2 {
		score := scoreMembers[i]
		member, memberkey := scoreMembers[i+1], l.memberKey(scoreMembers[i+1])
		// set member
		batch.Put(memberkey, score)
		// remove old score
		oldscore, e1 := l.redis.db.Get(memberkey, l.redis.ro)
		if e1 == nil && oldscore != nil {
			batch.Delete(l.scoreKey(member, oldscore))
		} else {
			l.totalCount++
		}
		// new score
		batch.Put(l.scoreKey(member, score), nil)
		n++
	}
	batch.Put(l.zsetKey(), l.zsetValue())
	l.redis.db.Write(batch, l.redis.wo)
	return
}

func (l *LevelZSet) Score(member []byte) (score []byte) {
	var err error
	score, err = l.redis.db.Get(l.memberKey(member), l.redis.ro)
	if err != nil || score == nil {
		return
	}
	return
}

func (l *LevelZSet) IncrBy(member []byte, incr int64) (newscore []byte) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.initOnce()
	score := l.Score(member)
	batch := new(leveldb.Batch)
	if score == nil {
		newscore = Int64ToBytes(incr)
	} else {
		batch.Delete(l.scoreKey(member, score))
		scoreInt := BytesToInt64(score)
		newscore = Int64ToBytes(scoreInt + incr)
	}
	batch.Put(l.memberKey(member), newscore)
	batch.Put(l.scoreKey(member, newscore), nil)
	l.redis.db.Write(batch, l.redis.wo)
	return
}

func (l *LevelZSet) RangeByIndex(high2low bool, start, stop int) (scoreMembers [][]byte) {
	direction := IteratorForward
	if high2low {
		direction = IteratorBackward
	}
	min := l.scoreKeyPrefix()
	max := append(l.scoreKeyPrefix(), 254)
	scoreMembers = make([][]byte, 0, 2)
	l.redis.Enumerate(min, max, direction, func(i int, key, value []byte, quit *bool) {
		// fmt.Println(string(min), string(max), start, stop, i, string(key), string(value))
		if i < start {
			return
		} else if i >= start && (stop == -1 || i <= stop) {
			score, member := l.splitScoreKey(key)
			scoreMembers = append(scoreMembers, score)
			scoreMembers = append(scoreMembers, member)
		} else {
			*quit = true
		}
	})
	return
}

func (l *LevelZSet) RangeByScore(high2low bool, min, max []byte, offset, count int) (scoreMembers [][]byte) {
	direction := IteratorForward
	if high2low {
		direction = IteratorBackward
	}
	min2 := bytes.Join([][]byte{l.scoreKeyPrefix(), min}, nil)
	max2 := bytes.Join([][]byte{l.scoreKeyPrefix(), max, []byte{254}}, nil)
	scoreMembers = make([][]byte, 0, 2)
	l.redis.Enumerate(min2, max2, direction, func(i int, key, value []byte, quit *bool) {
		if i < offset { // skip
			return
		}
		if count != -1 && i >= offset+count {
			*quit = true
			return
		}
		score, member := l.splitScoreKey(key)
		scoreMembers = append(scoreMembers, score)
		scoreMembers = append(scoreMembers, member)
	})
	return
}

func (l *LevelZSet) Remove(members ...[]byte) (n int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.initOnce()
	batch := new(leveldb.Batch)
	for _, member := range members {
		score, err := l.redis.db.Get(l.memberKey(member), l.redis.ro)
		if err != nil || score == nil {
			continue
		}
		batch.Delete(l.memberKey(member))
		batch.Delete(l.scoreKey(member, score))
		n++
	}
	l.totalCount -= n
	if l.totalCount == 0 {
		batch.Delete(l.zsetKey())
	} else {
		batch.Put(l.zsetKey(), l.zsetValue())
	}
	l.redis.db.Write(batch, l.redis.wo)
	return
}

func (l *LevelZSet) RemoveByIndex(start, stop int) (n int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.initOnce()
	min := l.scoreKeyPrefix()
	max := append(l.scoreKeyPrefix(), 254)
	batch := new(leveldb.Batch)
	l.redis.Enumerate(min, max, IteratorForward, func(i int, key, value []byte, quit *bool) {
		if i < start {
			return
		} else if i >= start && (stop == -1 || i <= stop) {
			score, member := l.splitScoreKey(key)
			batch.Delete(l.memberKey(member))
			batch.Delete(l.scoreKey(member, score))
			n++
		} else {
			*quit = true
		}
	})
	l.totalCount -= n
	if l.totalCount == 0 {
		batch.Delete(l.zsetKey())
	} else {
		batch.Put(l.zsetKey(), l.zsetValue())
	}
	l.redis.db.Write(batch, l.redis.wo)
	return
}

func (l *LevelZSet) RemoveByScore(min, max []byte) (n int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.initOnce()
	min2 := bytes.Join([][]byte{l.scoreKeyPrefix(), min}, nil)
	max2 := bytes.Join([][]byte{l.scoreKeyPrefix(), max, []byte{254}}, nil)
	batch := new(leveldb.Batch)
	l.redis.Enumerate(min2, max2, IteratorForward, func(i int, key, value []byte, quit *bool) {
		score, member := l.splitScoreKey(key)
		batch.Delete(l.memberKey(member))
		batch.Delete(l.scoreKey(member, score))
		n++
	})
	l.totalCount -= n
	if l.totalCount == 0 {
		batch.Delete(l.zsetKey())
	} else {
		batch.Put(l.zsetKey(), l.zsetValue())
	}
	l.redis.db.Write(batch, l.redis.wo)
	return
}

func (l *LevelZSet) Len() (n int) {
	l.initOnce()
	return l.totalCount
}

func (l *LevelZSet) Drop() (ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.totalCount == 0 {
		return true
	}
	batch := new(leveldb.Batch)
	min := joinStringBytes(KEY_PREFIX, SEP_LEFT, l.key, SEP_RIGHT)
	max := append(min, 254)
	l.redis.Enumerate(min, max, IteratorForward, func(i int, key, value []byte, quit *bool) {
		batch.Delete(key)
	})
	batch.Delete(l.zsetKey())
	l.redis.db.Write(batch, l.redis.wo)
	l.totalCount = 0
	ok = true
	return
}
