package levelredis

import (
	"GoRedis/libs/msgpackgo/codec"
	"reflect"
	"sync"
)

var docmh *codec.MsgpackHandle

func init() {
	docmh = &codec.MsgpackHandle{}
	docmh.RawToString = true
	docmh.MapType = reflect.TypeOf(make(map[string]interface{}))
}

type LevelDoc struct {
	LevelElem
	redis *LevelRedis
	key   string
	mu    sync.RWMutex
	doc   *MapDoc
}

func NewLevelDoc(redis *LevelRedis, key string) (l *LevelDoc) {
	l = &LevelDoc{
		redis: redis,
		key:   key,
	}
	l.initOnce()
	return
}

func (l *LevelDoc) Key() string {
	return l.key
}

func (l *LevelDoc) Size() int {
	return 1
}

// 初始化一次
func (l *LevelDoc) initOnce() {
	in, err := l.redis.RawGet(l.docKey())
	if err != nil || in == nil {
		l.doc = NewMapDoc(nil)
		return
	}

	dec := codec.NewDecoderBytes(in, docmh)
	m := make(map[string]interface{})
	if err := dec.Decode(&m); err == nil {
		l.doc = NewMapDoc(m)
	} else {
		l.doc = NewMapDoc(nil)
	}
}

func (l *LevelDoc) docKey() []byte {
	return joinStringBytes(KEY_PREFIX, SEP_LEFT, l.key, SEP_RIGHT, DOC_SUFFIX)
}

func (l *LevelDoc) docValue() (out []byte) {
	enc := codec.NewEncoderBytes(&out, docmh)
	enc.Encode(l.doc.Map())
	return
}

func (l *LevelDoc) Set(m map[string]interface{}) (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	err = l.doc.Set(m)
	if err == nil {
		err = l.redis.RawSet(l.docKey(), l.docValue())
	}
	return
}

func (l *LevelDoc) Get(fields ...string) (result map[string]interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result = l.doc.Get(fields...)
	return
}

func (l *LevelDoc) Type() string {
	return DOC_SUFFIX
}

func (l *LevelDoc) Drop() (ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	in, _ := l.redis.RawGet(l.docKey())
	if in != nil {
		l.redis.RawDel(l.docKey())
	}
	l.doc = nil
	ok = true
	return
}
