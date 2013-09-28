package storage

// 数据源接口
type DataSource interface {
	Get(key string) (entry Entry, exist bool)
	Set(key string, entry Entry) (err error)
	Remove(key string) (err error)
}
