package storage

import (
	"sync"
)

type BufferDataSource struct {
	DataSource
	// 下层
	backend DataSource
	mutex   sync.Mutex
	cache   map[string]Entry
}

func NewBufferDataSource(innerds DataSource) (ds *BufferDataSource) {
	ds = &BufferDataSource{}
	ds.backend = innerds
	ds.cache = make(map[string]Entry)
	return
}

func (ds *BufferDataSource) Get(key string) (entry Entry) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()
	var exist bool
	if entry, exist = ds.cache[key]; !exist {
		entry = ds.backend.Get(key)
		ds.cache[key] = entry
	}
	return
}

func (ds *BufferDataSource) Set(key string, entry Entry) (err error) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()
	err = ds.backend.Set(key, entry)
	return
}

func (ds *BufferDataSource) Keys(pattern string) (keys []string) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()
	keys = ds.backend.Keys(pattern)
	return
}

func (ds *BufferDataSource) Remove(key string) (err error) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()
	err = ds.backend.Remove(key)
	return
}

func (ds *BufferDataSource) NotifyEntryUpdate(key string, entry Entry) {
	ds.NotifyEntryUpdate(key, entry)
}
