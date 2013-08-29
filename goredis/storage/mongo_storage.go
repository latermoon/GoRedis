package storage

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

// MongoStorage
// 使用MongoDB来存储Redis数据
type MongoStorage struct {
	session *mgo.Session
	kvColl  *mgo.Collection
}

func NewMongoStorage() (mongo *MongoStorage) {
	mongo = &MongoStorage{}
	return
}

func (m *MongoStorage) Connect(url string) (err error) {
	m.session, err = mgo.Dial(url)
	if err == nil {
		m.kvColl = m.session.DB("goredis").C("kv")
		// 确保一个名为key的唯一索引
		idx := mgo.Index{Key: []string{"key"}, Unique: true}
		m.kvColl.EnsureIndex(idx)
	}
	return
}

func (m *MongoStorage) Close() {
	m.session.Close()
}

func (m *MongoStorage) Set(key string, value string) (err error) {
	_, err = m.kvColl.Upsert(bson.M{"key": key}, bson.M{"key": key, "val": value, "uptime": bson.Now()})
	return
}

func (m *MongoStorage) Get(key string) (value string, err error) {
	row := bson.M{}
	err = m.kvColl.Find(bson.M{"key": key}).One(&row)
	if err == nil {
		value = row["val"].(string)
	}
	return
}
