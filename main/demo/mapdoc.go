package main

import (
	. "../../goredis_server/libs/levelredis"
	"fmt"
)

func main() {
	doc := NewMapDocument(nil)
	//
	m := make(map[string]interface{})
	m["$set"] = map[string]interface{}{
		"name":         "latermoon",
		"is_vip":       false,
		"setting.mute": map[string]interface{}{"start": 23, "end": 8},
	}
	m["setting.mute.status"] = false
	m["$rpush"] = map[string]interface{}{"profile.photos": []interface{}{"A.jpg", "B.jpg"}}
	m["$incr"] = map[string]interface{}{"profile.version": 10, "count": 3}
	doc.RichSet(m)

	jsonString := doc.String()
	fmt.Println(jsonString)

	// 格式化校验
	// m2 := make(map[string]interface{})
	// json.Unmarshal([]byte(jsonString), &m2)
	// fmt.Println(m2)
}
