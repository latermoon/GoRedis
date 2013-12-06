package main

import (
	. "../../goredis"
	"bufio"
	"bytes"
	"fmt"
	"github.com/latermoon/redigo/redis"
	"net"
	"time"
)

var pool *redis.Pool

func main() {
	conn, err := net.Dial("tcp", "redis-event-a001:8400")
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(conn)
	fmt.Println("MONITOR...")
	conn.Write([]byte("MONITOR\r\n"))
	line, err := reader.ReadBytes('\n')
	if err != nil {
		panic(err)
	} else {
		fmt.Println(string(line))
	}

	cmd := &Command{}
	cmd.Args = make([][]byte, 0)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}
		cmd.Args = cmd.Args[0:0]
		splitMonitorLine(line, cmd)
		rd := pool.Get()
		defer rd.Close()
		objs := make([]interface{}, 0, len(cmd.Args)-1)
		for _, arg := range cmd.Args[1:] {
			objs = append(objs, arg)
		}
		reply, err := rd.Do(cmd.Name(), objs)
		// go func() {

		// }()
		fmt.Println(cmd)
		fmt.Println(reply)
	}
}

func init() {
	pool = &redis.Pool{
		MaxIdle:     200,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			// c, err := redis.Dial("tcp", "goredis-nearby-a001:18400")
			c, err := redis.Dial("tcp", "localhost:1602")
			return c, err
		},
	}
}

// 将monitor里输出的 +1386347668.732167 [0 10.80.101.169:8400] "ZADD" "user:update:timestamp" "1.386347668E9" "40530990"
// 转换为Command对象
func splitMonitorLine(line []byte, cmd *Command) {
	firstQuote := bytes.Index(line, []byte("\""))    // 第一个引号
	lastQuote := bytes.LastIndex(line, []byte("\"")) // 最后一个引号，主要是为了去掉最后的换行符

	cmdline := line[firstQuote:lastQuote]
	reader := bytes.NewReader(cmdline)

	var argidx int        // 当前操作的Args元素
	quoteMatched := false // 引号成双匹配
	for {
		c, err := reader.ReadByte()
		if err != nil {
			break
		}
		switch c {
		case '"':
			// 遇到第一个引号，创建内存空间
			if !quoteMatched {
				cmd.Args = append(cmd.Args, []byte{})
				argidx = len(cmd.Args) - 1
			} else {
				// 遇到另一个引号，标记关闭
				quoteMatched = true
			}
		case ' ':
			//  引号内的空格属于内容
			if !quoteMatched {
				cmd.Args[argidx] = append(cmd.Args[argidx], c)
			}
		case '\\':
			// 转义字符，添加下一个字符
			c, err = reader.ReadByte()
			if err != nil {
				break
			}
			cmd.Args[argidx] = append(cmd.Args[argidx], c)
		default:
			cmd.Args[argidx] = append(cmd.Args[argidx], c)
		}
	}
	return
}
