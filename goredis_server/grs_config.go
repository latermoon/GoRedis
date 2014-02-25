package goredis_server

import (
	"GoRedis/libs/levelredis"
	"strconv"
	"strings"
)

// 配置读写
type Config struct {
	redis  *levelredis.LevelRedis
	prefix string
}

func NewConfig(redis *levelredis.LevelRedis, prefix string) (c *Config) {
	c = &Config{}
	c.redis = redis
	c.prefix = prefix
	return
}

func (c *Config) fieldKey(key string) []byte {
	return []byte(c.prefix + key)
}

func (c *Config) Keys() (keys []string) {
	keys = make([]string, 0, 10)
	c.redis.PrefixEnumerate([]byte(c.prefix), levelredis.IterForward, func(i int, key, value []byte, quit *bool) {
		s := strings.TrimPrefix(string(key), c.prefix)
		keys = append(keys, s)
	})
	return
}

func (c *Config) Set(key string, value []byte) {
	c.redis.RawSet(c.fieldKey(key), value)
}

func (c *Config) Get(key string) []byte {
	return c.redis.RawGet(c.fieldKey(key))
}

func (c *Config) StringForKey(key string) string {
	v := c.Get(key)
	return string(v)
}

func (c *Config) IntForKey(key string, defval int64) (n int64) {
	v := c.Get(key)
	if v == nil || len(v) == 0 {
		n = defval
	} else {
		var err error
		n, err = strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			n = defval
		}
	}
	return
}
