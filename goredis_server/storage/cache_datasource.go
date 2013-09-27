package storage

import (
	"sync"
)

type CacheDataSource struct {
	DataSource
	table map[string]interface{}
	mutex sync.Mutex
}

func NewCacheDataSource() (m *CacheDataSource) {
	m = &CacheDataSource{}
	m.table = make(map[string]interface{})
	return
}

func (m *CacheDataSource) GetObject(key string) (val interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	val = m.table[key]
	return
}

func (m *CacheDataSource) SetObject(key string, val interface{}) (err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.table[key] = val
	return
}

func (m *CacheDataSource) DelObject(key string) (err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.table, key)
	return
}
