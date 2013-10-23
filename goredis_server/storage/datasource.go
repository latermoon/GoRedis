package storage

// 数据源接口
type DataSource interface {
	Get(key []byte) (entry Entry)
	Set(key []byte, entry Entry) (err error)
	Keys(pattern string) (keys []string)
	Remove(key []byte) (err error)
	// 通知数据源，某个条目内容更新了
	NotifyUpdate(key []byte, event interface{})
}
