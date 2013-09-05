package storage

// 数据类型
type KeyType int

const (
	KeyTypeUnknown = iota
	KeyTypeString
	KeyTypeHash
	KeyTypeList
	KeyTypeSet
	KeyTypeSortedSet
)

// 存储支持
type RedisStorages struct {
	KeyTypeStorage KeyTypeStorage
	StringStorage  StringStorage
	HashStorage    HashStorage
	ListStorage    ListStorage
	SetStorage     SetStorage
}

type KeyTypeStorage interface {
	GetType(key string) (keytype int, err error)
	SetType(key string, keytype int) (err error)
}

type StringStorage interface {
	Get(key string) (value interface{}, err error)
	Set(key string, value string) (err error)
	MGet(keys ...string) (values []interface{}, err error)
	MSet(keyvals ...string) (err error)
	Del(keys ...string) (n int, err error)
}

type HashStorage interface {
	HGet(key string, field string) (value string, err error)
	HSet(key string, field string, value string) (result int, err error)
	HGetAll(key string) (values []interface{}, err error)
	HDel(key string, fields ...string) (n int, err error)
	Del(keys ...string) (n int, err error)
}

type ListStorage interface {
	LPop(key string) (value string, err error)
	LPush(key string, value string) (n int, err error)
	RPop(key string) (value string, err error)
	RPush(key string, value string) (n int, err error)
	LRange(key string, start int, end int) (values []interface{}, err error)
	LLen(key string) (length int, err error)
	Del(keys ...string) (n int, err error)
}

type SetStorage interface {
	SAdd(key string, members ...string) (n int, err error)
	SCard(key string) (count int, err error)
	SIsMember(key string, member string) (n int, err error)
	SMembers(key string) (values []interface{}, err error)
	SRem(key string, members ...string) (n int, err error)
	Del(keys ...string) (n int, err error)
}
