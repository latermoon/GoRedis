package goredis_server

import (
	. "../goredis"
	. "./storage"
	"bytes"
	"fmt"
)

var typeTable = map[EntryType]string{EntryTypeString: "string", EntryTypeHash: "hash", EntryTypeList: "list", EntryTypeSet: "set", EntryTypeSortedSet: "zset"}

func (server *GoRedisServer) OnDEL(cmd *Command) (reply *Reply) {
	keys := cmd.StringArgs()[1:]
	n := 0
	for _, key := range keys {
		entry := server.datasource.Get(key)
		if entry != nil {
			server.datasource.Remove(key)
			n++
		}
	}
	reply = IntegerReply(n)
	return
}

func (server *GoRedisServer) OnTYPE(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry := server.datasource.Get(key)
	if entry != nil {
		if desc, exist := typeTable[entry.Type()]; exist {
			return StatusReply(desc)
		}
	}
	return StatusReply("none")
}

// [Custom]，单独保存一个key
func (server *GoRedisServer) OnSAVEKEY(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	entry := server.datasource.Get(key)
	if entry == nil {
		return ErrorReply(key + " not exist")
	}
	err := server.datasource.Set(key, entry)
	reply = ReplySwitch(err, StatusReply("OK"))
	return
}

// [Custom] 描述一个key
func (server *GoRedisServer) OnDESC(cmd *Command) (reply *Reply) {
	keys := cmd.StringArgs()[1:]
	bulks := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		entry := server.datasource.Get(key)
		if entry == nil {
			bulks = append(bulks, key+" [nil]")
			continue
		}
		buf := bytes.Buffer{}
		buf.WriteString(key + " ")
		switch entry.Type() {
		case EntryTypeString:
			buf.WriteString("[String] ")
			buf.WriteString(entry.(*StringEntry).String())
		case EntryTypeHash:
			buf.WriteString("[Hash] ")
			buf.WriteString(fmt.Sprintf("%s", entry.(*HashEntry).Map()))
		case EntryTypeSortedSet:
			buf.WriteString("[SortedSet] ")
			iter := entry.(*SortedSetEntry).SortedSet().Iterator()
			for iter.Next() {
				members := iter.Value().([]string)
				score := iter.Key().(float64)
				for _, member := range members {
					buf.WriteString(fmt.Sprintf("%s(%s) ", member, server.formatFloat(score)))
				}
			}
		case EntryTypeSet:
			buf.WriteString("[Set] ")
			buf.WriteString(fmt.Sprintf("%s", entry.(*SetEntry).Keys()))
		case EntryTypeList:
			buf.WriteString("[List] ")
			for e := entry.(*ListEntry).List().Front(); e != nil; e = e.Next() {
				buf.WriteString(e.Value.(string))
				buf.WriteString(" ")
			}
		default:
			buf.WriteString(key + " [unknown]")
		}
		// append
		bulks = append(bulks, buf.String())
	}
	reply = MultiBulksReply(bulks)
	return
}
