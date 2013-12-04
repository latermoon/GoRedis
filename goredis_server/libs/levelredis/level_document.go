package levelredis

import (
	"github.com/latermoon/msgpackgo/codec"
	"sync"
)

type LevelDocument struct {
	redis     *LevelRedis
	key       string
	mu        sync.Mutex
	initMutex sync.Mutex
	inited    bool
	doc       *MapDocument
	mh        codec.MsgpackHandle
}

func NewLevelDocument(redis *LevelRedis, key string) (l *LevelDocument) {
	l = &LevelDocument{}
	l.redis = redis
	l.key = key
	return
}

func (l *LevelDocument) Size() int {
	return 0
}

// 初始化一次
func (l *LevelDocument) initOnce() {
	l.initMutex.Lock()
	defer l.initMutex.Unlock()
	if l.inited {
		return
	}

	in, _ := l.redis.db.Get(l.redis.ro, l.docKey())
	if in != nil {
		dec := codec.NewDecoderBytes(in, &l.mh)
		m := make(map[string]interface{})
		err := dec.Decode(&m)
		if err != nil {
			l.doc = NewMapDocument(nil)
		} else {
			l.doc = NewMapDocument(m)
		}
	} else {
		l.doc = NewMapDocument(nil)
	}
	l.inited = true
}

func (l *LevelDocument) docKey() []byte {
	return joinStringBytes(KEY_PREFIX, SEP_LEFT, l.key, SEP_RIGHT, DOC_SUFFIX)
}

func (l *LevelDocument) docValue() (out []byte) {
	enc := codec.NewEncoderBytes(&out, &l.mh)
	enc.Encode(l.doc.Map())
	return
}

func (l *LevelDocument) Set(m map[string]interface{}) (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.initOnce()

	err = l.doc.RichSet(m)
	if err == nil {
		err = l.redis.db.Put(l.redis.wo, l.docKey(), l.docValue())
	}
	return
}

func (l *LevelDocument) Get(fields ...string) (result map[string]interface{}) {
	l.initOnce()
	result = l.doc.RichGet(fields...)
	return
}
