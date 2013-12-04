package main

import (
	. "../../goredis_server/libs/levelredis"
	"encoding/json"
	"fmt"
)

func main() {
	doc := NewMapDocument(nil)
	//
	m := make(map[string]interface{})
	m["name"] = "latermoon"
	m["setting.mute.start"] = 10
	m["setting.mute"] = map[string]interface{}{"start": 23, "end": 8}
	m["$rpush"] = []interface{}{"profile.photos", "A.jpg", "B.jpg"}
	m["$incr"] = []interface{}{"profile.version", 10}
	doc.RichSet(m)

	jsonString := doc.String()
	fmt.Println(jsonString)

	m2 := make(map[string]interface{})
	json.Unmarshal([]byte(jsonString), &m2)
	fmt.Println(m2)

	doc2 := NewMapDocument(m2)
	fmt.Println(doc2)
}
