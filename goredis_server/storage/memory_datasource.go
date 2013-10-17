package storage

import (
	"sync"
)

// 内存数据源
type MemoryDataSource struct {
	DataSource
	table map[string]Entry
	mutex sync.Mutex
}

func NewMemoryDataSource() (m *MemoryDataSource) {
	m = &MemoryDataSource{}
	m.table = make(map[string]Entry)
	return
}

func (m *MemoryDataSource) Get(key []byte) (entry Entry) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	entry = m.table[string(key)]
	return
}

func (m *MemoryDataSource) Set(key []byte, entry Entry) (err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.table[string(key)] = entry
	return
}

func (m *MemoryDataSource) Keys(pattern string) (keys []string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	keys = make([]string, 0, len(m.table))
	for key, _ := range m.table {
		keys = append(keys, key)
	}
	return
}

func (m *MemoryDataSource) Remove(key []byte) (err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.table, string(key))
	return
}

func (m *MemoryDataSource) NotifyEntryUpdate(key []byte, entry Entry) {
	// do nothing
}
