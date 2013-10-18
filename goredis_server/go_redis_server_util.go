package goredis_server

import (
	// . "../goredis"
	. "./storage"
	"strconv"
)

func (server *GoRedisServer) formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'g', 12, 64)
}

func (server *GoRedisServer) slavesEntry() (entry *SetEntry) {
	entry = server.getConfigEntry("slaves", EntryTypeSet).(*SetEntry)
	return
}

// 发送snapshot完成后的回调
func (server *GoRedisServer) snapshotSentCallback(session *SlaveSession) {
	entry := server.slavesEntry()
	if session.AofEnabled() {
		entry.Put(session.UID())
	}
	server.setConfigEntry("slaves", entry)
}

func (server *GoRedisServer) getConfigEntry(key string, et EntryType) (entry Entry) {
	dbkey := []byte(goredisPrefix + key)
	entry = server.datasource.Get(dbkey)
	if entry == nil {
		entry = NewEmptyEntry(et)
		server.datasource.Set(dbkey, entry)
	}
	return
}

func (server *GoRedisServer) setConfigEntry(key string, entry Entry) {
	dbkey := []byte(goredisPrefix + key)
	server.datasource.Set(dbkey, entry)
}
