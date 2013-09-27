package storage

import (
	"sync"
)

// 基于内存的Storage
type MemoryStorage struct {
	datamap map[string]interface{}
	mutex   *sync.Mutex
}

func NewMemoryStorage() (storage *MemoryStorage) {
	storage = &MemoryStorage{}
	storage.datamap = make(map[string]interface{})
	storage.mutex = &sync.Mutex{}
	return
}

// ============================================================
// 						String
// ============================================================

func (s *MemoryStorage) Get(key string) (value interface{}, err error) {
	value, _ = s.datamap[key]
	return
}

func (s *MemoryStorage) Set(key string, value string) (err error) {
	s.mutex.Lock()
	s.datamap[key] = value
	s.mutex.Unlock()
	return
}

func (s *MemoryStorage) MGet(keys ...string) (values []interface{}, err error) {
	values = make([]interface{}, len(keys))
	for i, key := range keys {
		values[i] = s.datamap[key]
	}
	return
}

func (s *MemoryStorage) MSet(keyvals ...string) (err error) {
	s.mutex.Lock()
	for i := 0; i < len(keyvals); i += 2 {
		key := keyvals[i]
		value := keyvals[i+1]
		s.datamap[key] = value
	}
	s.mutex.Unlock()
	return
}

// ============================================================
// 						Key
// ============================================================

func (s *MemoryStorage) Del(keys ...string) (n int, err error) {
	s.mutex.Lock()
	n = 0
	for _, key := range keys {
		_, exists := s.datamap[key]
		if exists {
			delete(s.datamap, key)
			n++
		}
	}
	s.mutex.Unlock()
	return
}

func (s *MemoryStorage) TypeOf(key string) (kt KeyType) {
	val, exist := s.datamap[key]
	if !exist {
		kt = KeyTypeUnknown
	} else {
		switch val.(type) {
		case string:
			kt = KeyTypeString
		case *SafeList:
			kt = KeyTypeList
		default:
			kt = KeyTypeUnknown
		}
	}
	return
}

// ============================================================
// 						List
// ============================================================

// 获取指定key的列表，不存在时自动创建
func (s *MemoryStorage) listByKey(key string) (sl *SafeList) {
	s.mutex.Lock()
	obj, exist := s.datamap[key]
	if !exist {
		sl = NewSafeList()
		s.datamap[key] = sl
	} else {
		sl = obj.(*SafeList)
	}
	s.mutex.Unlock()
	return
}

func (s *MemoryStorage) LPop(key string) (value interface{}, err error) {
	sl := s.listByKey(key)
	value = sl.LPop()
	return
}

func (s *MemoryStorage) LPush(key string, values ...string) (n int, err error) {
	sl := s.listByKey(key)
	n = sl.LPush(values...)
	return
}

func (s *MemoryStorage) RPop(key string) (value interface{}, err error) {
	sl := s.listByKey(key)
	value = sl.RPop()
	return
}

func (s *MemoryStorage) RPush(key string, values ...string) (n int, err error) {
	sl := s.listByKey(key)
	n = sl.RPush(values...)
	return
}

func (s *MemoryStorage) LRange(key string, start int, end int) (values []interface{}, err error) {
	sl := s.listByKey(key)
	values = sl.Range(start, end)
	return
}

func (s *MemoryStorage) LIndex(key string, index int) (value interface{}, err error) {
	sl := s.listByKey(key)
	value = sl.Index(index)
	return
}

func (s *MemoryStorage) LLen(key string) (length int, err error) {
	length = s.listByKey(key).Len()
	return
}

// ============================================================
// 						Hash
// ============================================================
// 内存结构
type MapItem struct {
	Map   map[string]interface{}
	Mutex *sync.Mutex
}

func NewMapItem() (m *MapItem) {
	m = &MapItem{}
	m.Map = make(map[string]interface{})
	m.Mutex = &sync.Mutex{}
	return
}

func (s *MemoryStorage) mapByKey(key string) (m *MapItem) {
	s.mutex.Lock()
	obj, exist := s.datamap[key]
	if !exist {
		m = NewMapItem()
		s.datamap[key] = m
	} else {
		m = obj.(*MapItem)
	}
	s.mutex.Unlock()
	return
}

func (s *MemoryStorage) HGet(key string, field string) (value interface{}, err error) {
	m := s.mapByKey(key)
	m.Mutex.Lock()
	value, _ = m.Map[field]
	m.Mutex.Unlock()
	return
}

// http://redis.io/commands/hset
// Integer reply, specifically:
// 1 if field is a new field in the hash and value was set.
// 0 if field already exists in the hash and the value was updated.
func (s *MemoryStorage) HSet(key string, field string, value string) (result int, err error) {
	m := s.mapByKey(key)
	m.Mutex.Lock()
	_, exists := m.Map[field]
	m.Map[field] = value
	if !exists {
		result = 1
	} else {
		result = 0
	}
	m.Mutex.Unlock()
	return
}

// http://redis.io/commands/hgetall
func (s *MemoryStorage) HGetAll(key string) (keyvals []interface{}, err error) {
	m := s.mapByKey(key)
	m.Mutex.Lock()
	length := len(m.Map)
	keyvals = make([]interface{}, 0, length*2) // len(keys)+len(vals)
	for k, v := range m.Map {
		keyvals = append(keyvals, k)
		keyvals = append(keyvals, v)
	}
	m.Mutex.Unlock()
	return
}

func (s *MemoryStorage) HDel(key string, fields ...string) (n int, err error) {
	m := s.mapByKey(key)
	m.Mutex.Lock()
	n = 0
	for _, field := range fields {
		_, exists := m.Map[field]
		if exists {
			n++
			delete(m.Map, field)
		}
	}
	m.Mutex.Unlock()
	return
}

func (s *MemoryStorage) HMGet(key string, fields ...string) (values []interface{}, err error) {
	m := s.mapByKey(key)
	m.Mutex.Lock()
	values = make([]interface{}, len(fields))
	for i, field := range fields {
		values[i], _ = m.Map[field]
	}
	m.Mutex.Unlock()
	return
}

func (s *MemoryStorage) HMSet(key string, keyvals ...string) (err error) {
	m := s.mapByKey(key)
	m.Mutex.Lock()
	count := len(keyvals)
	for i := 0; i < count; i += 2 {
		k := keyvals[i]
		v := keyvals[i+1]
		m.Map[k] = v
	}
	m.Mutex.Unlock()
	return
}

func (s *MemoryStorage) HLen(key string) (length int, err error) {
	m := s.mapByKey(key)
	m.Mutex.Lock()
	length = len(m.Map)
	m.Mutex.Unlock()
	return
}

// ============================================================
// 						Set
// ============================================================
