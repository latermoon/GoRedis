package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var (
	WrongKindError   = errors.New("wrong kind error")
	BadArgumentCount = errors.New("bad argument count")
	BadArgumentType  = errors.New("bad argument type")
	msitype          = reflect.TypeOf(make(map[string]interface{}))
)

const (
	dot = "."
)

func main() {
	doc := NewMapDocument(nil)

	in := map[string]interface{}{"name": "latermoon"}
	err := doc.Set(in)

	in = map[string]interface{}{"tianya": map[string]interface{}{"u": "latermoon",
		"p":    "1234",
		"more": map[string]interface{}{"a": 1, "b": 2, "c": 3},
		"arr":  []interface{}{1, 2, 3}}}
	err = doc.Set(in)
	fmt.Println(doc, err)

	in = map[string]interface{}{"tianya.more.a": 4, "$del": []interface{}{"tianya.arr"}}
	err = doc.Set(in)
	fmt.Println(doc, err)

	in = map[string]interface{}{"$rpush": map[string]interface{}{"tianya.more": []interface{}{100}}}
	err = doc.Set(in)
	fmt.Println(doc, err)
}

// 提供面向document操作的map
// doc := New()
// doc.Set(jsonObj)
// doc.Get(fields)
type MapDocument struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewMapDocument(data map[string]interface{}) (m *MapDocument) {
	m = &MapDocument{}
	if m.data = data; m.data == nil {
		m.data = make(map[string]interface{})
	}
	return
}

// doc_set(key, {"name":"latermoon", "$rpush":["photos", "c.jpg", "d.jpg"], "$incr":["version", 1]})
func (m *MapDocument) Set(in map[string]interface{}) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	defer func() {
		if v := recover(); v != nil {
			if e, ok := v.(error); ok {
				err = e
			} else {
				err = errors.New(fmt.Sprint(v))
			}
		}
	}()
	for k, v := range in {
		if !strings.HasPrefix(k, "$") {
			parent, key, _, _ := m.findElement(k, true)
			parent[key] = v
			continue
		}
		action := k[1:]
		switch action {
		case "set":
			argmap := v.(map[string]interface{})
			for field, value := range argmap {
				parent, key, _, _ := m.findElement(field, true)
				parent[key] = value
			}
		case "rpush":
			argmap := v.(map[string]interface{})
			for field, value := range argmap {
				parent, key, _, _ := m.findElement(field, true)
				m.doRpush(parent, key, value.([]interface{}))
			}
		case "incr":
			argmap := v.(map[string]interface{})
			for field, value := range argmap {
				parent, key, _, _ := m.findElement(field, true)
				err = m.doIncr(parent, key, value)
			}
		case "del":
			arglist := v.([]interface{})
			for _, field := range arglist {
				parent, key, _, exist := m.findElement(field.(string), false)
				if exist {
					delete(parent, key)
				}
			}
		default:
		}
	}
	return
}

// doc_get(key, ["name", "setting.mute", "photos.$1"])
func (m *MapDocument) Get(fields ...string) (out map[string]interface{}) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out = make(map[string]interface{})
	if len(fields) == 0 || (len(fields) == 1 && fields[0] == "") {
		for k, v := range m.data {
			out[k] = v
		}
		return
	}

	for _, field := range fields {
		dst := out
		src := m.data
		// 逐个字段扫描copy
		pairs := strings.Split(field, dot)
		count := len(pairs)
		for i := 0; i < count; i++ {
			curkey := pairs[i]
			obj, ok := src[curkey]
			if !ok {
				break
			}
			if i > 0 && i == count-1 {
				dst[curkey] = obj
				continue
			}
			if reflect.TypeOf(obj) != msitype {
				// 基础类型
				dst[curkey] = obj
				continue
			}
			if dst[curkey] == nil {
				dst[curkey] = make(map[string]interface{})
			}
			src = src[curkey].(map[string]interface{})
			dst = dst[curkey].(map[string]interface{})
		}
	}
	return
}

/**
 * 根据field路径查找元素
 * @param field 多级的field使用"."分隔
 * @return parent[key] == obj，其中 parent 目标元素父对象，必定是map[string]interface{}，key 目标元素key，obj，目标元素
 */
func (m *MapDocument) findElement(field string, createIfMissing bool) (parent map[string]interface{}, key string, obj interface{}, exist bool) {
	pairs := strings.Split(field, dot)
	parent = m.data
	for i := 0; i < len(pairs)-1; i++ {
		curkey := pairs[i]
		var ok bool
		_, ok = parent[curkey]
		// 初始化或覆盖
		if !ok || reflect.TypeOf(parent[curkey]) != msitype {
			if createIfMissing {
				parent[curkey] = make(map[string]interface{})
			} else {
				exist = false
				return
			}
		}
		parent = parent[curkey].(map[string]interface{})
	}
	exist = true
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
		oldint, e1 := toInt(obj)
		incrint, e2 := toInt(value)
		if e1 != nil || e2 != nil {
			return BadArgumentType
		}
		parent[key] = oldint + incrint
	}
	return
}

func toInt(obj interface{}) (n int, err error) {
	switch obj.(type) {
	case int:
		n = obj.(int)
	case float64:
		n = int(obj.(float64))
	default:
		err = BadArgumentType
	}
	return
}

func (m *MapDocument) String() string {
	b, _ := json.Marshal(m.data)
	return string(b)
}

func (m *MapDocument) Map() map[string]interface{} {
	return m.data
}
