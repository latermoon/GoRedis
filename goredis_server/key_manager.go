package goredis_server

import (
	// . "../goredis"
	leveltool "./libs/leveltool"
	lru "./libs/lrucache"
	"sync"
)

// 使用lru管理key
type KeyManager struct {
	server   *GoRedisServer
	lruCache *lru.LRUCache // LRU缓存层
	lstring  *leveltool.LevelString
	lkey     *leveltool.LevelKey
	mu       sync.Mutex
}

func NewKeyManager(server *GoRedisServer, capacity uint64) (km *KeyManager) {
	km = &KeyManager{}
	km.server = server
	km.lstring = leveltool.NewLevelString(server.DB()) // string
	km.lkey = leveltool.NewLevelKey(server.DB())       // key
	km.lruCache = lru.NewLRUCache(10000)
	return
}

func (k *KeyManager) objFromCache(key string, fn func() interface{}) (obj interface{}) {
	k.mu.Lock()
	defer k.mu.Unlock()
	var ok bool
	obj, ok = k.lruCache.Get(key)
	if !ok {
		obj = fn()
		k.lruCache.Set(key, obj.(lru.Value))
	}
	return
}

func (k *KeyManager) hashByKey(key string) (item *leveltool.LevelHash) {
	obj := k.objFromCache(key, func() interface{} {
		return leveltool.NewLevelHash(k.server.DB(), key)
	})
	item = obj.(*leveltool.LevelHash)
	return
}

func (k *KeyManager) setByKey(key string) (item *leveltool.LevelHash) {
	return k.hashByKey(key)
}

func (k *KeyManager) listByKey(key string) (item *leveltool.LevelList) {
	obj := k.objFromCache(key, func() interface{} {
		return leveltool.NewLevelList(k.server.DB(), key)
	})
	item = obj.(*leveltool.LevelList)
	return
}

func (k *KeyManager) zsetByKey(key string) (item *leveltool.LevelSortedSet) {
	obj := k.objFromCache(key, func() interface{} {
		return leveltool.NewLevelSortedSet(k.server.DB(), key)
	})
	item = obj.(*leveltool.LevelSortedSet)
	return
}

func (k *KeyManager) levelString() (item *leveltool.LevelString) {
	return k.lstring
}

func (k *KeyManager) levelKey() (item *leveltool.LevelKey) {
	return k.lkey
}
