package server

import (
	. "github.com/latermoon/GoRedis/redis"
	"github.com/latermoon/GoRedis/rocks"
	"log"
	"reflect"
	"strings"
)

type GoRedisServer struct {
	ServerHandler
	db      *rocks.DB
	cmdfunc map[string]HandlerFunc
}

func New(db *rocks.DB) *GoRedisServer {
	s := &GoRedisServer{db: db}
	s.registerCmdFunc()
	return s
}

func (s *GoRedisServer) SessionOpened(sess *Session) {
	log.Println("connection accepted from", sess.RemoteAddr())
}

func (s *GoRedisServer) SessoinClosed(sess *Session, err error) {
	log.Println("end connection", sess.RemoteAddr(), err)
}

func (s *GoRedisServer) RecvCommand(sess *Session, c Command) {
	log.Println("command:", c)

	// invoke On[Command] functions
	cmdname := strings.ToUpper(string(c[0]))
	cmdFunc, ok := s.cmdfunc[cmdname]
	if !ok {
		sess.WriteReply(ErrorReply("Handler Not Found"))
	}
	cmdFunc(sess, c)
}

// register all On[Comamd Name] functions, such as OnPING/OnGET/OnSET, into HandlerFunc map
func (s *GoRedisServer) registerCmdFunc() {
	s.cmdfunc = make(map[string]HandlerFunc)

	objval := reflect.ValueOf(s)
	objtyp := reflect.TypeOf(s)
	for i := 0; i < objtyp.NumMethod(); i++ {
		name := objtyp.Method(i).Name
		if len(name) > 2 && strings.HasPrefix(name, "On") {
			// tricks
			func(name string, method reflect.Value) {
				s.cmdfunc[name] = HandlerFunc(func(r ReplyWriter, c Command) {
					in := []reflect.Value{reflect.ValueOf(r), reflect.ValueOf(c)}
					method.Call(in)
				})
			}(strings.ToUpper(name[2:]), objval.Method(i))
		}
	}
}

type ReplyWriter interface {
	WriteReply(Reply) (int, error)
}

type HandlerFunc func(ReplyWriter, Command)

func (f HandlerFunc) Serve(r ReplyWriter, c Command) {
	f(r, c)
}
