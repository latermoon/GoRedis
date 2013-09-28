package storage

import (
	"sync"
)

type MemoryDataSource struct {
	DataSource
	table map[string]*Entry
	mutex sync.Mutex
}

func NewMemoryDataSource() (m *MemoryDataSource) {
	m = &MemoryDataSource{}
	m.table = make(map[string]*Entry)
	return
}

func (m *MemoryDataSource) Get(key string) (entry *Entry, exist bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	entry, exist = m.table[key]
	return
}

func (m *MemoryDataSource) Set(key string, entry *Entry) (err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.table[key] = entry
	return
}

func (m *MemoryDataSource) Remove(key string) (err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.table, key)
	return
}
