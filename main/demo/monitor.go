package main

import (
	. "../../goredis"
	qp "../../goredis_server/libs/queueprocess"
	"bufio"
	"bytes"
	"fmt"
	"github.com/latermoon/redigo/redis"
	"net"
	"runtime"
	"strings"
	"time"
)

var pool *redis.Pool

var srchost = "redis-event-a001:8401"

// var dsthost = "goredis-nearby-a001:18402"

var dsthost = "localhost:1602"

func main() {
	runtime.GOMAXPROCS(8)
	conn, err := net.Dial("tcp", srchost)
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

	queue := qp.NewQueueProcess(100, writeCommand)

	go func() {
		ticker := time.NewTicker(time.Millisecond * 1000)
		for _ = range ticker.C {
			fmt.Println("queue:", queue.QueueLen())
			// fmt.Println("queue:", queue.Len(), "active:", pool.ActiveCount())
		}
	}()

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}

		cmd := &Command{}
		cmd.Args = make([][]byte, 0, 5)
		splitMonitorLine(line, cmd)

		if len(cmd.Args) >= 2 {
			// fmt.Println(SumOfBytesChars(cmd.Args[1]), cmd)
			queue.Process(SumOfBytesChars(cmd.Args[1]), cmd)
		}
	}
}

func writeCommand(t qp.Task) {
	cmd := t.(*Command)
	cmdname := cmd.StringAtIndex(0)
	// if strings.HasPrefix(cmdname, "L") {
	// 	return
	// }

	if strings.HasPrefix(cmdname, "111L") {
		return
	}

	if cmd.StringAtIndex(1) == "user:update:timestamp" {
		return
	}

	objs := make([]interface{}, 0, len(cmd.Args)-1)
	for _, arg := range cmd.Args[1:] {
		objs = append(objs, arg)
	}

	rd := pool.Get()
	_, err := rd.Do(cmd.Name(), objs...)
	rd.Close()

	// fmt.Println(len(cmd.Args), cmd)
	if err == nil {
		// fmt.Println("+reply:", reply)
		// printReply(reply)
	} else {
		fmt.Println("-err:", err)
		panic(err)
	}
}

func SumOfBytesChars(bs []byte) (n int) {
	count := len(bs)
	for i := 0; i < count; i++ {
		n += int(bs[i])
	}
	return
}

func printReply(reply interface{}) {
	fmt.Print("+reply: ")
	switch reply.(type) {
	case []interface{}:
		arr := reply.([]interface{})
		fmt.Print("[")
		for _, e := range arr {
			fmt.Print(string(e.([]byte)), " ")
		}
		fmt.Println("]")
	case int:
		fmt.Println(reply)
	case []byte:
		fmt.Println(string(reply.([]byte)))
	default:
		fmt.Println(reply)
	}
}

func init() {
	pool = &redis.Pool{
		MaxIdle:     2000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", dsthost)
			return c, err
		},
	}
}

// 将monitor里输出的 +1386347668.732167 [0 10.80.101.169:8400] "ZADD" "user:update:timestamp" "1.386347668E9" "40530990"
// 转换为Command对象
func splitMonitorLine(line []byte, cmd *Command) {
	firstQuote := bytes.Index(line, []byte("\""))    // 第一个引号
	lastQuote := bytes.LastIndex(line, []byte("\"")) // 最后一个引号，主要是为了去掉最后的换行符

	cmdline := line[firstQuote : lastQuote+1]
	reader := bytes.NewReader(cmdline)

	var argidx int    // 当前操作的Args元素
	quoteMatched := 0 // 引号出现次数
	for {
		c, err := reader.ReadByte()
		if err != nil {
			break
		}
		switch c {
		case '"':
			quoteMatched++
			// 遇到第一个引号，创建内存空间
			if quoteMatched == 1 {
				cmd.Args = append(cmd.Args, []byte{})
				argidx = len(cmd.Args) - 1
			} else if quoteMatched == 2 {
				// 遇到另一个引号，标记关闭
				quoteMatched = 0
			}
		case ' ':
			//  引号内的空格属于内容
			if quoteMatched == 1 {
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
