package goredis_server

import (
	. "GoRedis/goredis"
	"GoRedis/libs/levelredis"
	"bufio"
	"io"
	"strconv"
	"sync"
)

type AOFWriter struct {
	io.Writer
	fd     *bufio.Writer
	mu     sync.Mutex
	closed bool
}

func NewAOFWriter(fd *bufio.Writer) (a *AOFWriter) {
	a = &AOFWriter{
		fd: fd,
	}
	return
}

func (a *AOFWriter) Write(p []byte) (n int, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.fd.Write(p)
}

func (a *AOFWriter) Flush() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.fd.Flush()
}

func (a *AOFWriter) AppendString(key, value []byte) {
	cmd := NewCommand([]byte("SET"), key, value)
	a.Write(cmd.Bytes())
}

func (a *AOFWriter) AppendHash(h *levelredis.LevelHash) {
	h.Enumerate(func(i int, key, value []byte, quit *bool) {

	})
}

func (a *AOFWriter) AppendSet(h *levelredis.LevelHash) {
	h.Enumerate(func(i int, key, value []byte, quit *bool) {

	})
}

func (a *AOFWriter) AppendList(h *levelredis.LevelList) {
	h.Enumerate(func(i int, value []byte, quit *bool) {

	})
}

func (a *AOFWriter) AppendZSet(z *levelredis.LevelZSet) {
	var buf [][]byte
	bufsize := 200
	z.Enumerate(func(i int, member, score []byte, quit *bool) {
		if buf == nil {
			buf = make([][]byte, 0, bufsize+4)
			buf = append(buf, []byte("ZADD"), []byte(z.Key()))
		}
		buf = append(buf, []byte(strconv.FormatInt(BytesToInt64(score), 10)), member)
		if len(buf) > bufsize {
			cmd := NewCommand(buf...)
			a.Write(cmd.Bytes())
			buf = nil
		}
	})
	if len(buf) > 0 {
		cmd := NewCommand(buf...)
		a.Write(cmd.Bytes())
	}
	a.fd.Flush()
	return
}

func (a *AOFWriter) AppendDoc(d *levelredis.LevelDoc) {
	return
}

func (a *AOFWriter) IsClosed() bool {
	return a.closed
}

func (a *AOFWriter) Close() {
	a.closed = true
}
