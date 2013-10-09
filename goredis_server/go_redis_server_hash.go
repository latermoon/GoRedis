package goredis_server

import (
	. "../goredis"
	. "./storage"
)

// 获取Hash，不存在则自动创建
func (server *GoRedisServer) hashByKey(key string, create bool) (hash *HashEntry, err error) {
	entry := server.datasource.Get(key)
	if entry != nil && entry.Type() != EntryTypeHash {
		err = WrongKindError
		return
	}
	if entry != nil {
		hash = entry.(*HashEntry)
	} else if create {
		hash = NewHashEntry()
		server.datasource.Set(key, hash)
	}
	return
}

func (server *GoRedisServer) OnHGET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, err := server.hashByKey(key, false)
	if err != nil {
		return ErrorReply(err)
	} else if entry == nil {
		return BulkReply(nil)
	}

	field := cmd.StringAtIndex(2)
	val := entry.Get(field)
	reply = BulkReply(val)
	return
}

func (server *GoRedisServer) OnHSET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, err := server.hashByKey(key, true)
	if err != nil {
		return ErrorReply(err)
	}

	entry.Mutex.Lock()
	defer entry.Mutex.Unlock()

	field := cmd.StringAtIndex(2)
	value := cmd.StringAtIndex(3)

	entry.Set(field, value)
	// update
	server.datasource.NotifyEntryUpdate(key, entry)
	return IntegerReply(1)
}

func (server *GoRedisServer) OnHGETALL(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, err := server.hashByKey(key, false)
	if err != nil {
		return ErrorReply(err)
	} else if entry == nil {
		return MultiBulksReply([]interface{}{})
	}

	entry.Mutex.Lock()
	defer entry.Mutex.Unlock()

	// response
	keyvals := make([]interface{}, 0, len(entry.Map())*2)
	for key, val := range entry.Map() {
		keyvals = append(keyvals, key)
		keyvals = append(keyvals, val)
	}
	reply = MultiBulksReply(keyvals)
	return
}

func (server *GoRedisServer) OnHMGET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	fields := cmd.StringArgs()[2:]
	entry, err := server.hashByKey(key, false)
	if err != nil {
		return ErrorReply(err)
	} else if entry == nil {
		return MultiBulksReply(make([]interface{}, len(fields)))
	}

	keyvals := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		val := entry.Get(field)
		keyvals = append(keyvals, field)
		keyvals = append(keyvals, val)
	}
	reply = MultiBulksReply(keyvals)
	return
}

func (server *GoRedisServer) OnHMSET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	keyvals := cmd.StringArgs()[2:]
	if len(keyvals)%2 != 0 {
		reply = ErrorReply("Bad field/value paires")
		return
	}

	entry, err := server.hashByKey(key, true)
	if err != nil {
		return ErrorReply(err)
	}

	entry.Mutex.Lock()
	defer entry.Mutex.Unlock()

	for i := 0; i < len(keyvals); i += 2 {
		field := keyvals[i]
		val := keyvals[i+1]
		entry.Set(field, val)
	}
	// update
	server.datasource.NotifyEntryUpdate(key, entry)
	reply = StatusReply("OK")
	return
}

func (server *GoRedisServer) OnHLEN(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, err := server.hashByKey(key, false)
	if err != nil {
		return ErrorReply(err)
	} else if entry == nil {
		return IntegerReply(0)
	}

	length := len(entry.Map())
	reply = IntegerReply(length)
	return
}

func (server *GoRedisServer) OnHDEL(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, err := server.hashByKey(key, false)
	if err != nil {
		return ErrorReply(err)
	} else if entry == nil {
		return IntegerReply(0)
	}

	entry.Mutex.Lock()
	defer entry.Mutex.Unlock()

	fields := cmd.StringArgs()[2:]
	n := 0
	for _, field := range fields {
		_, exist := entry.Map()[field]
		if exist {
			delete(entry.Map(), field)
			n++
		}
	}

	if len(entry.Map()) == 0 {
		server.datasource.Remove(key)
	} else {
		if n > 0 {
			server.datasource.NotifyEntryUpdate(key, entry)
		}
	}
	reply = IntegerReply(n)
	return
}
