package goredis_server

import (
	. "../goredis"
	. "./storage"
	"fmt"
)

func entryToCommand(key []byte, entry Entry) (cmd *Command) {
	args := make([][]byte, 0, 10)

	switch entry.Type() {
	case EntryTypeString:
		args = append(args, []byte("set"))
		args = append(args, key)
		args = append(args, entry.(*StringEntry).Bytes())
	case EntryTypeHash:
		table := entry.(*HashEntry).Map()
		args = append(args, []byte("hmset"))
		args = append(args, key)
		for field, value := range table {
			switch value.(type) {
			case string:
				args = append(args, []byte(field))
				args = append(args, []byte(value.(string)))
			case []byte:
				args = append(args, []byte(field))
				args = append(args, value.([]byte))
			default:
				fmt.Println("bad hset type", field, value)
			}

		}
	case EntryTypeSortedSet:

	case EntryTypeSet:

	case EntryTypeList:

	default:
	}
	if len(args) > 0 {
		cmd = NewCommand(args...)
	}
	return
}
