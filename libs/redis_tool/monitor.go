package redis_tool

import (
	. "GoRedis/goredis"
	"bufio"
	"bytes"
	"net"
)

/**

r := NewMonitorReader(host)
r.Filter = func() bool {
	return true
}
err := r.Connect()
cmd, err := r.ReadCommand()

*/
type MonitorReader struct {
	host               string
	reader             *bufio.Reader
	DidRecvMonitorLine func(line string)
	DidRecvCommand     func(cmd *Command, prefix string)
}

func NewMonitorReader(host string) (m *MonitorReader) {
	m = &MonitorReader{}
	m.host = host
	return
}

func (m *MonitorReader) Connect() (err error) {
	var conn net.Conn
	conn, err = net.Dial("tcp", m.host)
	if err != nil {
		return
	}
	m.reader = bufio.NewReader(conn)
	conn.Write(NewCommand([]byte("MONITOR")).Bytes())

	// 首先会返回一个"OK"
	var line []byte
	line, err = m.reader.ReadBytes('\n')
	if err != nil {
		return
	}
	linestr := string(line)
	if linestr != "+OK\r\n" {
		panic("connect error " + linestr)
	}

	// recv loop

	for {
		line, err = m.reader.ReadBytes('\n')
		if err != nil {
			break
		}
		line = bytes.TrimSuffix(line, []byte("\r\n"))

		linestr := string(line)
		if m.DidRecvMonitorLine != nil {
			m.DidRecvMonitorLine(linestr)
		}

		if m.DidRecvCommand != nil {
			args, _ := splitMonitorLine(line)
			cmd := NewCommand(args...)
			rightSep := bytes.Index(line, []byte("]"))
			prefix := line[:rightSep+1]
			m.DidRecvCommand(cmd, string(prefix))
		}
	}
	return
}

// 将monitor里输出的 +1386347668.732167 [0 10.80.101.169:8400] "ZADD" "user:update:timestamp" "1.386347668E9" "40530990"
// 转换为Command对象
func splitMonitorLine(line []byte) (args [][]byte, err error) {
	firstQuote := bytes.Index(line, []byte("\""))    // 第一个引号
	lastQuote := bytes.LastIndex(line, []byte("\"")) // 最后一个引号，主要是为了去掉最后的换行符

	cmdline := line[firstQuote : lastQuote+1]
	reader := bytes.NewReader(cmdline)

	var argidx int // 当前操作的Args元素
	args = make([][]byte, 0, 5)
	quoteMatched := 0 // 引号出现次数
	for {
		var c byte
		c, err = reader.ReadByte()
		if err != nil {
			break
		}
		switch c {
		case '"':
			quoteMatched++
			// 遇到第一个引号，创建内存空间
			if quoteMatched == 1 {
				args = append(args, []byte{})
				argidx = len(args) - 1
			} else if quoteMatched == 2 {
				// 遇到另一个引号，标记关闭
				quoteMatched = 0
			}
		case ' ':
			//  引号内的空格属于内容
			if quoteMatched == 1 {
				args[argidx] = append(args[argidx], c)
			}
		case '\\':
			// 转义字符，添加下一个字符
			c, err = reader.ReadByte()
			if err != nil {
				break
			}
			args[argidx] = append(args[argidx], c)
		default:
			args[argidx] = append(args[argidx], c)
		}
	}
	return
}
