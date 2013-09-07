package storage

// 内存结构
type MapItem struct {
	Lock chan int
	Map  map[string]interface{}
}

func NewMapItem() (m *MapItem) {
	m = &MapItem{}
	m.Lock = make(chan int, 1)
	m.Map = make(map[string]interface{})
	return
}

// 暂时实现的map操作有同步问题，上线前需要修改
type MemoryHashStorage struct {
	HashStorage
	kvCache map[string]*MapItem
	kvLock  chan int
}

func NewMemoryHashStorage() (storage *MemoryHashStorage) {
	storage = &MemoryHashStorage{}
	storage.kvCache = make(map[string]*MapItem)
	storage.kvLock = make(chan int, 1)
	return
}

func (s *MemoryHashStorage) mapByKey(key string) (m *MapItem) {
	s.kvLock <- 1
	var exists bool
	m, exists = s.kvCache[key]
	if !exists {
		m = NewMapItem()
		s.kvCache[key] = m
	}
	<-s.kvLock
	return
}

func (s *MemoryHashStorage) HGet(key string, field string) (value interface{}, err error) {
	m := s.mapByKey(key)
	m.Lock <- 1
	value, _ = m.Map[field]
	<-m.Lock
	return
}

// http://redis.io/commands/hset
// Integer reply, specifically:
// 1 if field is a new field in the hash and value was set.
// 0 if field already exists in the hash and the value was updated.
func (s *MemoryHashStorage) HSet(key string, field string, value string) (result int, err error) {
	m := s.mapByKey(key)
	m.Lock <- 1
	_, exists := m.Map[field]
	m.Map[field] = value
	if !exists {
		result = 1
	} else {
		result = 0
	}
	<-m.Lock
	return
}

// http://redis.io/commands/hgetall
func (s *MemoryHashStorage) HGetAll(key string) (keyvals []interface{}, err error) {
	m := s.mapByKey(key)
	m.Lock <- 1
	length := len(m.Map)
	keyvals = make([]interface{}, 0, length*2) // len(keys)+len(vals)
	for k, v := range m.Map {
		keyvals = append(keyvals, k)
		keyvals = append(keyvals, v)
	}
	<-m.Lock
	return
}

func (s *MemoryHashStorage) HDel(key string, fields ...string) (n int, err error) {
	m := s.mapByKey(key)
	m.Lock <- 1
	n = 0
	for _, field := range fields {
		_, exists := m.Map[field]
		if exists {
			n++
			delete(m.Map, field)
		}
	}
	<-m.Lock
	return
}

func (s *MemoryHashStorage) HMGet(key string, fields ...string) (values []interface{}, err error) {
	m := s.mapByKey(key)
	m.Lock <- 1
	values = make([]interface{}, len(fields))
	for i, field := range fields {
		values[i], _ = m.Map[field]
	}
	<-m.Lock
	return
}

func (s *MemoryHashStorage) HMSet(key string, keyvals ...string) (err error) {
	m := s.mapByKey(key)
	m.Lock <- 1
	count := len(keyvals)
	for i := 0; i < count; i += 2 {
		k := keyvals[i]
		v := keyvals[i+1]
		m.Map[k] = v
	}
	<-m.Lock
	return
}

func (s *MemoryHashStorage) HLen(key string) (length int, err error) {
	m := s.mapByKey(key)
	m.Lock <- 1
	length = len(m.Map)
	<-m.Lock
	return
}

func (s *MemoryHashStorage) Del(keys ...string) (n int, err error) {
	s.kvLock <- 1
	n = 0
	for _, key := range keys {
		_, exists := s.kvCache[key]
		if exists {
			n++
			delete(s.kvCache, key)
		}
	}
	<-s.kvLock
	return
}
