package levelredis

// TODO 本类未完成

import (
	"encoding/json"
	"errors"
	// "fmt"
	"reflect"
	"strings"
)

var (
	WrongKindError   = errors.New("wrong kind error")
	BadArgumentCount = errors.New("bad argument count")
	MapInterfaceType = reflect.TypeOf(make(map[string]interface{}))
)

const (
	dot = "."
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
			if strings.Index(k, dot) < 0 {
				m.data[k] = v
			} else {
				// 对于多层级的元素，需要分拆查找
				parent, key, _ := m.findElement(k)
				parent[key] = v
			}
			continue
		}
		action := k[1:]
		switch action {
		case "rpush":
			arglist := v.([]interface{}) // [0] 表示 field，[1:] 是要添加的元素
			parent, key, _ := m.findElement(arglist[0].(string))
			m.doRpush(parent, key, arglist[1:])
		case "incr":
			arglist := v.([]interface{}) // [0] 表示 field，[1] 是要incr的数字
			parent, key, _ := m.findElement(arglist[0].(string))
			m.doIncr(parent, key, arglist[1])
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

/**
 * 根据field路径查找元素
 * @param field 多级的field使用"."分隔
 * @return parent[key] == obj，其中 parent 目标元素父对象，必定是map[string]interface{}，key 目标元素key，obj，目标元素
 */
func (m *MapDocument) findElement(field string) (parent map[string]interface{}, key string, obj interface{}) {
	pairs := strings.Split(field, dot)
	parent = m.data
	for i := 0; i < len(pairs)-1; i++ {
		curkey := pairs[i]
		var ok bool
		_, ok = parent[curkey]
		// 初始化或覆盖
		if !ok || reflect.TypeOf(parent[curkey]) != MapInterfaceType {
			parent[curkey] = make(map[string]interface{})
		}
		parent = parent[curkey].(map[string]interface{})
	}
	key = pairs[len(pairs)-1]
	obj = parent[key]
	return
}

func (m *MapDocument) doRpush(parent map[string]interface{}, key string, elems []interface{}) (err error) {
	obj := parent[key]
	if obj != nil {
		for i := 0; i < len(elems); i++ {
			parent[key] = append(parent[key].([]interface{}), elems[i])
		}
	} else {
		parent[key] = elems
	}
	return
}

func (m *MapDocument) doIncr(parent map[string]interface{}, key string, value interface{}) (err error) {
	obj := parent[key]
	if obj == nil {
		parent[key] = value
	} else {
		parent[key] = obj.(int) + value.(int)
	}
	// fmt.Println(obj, reflect.TypeOf(obj), value, reflect.TypeOf(value))
	return
}

func parseFloat64(obj interface{}) {

}

func (m *MapDocument) String() string {
	b, _ := json.Marshal(m.data)
	return string(b)
}
