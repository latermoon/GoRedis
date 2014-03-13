package goredis_server

import (
	. "GoRedis/goredis"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

func ParseInt64(b []byte) (i int64, err error) {
	i, err = strconv.ParseInt(string(b), 10, 64)
	return
}

func openfile(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
}

func directoryTotalSize(root string) (size int64) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return
}

func bytesInHuman(size int64) string {
	f := float64(size)
	if f > 1024*1024*1024*1024 {
		return fmt.Sprintf("%0.2fT", f/1024/1024/1024/1024)
	}
	if f > 1024*1024*1024 {
		return fmt.Sprintf("%0.2fG", f/1024/1024/1024)
	}
	if f > 1024*1024 {
		return fmt.Sprintf("%0.2fM", f/1024/1024)
	}
	if f > 1024 {
		return fmt.Sprintf("%0.2fK", f/1024)
	}
	return fmt.Sprintf("%dB", size)
}

// 将各种对象，转换为字符串形式，再转为[]byte数组
func formatByteSlice(v ...interface{}) (buf [][]byte) {
	buf = make([][]byte, 0, len(v))
	for i := 0; i < len(v); i++ {
		buf = append(buf, []byte(fmt.Sprint(v[i])))
	}
	return
}

func redisInfo(session *Session) (isgoredis bool, version string, err error) {
	cmdinfo := NewCommand([]byte("info"), []byte("server"))
	session.WriteCommand(cmdinfo)
	var reply *Reply
	reply, err = session.ReadReply()
	if err != nil {
		return
	}
	if reply.Value == nil {
		err = errors.New("reply nil")
		return
	}

	var info string
	switch reply.Value.(type) {
	case string:
		info = reply.Value.(string)
	case []byte:
		info = string(reply.Value.([]byte))
	default:
		info = reply.String()
	}

	// 切分info返回的数据，存放到map里
	kv := make(map[string]string)
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		line = strings.TrimPrefix(line, " ")
		if strings.HasPrefix(line, "#") {
			continue
		}
		pairs := strings.Split(line, ":")
		if len(pairs) != 2 {
			continue
		}
		// done
		kv[pairs[0]] = pairs[1]
	}

	_, isgoredis = kv["goredis_version"]
	if isgoredis {
		version = kv["goredis_version"]
	} else {
		version = kv["redis_version"]
	}

	return
}
