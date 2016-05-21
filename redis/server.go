package redis

import (
	"io"
	"log"
	"net"
	"reflect"
	"strings"
)

// redis.Register(&server.GoRedisServer{})
// lis, err := net.Listen("tcp", "localhost:6380")
// if err != nil {
//     panic(err)
// }
// redis.Serve(lis)

func Register(recv interface{}) { DefaultServer.Register(recv) }

func Serve(lis net.Listener) { DefaultServer.Serve(lis) }

var DefaultServer = NewServer()

type ReplyWriter interface {
	WriteReply(Reply) (int, error)
}

type HandlerFunc func(ReplyWriter, Command)

func (f HandlerFunc) Serve(r ReplyWriter, c Command) {
	f(r, c)
}

type Server struct {
	handlers map[string]HandlerFunc
}

func NewServer() *Server {
	s := &Server{}
	s.handlers = make(map[string]HandlerFunc)
	return s
}

func (s *Server) Register(recv interface{}) {
	objval := reflect.ValueOf(recv)
	objtyp := reflect.TypeOf(recv)
	// log.Println("reflect", objtyp, objval, objval.NumMethod())
	for i := 0; i < objtyp.NumMethod(); i++ {
		name := objtyp.Method(i).Name
		if len(name) > 2 && strings.HasPrefix(name, "On") {
			s.registerHandler(name[2:], objval.Method(i))
		}
	}
}

func (s *Server) registerHandler(name string, method reflect.Value) {
	name = strings.ToUpper(name)
	s.handlers[name] = HandlerFunc(func(r ReplyWriter, c Command) {
		in := []reflect.Value{reflect.ValueOf(r), reflect.ValueOf(c)}
		method.Call(in)
	})
}

func (s *Server) Serve(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("redis.Serve: accept:", err.Error())
			return
		}
		go s.ServeConn(conn)
	}
}

func (s *Server) ServeConn(conn io.ReadWriteCloser) {
	session := NewSession(conn)
	defer session.Close()

	for {
		cmd, err := session.ReadCommand()
		if err != nil {
			break
		}

		name := strings.ToUpper(string(cmd[0]))
		handler, ok := s.handlers[name]

		if !ok {
			session.Write(ErrorReply("Handler Not Found").Bytes())
		} else {
			handler.Serve(session, cmd)
		}

	}
}
