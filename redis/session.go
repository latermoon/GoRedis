package redis

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
)

// cmd, err := session.ReadCommand()
// session.WriteReply(reply)
// session.Write(reply.Bytes())
type Session struct {
	net.Conn
	rd    *bufio.Reader
	rlock sync.Mutex
}

func NewSession(conn net.Conn) *Session {
	return &Session{
		Conn: conn,
		rd:   bufio.NewReader(conn),
	}
}

func (s *Session) ReadCommand() (Command, error) {
	s.rlock.Lock()
	defer s.rlock.Unlock()
	// Read ( *<number of arguments> CR LF )
	if err := s.skipByte('*'); err != nil { // io.EOF
		return nil, err
	}
	// number of arguments
	argCount, err := s.readInt()
	if err != nil {
		return nil, err
	}
	args := make([][]byte, argCount)
	for i := 0; i < argCount; i++ {
		// Read ( $<number of bytes of argument 1> CR LF )
		if err := s.skipByte('$'); err != nil {
			return nil, err
		}

		var argSize int
		argSize, err = s.readInt()
		if err != nil {
			return nil, err
		}

		// Read ( <argument data> CR LF )
		args[i] = make([]byte, argSize)
		_, err = io.ReadFull(s, args[i])
		if err != nil {
			return nil, err
		}

		err = s.skipBytes([]byte{CR, LF})
		if err != nil {
			return nil, err
		}
	}
	return Command(args), nil
}

func (s *Session) Read(p []byte) (int, error) {
	return s.rd.Read(p)
}

func (s *Session) WriteReply(r Reply) (int, error) {
	return s.Write(r.Bytes())
}

func (s *Session) skipByte(c byte) (err error) {
	var tmp byte
	tmp, err = s.rd.ReadByte()
	if err != nil {
		return
	}
	if tmp != c {
		err = errors.New(fmt.Sprintf("Illegal Byte [%d] != [%d]", tmp, c))
	}
	return
}

func (s *Session) skipBytes(bs []byte) error {
	for _, c := range bs {
		if err := s.skipByte(c); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) readLine() (line []byte, err error) {
	line, err = s.rd.ReadSlice(LF)
	if err == bufio.ErrBufferFull {
		return nil, errors.New("line too long")
	}
	if err != nil {
		return
	}
	i := len(line) - 2
	if i < 0 || line[i] != CR {
		err = errors.New("bad line terminator:" + string(line))
	}
	return line[:i], nil
}

func (s *Session) readInt() (int, error) {
	if line, err := s.readLine(); err == nil {
		return strconv.Atoi(string(line))
	} else {
		return 0, err
	}
}
