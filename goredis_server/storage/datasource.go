package storage

// 数据源接口
type DataSource interface {
	Get(key string) (entry Entry)
	Set(key string, entry Entry) (err error)
	Remove(key string) (err error)
	// 通知数据源，某个条目更新了
	NotifyEntryUpdate(key string, entry Entry)
}
