package storage

type Storage interface {
	Set(key string, value string) (err error)
	Get(key string) (value string, err error)
}
