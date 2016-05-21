package redis

import (
	"errors"
	"fmt"
	"log"
	"net"
)

// handler = &server.GoRedisServer{}
// lis, err := net.Listen("tcp", "localhost:6380")
// if err != nil {
//     panic(err)
// }
// redis.Serve(lis, handler)

func Register(handler ServerHandler) { DefaultServer.Register(handler) }

func Serve(lis net.Listener) error { return DefaultServer.Serve(lis) }

var DefaultServer = NewServer()

type ServerHandler interface {
	SessionOpened(*Session)
	SessoinClosed(*Session, error)
	RecvCommand(*Session, Command)
}

type Server struct {
	handler ServerHandler
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Register(handler ServerHandler) {
	s.handler = handler
}

func (s *Server) Serve(lis net.Listener) error {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("redis.Serve: accept:", err.Error())
			return err
		}
		go s.ServeSession(NewSession(conn))
	}

	return nil
}

func (s *Server) ServeSession(session *Session) {
	defer func() {
		session.Close()
		if v := recover(); v != nil {
			err, ok := v.(error)
			if !ok {
				err = errors.New(fmt.Sprint(v))
			}
			s.handler.SessoinClosed(session, err)
		}
		s.handler.SessoinClosed(session, nil)
	}()

	s.handler.SessionOpened(session)

	for {
		cmd, err := session.ReadCommand()
		if err != nil {
			break
		}
		s.handler.RecvCommand(session, cmd)
	}
}
