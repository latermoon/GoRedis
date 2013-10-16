package storage

import (
	"sync"
)

type BufferDataSource struct {
	DataSource
	// 下层
	backend DataSource
	mutex   sync.Mutex
}

func NewBufferDataSource(ds DataSource) (bufds *BufferDataSource) {
	bufds = &BufferDataSource{}
	bufds.backend = ds
	return
}

func (ds *BufferDataSource) Get(key string) (entry Entry) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()
	return
}

func (ds *BufferDataSource) Set(key string, entry Entry) (err error) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()
	return
}

func (ds *BufferDataSource) Keys(pattern string) (keys []string) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()
	return
}

func (ds *BufferDataSource) Remove(key string) (err error) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()
	return
}

func (ds *BufferDataSource) NotifyEntryUpdate(key string, entry Entry) {
	// do nothing
}
