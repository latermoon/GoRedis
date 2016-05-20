package goredis_server

/*
实现类似于mongo的document base指令，
对一个key提供docuemnt存储，以及原子操作

user:100422:profile = {
	name: "latermoon", // string
	sex: 1 // int
	photos: ["a.jpg", "b.jpg", "c.jpg"], // array<string>
	setting: { // hash
		mute: {
			start: 23,
			end: 8
		}
	},
	is_vip: true, // bool
	version: 172 // int
}

// Update/Insert
doc_set(key, {"name":"latermoon"})
doc_set(key, {"$rpush":["photos", "d.jpg", "e.jpg"]}})
doc_set(key, {"$incr":["version", 1]})
doc_set(key, {"setting.mute":{"start":23, "end":8}})
doc_set(key, {"setting.mute.start":23, "setting.mute.end":8})
doc_set(key, {"$del":["name", "setting.mute.start"])

doc_set(key, {"$set":{"name":"latermoon", "sex":"M"}, "$inc":{"profile.version":1}})

// Get All
doc_get(key)
doc_get(key, "name,sex,photos,setting.mute,version")

*/

import (
	. "GoRedis/goredis"
	"encoding/json"
	"strings"
)

/*
doc_set hi '{"name":"latermoon", "sex":"M", "version":10, "setting":{"start":23, "end":8}}'
doc_set hi '{"$inc":{"version":1}}'
doc_set hi '{"$del":["version", "setting.start"]}'
*/
func (server *GoRedisServer) OnDOC_SET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	// 传入的json字节
	jsonbytes, err := cmd.ArgAtIndex(2)
	if err != nil {
		return ErrorReply(err)
	}
	// 反序列化为map
	jsonObj := make(map[string]interface{})
	err = json.Unmarshal(jsonbytes, &jsonObj)
	if err != nil {
		return ErrorReply(err)
	}
	// 调用LevelDocument更新数据
	doc := server.levelRedis.GetDoc(key)
	err = doc.Set(jsonObj)
	if err != nil {
		return ErrorReply(err)
	}
	reply = StatusReply("OK")
	return
}

func (server *GoRedisServer) OnDOC_GET(cmd *Command) (reply *Reply) {
	key := cmd.StringAtIndex(1)
	fields := strings.Split(cmd.StringAtIndex(2), ",")
	doc := server.levelRedis.GetDoc(key)
	result := doc.Get(fields...)
	if result == nil {
		return BulkReply(nil)
	}
	data, err := json.Marshal(result)
	if err != nil {
		return ErrorReply(err)
	}
	reply = BulkReply(data)
	return
}
