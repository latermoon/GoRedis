package shardredis

import (
	. "GoRedis/goredis"
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
			args, err := splitMonitorLine(line)
			if err != nil {
				fmt.Println(linestr)
				fmt.Println(err)
				continue
			}
			cmd := NewCommand(args...)
			rightSep := bytes.Index(line, []byte("]"))
			prefix := line[:rightSep+1]
			m.DidRecvCommand(cmd, string(prefix))
		}
	}
	return
}

func ConvertHexString(line []byte) (result []byte, err error) {
	buf := &bytes.Buffer{}
	for i, count := 0, len(line); i < count; i++ {
		b := line[i]
		if b == '\\' {
			if i+1 < count && line[i+1] == 'x' {
				if i+2+2 >= count {
					err = errors.New("invalid character \\x")
					return
				}
				b34hex := line[i+2 : i+2+2]
				b34 := make([]byte, hex.DecodedLen(len(b34hex)))
				_, err = hex.Decode(b34, b34hex)
				if err != nil {
					return
				}
				buf.Write(b34)
				i += 3 // skip \xff
			} else {
				buf.WriteByte('\\')
			}
		} else {
			buf.WriteByte(b)
		}
	}
	result = buf.Bytes()
	return
}

// 将monitor里输出的 +1386347668.732167 [0 10.80.101.169:8400] "ZADD" "user:update:timestamp" "1.386347668E9" "40530990"
// 转换为Command对象
func splitMonitorLine(line []byte) (args [][]byte, err error) {
	line = line[bytes.Index(line, []byte("] "))+2:]
	line, err = ConvertHexString(line)
	if err != nil {
		return
	}
	line = bytes.Replace(line, []byte("\" \""), []byte("\", \""), -1)
	line = bytes.Join([][]byte{[]byte("["), line, []byte("]")}, []byte(""))
	var obj []interface{}
	err = json.Unmarshal(line, &obj)
	if err != nil {
		return
	}
	args = make([][]byte, len(obj))
	for i := 0; i < len(obj); i++ {
		args[i] = []byte(obj[i].(string))
	}
	return
}
