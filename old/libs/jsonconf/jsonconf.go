package jsonconf

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type JsonConfig struct {
	m map[string]interface{}
}

func New() (j *JsonConfig) {
	j = &JsonConfig{}
	return
}

func (j *JsonConfig) Load(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	j.m = make(map[string]interface{})
	err = json.Unmarshal(b, &j.m)
	return err
}

func (j *JsonConfig) StringForKey(key string, defaultValue string) (s string) {
	v, ok := j.m[key]
	if ok {
		s = v.(string)
	} else {
		s = defaultValue
	}
	return
}

func (j *JsonConfig) IntForKey(key string, defaultValue int64) (n int64) {
	v, ok := j.m[key]
	if ok {
		n = int64(v.(float64))
	} else {
		n = defaultValue
	}
	return
}

func (j *JsonConfig) FloatForKey(key string, defaultValue float64) (n float64) {
	v, ok := j.m[key]
	if ok {
		n = v.(float64)
	} else {
		n = defaultValue
	}
	return
}
