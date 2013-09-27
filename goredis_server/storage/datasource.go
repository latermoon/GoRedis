package storage

type DataSource interface {
	GetObject(key string) (val interface{})
	SetObject(key string, val interface{}) (err error)
	DelObject(key string) (err error)
}
