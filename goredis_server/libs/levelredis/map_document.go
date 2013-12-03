package levelredis

// TODO 本类未完成

import (
	"errors"
	"fmt"
	"strings"
)

var (
	WrongKindError   = errors.New("wrong kind error")
	BadArgumentCount = errors.New("bad argument count")
)

// 提供面向document操作的map
type MapDocument struct {
	data map[string]interface{}
}

func NewMapDocument(data map[string]interface{}) (m *MapDocument) {
	m = &MapDocument{}
	m.data = data
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	return
}

// doc_set(key, {"name":"latermoon", "$rpush":["photos", "c.jpg", "d.jpg"], "$incr":["version", 1]})
func (m *MapDocument) RichSet(input map[string]interface{}) (err error) {
	for k, v := range input {
		if !strings.HasPrefix(k, "$") {
			fmt.Println("set", k, v)
			continue
		}
		action := k[1:]
		switch action {
		case "rpush":
		case "set":
		case "incr":
		case "del":
		default:
		}
	}
	return
}

// doc_get(key, ["name", "setting.mute", "photos.$1"])
func (m *MapDocument) RichGet(fields ...string) (result map[string]interface{}) {
	return
}

func (m *MapDocument) getField(field string) (ptr *interface{}) {
	return
}
