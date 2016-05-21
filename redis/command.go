package redis

import (
	"bytes"
	"encoding/json"
)

type Command [][]byte

func (c Command) Bytes() []byte {
	buf := bytes.Buffer{}
	buf.WriteByte('*')
	argCount := len(c)
	buf.WriteString(itoa(argCount)) //<number of arguments>
	buf.WriteString(CRLF)
	for i := 0; i < argCount; i++ {
		buf.WriteByte('$')
		buf.WriteString(itoa(len(c[i]))) //<number of bytes of argument i>
		buf.WriteString(CRLF)
		buf.Write(c[i]) //<argument data>
		buf.WriteString(CRLF)
	}
	return buf.Bytes()
}

func (c Command) Size() int {
	sum := 0
	for i := 0; i < len(c); i++ {
		sum += len(c[i])
	}
	return sum
}

func (c Command) String() string {
	arr := make([]string, len(c))
	for i := range c {
		arr[i] = string(c[i])
	}
	b, _ := json.Marshal(arr)
	return string(b)
}
