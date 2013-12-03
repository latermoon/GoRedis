package levelredisgo

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type LevelConfig struct {
	db     *leveldb.DB
	ro     *opt.ReadOptions
	wo     *opt.WriteOptions
	prefix string
}

func NewLevelConfig(db *leveldb.DB, prefix string) (l *LevelConfig) {
	l = &LevelConfig{}
	l.db = db
	l.ro = &opt.ReadOptions{}
	l.wo = &opt.WriteOptions{}
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
