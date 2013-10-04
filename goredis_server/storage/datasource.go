package storage

// 数据源接口
type DataSource interface {
	Get(key string) (entry Entry)
	Set(key string, entry Entry) (err error)
	Keys(pattern string) (keys []string)
	Remove(key string) (err error)
	// 通知数据源，某个条目更新了
	NotifyEntryUpdate(key string, entry Entry)
}

// 测试内容
type DataSourceSeeker interface {
	Interator() EntryIterator
}

type EntryIterator interface {
	Seek(key string)
	Key() string
	Entry() Entry
	Prev() bool
	Next() bool
}
