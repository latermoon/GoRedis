package goredis_server

// TODO 本类未完成

/*
实现类似于mongo的document base指令，
对一个key提供docuemnt存储，以及原子操作

user:100422:doc = {
	name: "latermoon", // string
	sex: 1 // int
	regtime: ISOTime("2013-12-3 22:12:48"), // time
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

// Replace/Insert
doc_replace(key, {"name":"latermoon", "sex":1, "photos":[...], ...})
// Update/Insert
doc_set(key, {Action:{Field1:Value1, ...}, ...})
doc_set(key, {"name":"latermoon"})
doc_set(key, {"$rpush":["photos", "d.jpg", "e.jpg"]}})
doc_set(key, {"$incr":["version", 1]})
doc_set(key, {"setting.mute":{"start":23, "end":8}})
doc_set(key, {"setting.mute.start":23, "setting.mute.end":8})
doc_set(key, {"$del":["name", "photos.$1", "setting.mute.start"])
// Get All
doc_get(key)
doc_get(key, "name,sex,photos,setting.mute,version")

*/

import (
	. "../goredis"
)

func (server *GoRedisServer) OnDOC_SET(cmd *Command) (reply *Reply) {
	return
}

func (server *GoRedisServer) OnDOC_GET(cmd *Command) (reply *Reply) {
	return
}
