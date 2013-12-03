package levelredis

import (
// "github.com/latermoon/levigo"
)

type LevelConfig struct {
	redis  *LevelRedis
	prefix string
}

func NewLevelConfig(redis *LevelRedis, prefix string) (l *LevelConfig) {
	l = &LevelConfig{}
	l.redis = redis
	l.prefix = prefix
	return
}

func (l *LevelConfig) GetString(key string) (value string) {
	return
}

func (l *LevelConfig) SetString(key string, value string) {

}

func (l *LevelConfig) GetMap(key string) (m map[string]interface{}) {
	return
}

func (l *LevelConfig) SetMap(key string, m map[string]interface{}) {

}

func (l *LevelConfig) GetArray(key string) (arr []interface{}) {
	return
}

func (l *LevelConfig) SetArray(key string, arr []interface{}) {

}
