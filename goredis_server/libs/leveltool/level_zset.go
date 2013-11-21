package leveltool

/**
基于leveldb实现的zset，用于海量存储，并节约内存
prefix:count = 2
prefix:score:1378000907596:0 = 100422
prefix:score:1378000907596:2 = 100428

prefix:key:100422 = 1378000907596:0
prefix:key:300000 = 1378000907596:2

*/

import (
	"bytes"
	// "fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"strconv"
	"strings"
	"sync"
)

// score和member都是字符串
type ZSetElem struct {
	Score  []byte
	Member []byte
}

type LevelSortedSet struct {
	db         *leveldb.DB
	ro         *opt.ReadOptions
	wo         *opt.WriteOptions
	prefix     string
	totalCount int
	mu         sync.Mutex
}

func NewLevelSortedSet(db *leveldb.DB, prefix string) (l *LevelSortedSet) {
	l = &LevelSortedSet{}
	l.db = db
	l.ro = &opt.ReadOptions{}
	l.wo = &opt.WriteOptions{}
	l.prefix = prefix
	// init totalCount
	data, _ := l.db.Get([]byte(l.countKey()), l.ro)
	if data != nil {
		l.totalCount, _ = strconv.Atoi(string(data))
	} else {
		l.totalCount = 0
	}
	return
}

func NewZSetElem(score, member []byte) (elem *ZSetElem) {
	elem = &ZSetElem{Score: score, Member: member}
	return
}

func (l *LevelSortedSet) countKey() (key string) {
	return l.prefix + ":count"
}
func (l *LevelSortedSet) baseScoreKey(score []byte) (key string) {
	return l.prefix + ":score:" + string(score) + ":"
}
func (l *LevelSortedSet) scoreKeyNoPrefix(scorekey string) (key string) {
	return scorekey[len(l.prefix+":score:"):]
}
func (l *LevelSortedSet) joinScoreKeyWithSuffix(suffix string) (key string) {
	return l.prefix + ":score:" + suffix
}
func (l *LevelSortedSet) memberKey(member []byte) (key string) {
	return l.prefix + ":member:" + string(member)
}

func (l *LevelSortedSet) Add(elems ...*ZSetElem) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, elem := range elems {
		// update
		memberkey := l.memberKey(elem.Member)
		exist := l.removeMember(elem.Member)
		if !exist {
			l.totalCount++
		}
		// seek
		scorekey, _ := l.findScoreKey(elem.Score, elem.Member)
		countkey := l.countKey()
		// fmt.Println("add", scorekey, memberkey, countkey)
		// insert
		batch := new(leveldb.Batch)
		batch.Put([]byte(scorekey), elem.Member)
		batch.Put([]byte(memberkey), []byte(l.scoreKeyNoPrefix(scorekey)))
		batch.Put([]byte(countkey), []byte(strconv.Itoa(l.totalCount)))
		err = l.db.Write(batch, l.wo)
		if err != nil {
			break
		}
	}
	return
}

/**
demo1
score:10000:
score:10002:0 = 100428

demo2
score:10000:
score:10000:0 = 100422
score:10000:2 = 100423
score:10002:0 = 100428
*/
// @param key 完整的key
// @param idx 后续索引
func (l *LevelSortedSet) findScoreKey(score []byte, member []byte) (key string, idx int) {
	iter := l.db.NewIterator(l.ro)
	defer iter.Release()
	basescorekey := []byte(l.baseScoreKey(score))
	if !iter.Seek(basescorekey) {
		key = string(basescorekey) + "0"
		idx = 0
	} else {
		var lastPrefixkey []byte
		for {
			lastkey := iter.Key()
			lastval := iter.Value()
			// fmt.Println("seek", string(lastkey), string(lastval))
			if bytes.HasPrefix(lastkey, basescorekey) {
				lastPrefixkey = copyBytes(lastkey)
				if bytes.Compare(lastval, member) == 0 {
					key = string(lastkey)
					idx, _ = strconv.Atoi(key[strings.LastIndex(key, ":")+1:])
					break
				}
				if !iter.Next() {
					key, idx = l.incrScoreKey(string(lastkey))
					break
				}
			} else {
				if lastPrefixkey == nil {
					key = string(basescorekey) + "0"
					idx = 0
				} else {
					key, idx = l.incrScoreKey(string(lastPrefixkey))
				}
				break
			}
		}
	}
	return
}

// @param member 注意这是member，不是memberkey
func (l *LevelSortedSet) removeMember(member []byte) (ok bool) {
	memberkey := l.memberKey(member)
	suffix, err := l.db.Get([]byte(memberkey), l.ro)
	if err != nil {
		return false
	}
	if suffix == nil {
		return false
	}
	// batch remove
	scorekey := l.joinScoreKeyWithSuffix(string(suffix))
	batch := new(leveldb.Batch)
	batch.Delete([]byte(memberkey))
	batch.Delete([]byte(scorekey))
	err = l.db.Write(batch, l.wo)
	if err != nil {
		return false
	}
	ok = true
	return
}

func copyBytes(src []byte) (dst []byte) {
	dst = make([]byte, len(src))
	copy(dst, src)
	return
}

func (l *LevelSortedSet) incrScoreKey(scorekey string) (key string, idx int) {
	pos := strings.LastIndex(scorekey, ":")
	prefix := scorekey[:pos]
	idx, _ = strconv.Atoi(scorekey[pos+1:])
	idx++
	key = prefix + ":" + strconv.Itoa(idx)
	return
}

func (l *LevelSortedSet) RangeByIndex(start, stop int, withScores bool) (elems []*ZSetElem) {
	return
}

func (l *LevelSortedSet) RangeByScore(min, max []byte, withScores bool, limit int) (elem []*ZSetElem) {
	return
}

func (l *LevelSortedSet) Remove(members ...[]byte) {

}

func (l *LevelSortedSet) RemoveByIndex(start, stop int) (n int) {
	return
}

func (l *LevelSortedSet) RemoveByScore(min, max []byte) (n int) {
	return
}

func (l *LevelSortedSet) Score(member []byte) (score []byte) {
	return
}

func (l *LevelSortedSet) Count() (n int) {
	return l.totalCount
}

// ==============================
// ==============================
