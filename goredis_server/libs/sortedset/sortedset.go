package sortedset

import (
	"../skiplist"
	"container/list"
	"sync"
)

/*
SotredSet的实现
1. member不可重复，使用map的key存放member，value存放score
2. score做索引，可以重复，使用skiplist的key存放score，value存放member数组
11 [B C]
100 [A]
103.1 [D]

当应用场景里score基本不重复，性能问题不大，但如果score重复太多，会导致value数组过大，不利于修改
因此可以根据value数组的大小选择合适的数据结构
*/
type SortedSet struct {
	skiplist *skiplist.SkipList
	table    map[string]float64
	mutex    sync.Mutex
}

type Iterator skiplist.Iterator

func NewSortedSet() (s *SortedSet) {
	s = &SortedSet{}
	s.skiplist = skiplist.NewCustomMap(func(l, r interface{}) bool {
		return l.(float64) < r.(float64)
	})
	s.table = make(map[string]float64)
	return
}

func (s *SortedSet) Add(member string, score float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	oldscore, exist := s.table[member]
	if exist {
		s.findAndRemove(oldscore, member)
	}
	s.findAndAdd(score, member)
	s.table[member] = score
	return
}

func (s *SortedSet) Iterator() (iter Iterator) {
	return s.skiplist.Iterator()
}

func (s *SortedSet) Len() int {
	return len(s.table)
}

func (s *SortedSet) Table() map[string]float64 {
	return s.table
}

func (s *SortedSet) Score(member string) (score float64, exist bool) {
	score, exist = s.table[member]
	return
}

func (s *SortedSet) IncrScore(member string, incr float64) (score float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	oldscore, exist := s.table[member]
	if exist {
		score = oldscore + incr
		s.findAndRemove(oldscore, member)
	} else {
		score = incr
	}
	s.findAndAdd(score, member)
	return
}

// Return a range of members in a sorted set, by index
func (s *SortedSet) RangeByIndex(start, stop int, withScore bool) (memberAndScores []interface{}) {
	return
}

// Return a range of members in a sorted set, by index, with scores ordered from high to low
func (s *SortedSet) RevRangeByIndex(start, stop int, withScore bool) (memberAndScores []interface{}) {
	return
}

// Return a range of members in a sorted set, by score
func (s *SortedSet) RangeByScore(min, max float64) (iter Iterator) {
	// skiplist实现了大于等于min，少于max，所以这里要增加预设的最小精度，实现少于等于max
	iter = s.skiplist.Range(min, max+0.000000000001)
	return
}

// Return a range of members in a sorted set, by score, with scores ordered from high to low
func (s *SortedSet) RevRangeByScore(min, max float64) (iter Iterator) {
	iter = s.skiplist.Range(min, max+0.000000000001)
	return
}

func (s *SortedSet) Remove(member string) (ok bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	score, exist := s.table[member]
	if exist {
		delete(s.table, member)
		s.findAndRemove(score, member)
		ok = true
	}
	return
}

// Remove all members in a sorted set within the given indexes
func (s *SortedSet) RemoveByIndex(start, stop int) {

}

// Remove all members in a sorted set within the given scores
func (s *SortedSet) RemoveByScore(min, max float64) (n int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// TODO 这个方法有点低效，需要改造一下skiplist，暂时没改
	iter := s.skiplist.Range(min, max+0.000000000001)
	// 先用list存放要删除的member
	lst := list.New()
	for iter.Next() {
		lst.PushBack(iter.Key())
	}
	n = 0
	for e := lst.Front(); e != nil; e = e.Next() {
		val, ok := s.skiplist.Delete(e.Value)
		if ok {
			arr := val.([]string)
			for _, member := range arr {
				delete(s.table, member)
				n++
			}
		}
	}
	return
}

// 从skiplist中移除member
func (s *SortedSet) findAndRemove(score float64, member string) {
	val, ok := s.skiplist.Get(score)
	if !ok {
		return
	}
	arr := val.([]string)
	if len(arr) == 1 {
		s.skiplist.Delete(score)
	} else {
		another := make([]string, 0, len(arr)-1)
		for _, elem := range arr {
			if elem != member {
				another = append(another, elem)
			}
		}
		s.skiplist.Set(score, another)
	}
}

// 添加member到skiplist
func (s *SortedSet) findAndAdd(score float64, member string) {
	val, ok := s.skiplist.Get(score)
	if !ok {
		s.skiplist.Set(score, []string{member})
	} else {
		arr := val.([]string)
		another := append(arr, member)
		s.skiplist.Set(score, another)
	}
}
