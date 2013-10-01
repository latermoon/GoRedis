package goredis_server

import (
	. "../goredis"
	. "./storage"
)

func (server *GoRedisServer) OnHGET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	field := cmd.StringAtIndex(2)
	entry, _ := server.datasource.Get(key)
	if entry != nil {
		reply = BulkReply(nil)
	} else if entry.Type() == EntryTypeHash {
		hashentry := entry.(*HashEntry)
		hashentry.Mutex.Lock()
		val := hashentry.Get(field)
		hashentry.Mutex.Unlock()
		reply = BulkReply(val)
	} else {
		reply = WrongKindReply
	}
	return
}

func (server *GoRedisServer) OnHSET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	field := cmd.StringAtIndex(2)
	value := cmd.StringAtIndex(3)
	entry, _ := server.datasource.Get(key)
	var hashentry *HashEntry
	if entry != nil {
		hashentry = entry.(*HashEntry)
	} else {
		hashentry = NewHashEntry()
		server.datasource.Set(key, hashentry)
	}

	hashentry.Mutex.Lock()
	hashentry.Set(field, value)
	hashentry.Mutex.Unlock()

	// update
	server.datasource.NotifyEntryUpdate(key, hashentry)
	reply = IntegerReply(1)
	return
}

func (server *GoRedisServer) OnHGETALL(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, _ := server.datasource.Get(key)
	if entry != nil {
		reply = MultiBulksReply([]interface{}{})
	} else if entry.Type() == EntryTypeHash {
		hashentry := entry.(*HashEntry)
		// response
		keyvals := make([]interface{}, 0, len(hashentry.Map())*2)
		hashentry.Mutex.Lock()
		for key, val := range hashentry.Map() {
			keyvals = append(keyvals, key)
			keyvals = append(keyvals, val)
		}
		hashentry.Mutex.Unlock()
		reply = MultiBulksReply(keyvals)
	} else {
		reply = WrongKindReply
	}
	return
}

func (server *GoRedisServer) OnHMGET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	fields := cmd.StringArgs()[2:]
	entry, _ := server.datasource.Get(key)
	if entry == nil {
		reply = MultiBulksReply(make([]interface{}, len(fields)))
	} else if entry.Type() == EntryTypeHash {
		keyvals := make([]interface{}, 0, len(fields))
		for _, field := range fields {
			val := entry.(*HashEntry).Get(field)
			keyvals = append(keyvals, field)
			keyvals = append(keyvals, val)
		}
		reply = MultiBulksReply(keyvals)
	} else {
		reply = WrongKindReply
	}
	return
}

func (server *GoRedisServer) OnHMSET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	keyvals := cmd.StringArgs()[2:]
	if len(keyvals)%2 != 0 {
		reply = ErrorReply("Bad field/value paires")
		return
	}
	entry, _ := server.datasource.Get(key)
	var hashentry *HashEntry
	if entry != nil {
		hashentry = entry.(*HashEntry)
	} else {
		hashentry = NewHashEntry()
		server.datasource.Set(key, hashentry)
	}
	for i := 0; i < len(keyvals); i += 2 {
		field := keyvals[i]
		val := keyvals[i+1]
		hashentry.Set(field, val)
	}
	// update
	server.datasource.NotifyEntryUpdate(key, hashentry)
	reply = StatusReply("OK")
	return
}

func (server *GoRedisServer) OnHLEN(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry, _ := server.datasource.Get(key)
	if entry == nil {
		reply = IntegerReply(0)
	} else if entry.Type() == EntryTypeHash {
		hashentry := entry.(*HashEntry)
		hashentry.Mutex.Lock()
		length := len(hashentry.Map())
		hashentry.Mutex.Unlock()
		reply = IntegerReply(length)
	} else {
		reply = WrongKindReply
	}
	return
}

func (server *GoRedisServer) OnHDEL(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	fields := cmd.StringArgs()[2:]
	entry, _ := server.datasource.Get(key)
	if entry != nil {
		reply = IntegerReply(0)
	} else if entry.Type() == EntryTypeHash {
		hashentry := entry.(*HashEntry)
		hashentry.Mutex.Lock()
		n := 0
		for _, field := range fields {
			_, exist := hashentry.Map()[field]
			if exist {
				delete(hashentry.Map(), field)
				n++
			}
		}
		hashentry.Mutex.Unlock()
		if len(hashentry.Map()) == 0 {
			server.datasource.Remove(key)
		} else {
			// update
			server.datasource.NotifyEntryUpdate(key, hashentry)
		}
		reply = IntegerReply(n)
	} else {
		reply = WrongKindReply
	}

	return
}
