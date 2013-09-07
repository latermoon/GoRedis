package storage

import (
	"sync"
)

type MemoryListStorage struct {
	ListStorage
	kvCache map[string]*SafeList
	mutex   *sync.Mutex
}

func NewMemoryListStorage() (storage *MemoryListStorage) {
	storage = &MemoryListStorage{}
	storage.kvCache = make(map[string]*SafeList)
	storage.mutex = &sync.Mutex{}
	return
}

// 获取指定key的列表，不存在时自动创建
func (l *MemoryListStorage) listByKey(key string) (sl *SafeList) {
	l.mutex.Lock()
	var exists bool
	sl, exists = l.kvCache[key]
	if !exists {
		sl = NewSafeList()
		l.kvCache[key] = sl
	}
	l.mutex.Unlock()
	return
}

func (l *MemoryListStorage) LPop(key string) (value interface{}, err error) {
	sl := l.listByKey(key)
	value = sl.LPop()
	return
}

func (l *MemoryListStorage) LPush(key string, values ...string) (n int, err error) {
	sl := l.listByKey(key)
	n = sl.LPush(values...)
	return
}

func (l *MemoryListStorage) RPop(key string) (value interface{}, err error) {
	sl := l.listByKey(key)
	value = sl.RPop()
	return
}

func (l *MemoryListStorage) RPush(key string, values ...string) (n int, err error) {
	sl := l.listByKey(key)
	n = sl.RPush(values...)
	return
}

func (l *MemoryListStorage) LRange(key string, start int, end int) (values []interface{}, err error) {
	sl := l.listByKey(key)
	values = sl.Range(start, end)
	return
}

func (l *MemoryListStorage) LIndex(key string, index int) (value interface{}, err error) {
	sl := l.listByKey(key)
	value = sl.Index(index)
	return
}

func (l *MemoryListStorage) LLen(key string) (length int, err error) {
	length = l.listByKey(key).Len()
	return
}

func (l *MemoryListStorage) Del(keys ...string) (n int, err error) {
	l.mutex.Lock()
	n = 0
	for _, key := range keys {
		_, exists := l.kvCache[key]
		if exists {
			delete(l.kvCache, key)
			n++
		}
	}
	l.mutex.Unlock()
	return
}
