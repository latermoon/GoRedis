package util

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// 简易conf文件读取
// conf文件格式参考redis.conf
/*
# goredis.conf
# comments
port 1602
bind 127.0.0.1

config := util.OpenConfig(filename)
host := config.StringForKey("host", "127.0.0.1")
port := config.IntForKey("port", 1602)
arr := config.StringArrayForKey("clusters")
*/
type ConfigReader struct {
	params map[string]string
}

// 打开配置文件
func OpenConfig(filename string) (cr *ConfigReader, err error) {
	cr = &ConfigReader{}
	cr.params = make(map[string]string)

	// 配置文件较少，一次读入内存
	var fileData []byte
	fileData, err = ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	// 转成字符串行拆分
	lines := strings.Split(string(fileData), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		keyVal := strings.Fields(line)
		if len(keyVal) < 2 {
			panic(fmt.Sprintf("Bad Config @%s \"line %d: %s\"", filename, i, line))
		}
		// fill params
		cr.params[keyVal[0]] = strings.Join(keyVal[1:], " ")
	}
	return
}

func (cr *ConfigReader) StringForKey(key string, defaultVal string) (val string) {
	var exists bool
	val, exists = cr.params[key]
	if !exists {
		val = defaultVal
	}
	return
}

func (cr *ConfigReader) StringArrayForKey(key string) (val []string) {
	str := cr.StringForKey(key, "")
	if len(str) > 0 {
		val = strings.Fields(str)
	} else {
		val = []string{}
	}
	return
}

func (cr *ConfigReader) IntForKey(key string, defaultVal int) (val int) {
	str := cr.StringForKey(key, "")
	if len(str) > 0 {
		var err error
		if val, err = strconv.Atoi(str); err != nil {
			panic(err)
		}
	} else {
		val = defaultVal
	}
	return
}

func (cr *ConfigReader) BoolForKey(key string, defaultVal bool) (val bool) {
	str := cr.StringForKey(key, "")
	if len(str) > 0 {
		// ParseBool无法识别yes/no，因此自己判断
		str = strings.ToLower(str)
		if str == "yes" {
			val = true
		} else if str == "no" {
			val = false
		} else {
			var err error
			if val, err = strconv.ParseBool(str); err != nil {
				panic(err)
			}
		}
	} else {
		val = defaultVal
	}
	return
}

func (cr *ConfigReader) Params() map[string]string {
	return cr.params
}
