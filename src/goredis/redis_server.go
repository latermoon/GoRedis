// Go版RedisServer
// @author latermoon
// @since 2013-08-14

package goredis

import (
	"fmt"
	"net"
	"reflect"
	"strings"
)

const (
	CR   = '\r'
	LF   = '\n'
	CRLF = "\r\n"
)

type CommandHandler interface {
	On(name string, cmd *Command) (reply *Reply)
}

// ==============================
// RedisServer只实现最基本的Redis协议
// 提供On接口处理传入的各种指令，使用session返回数据
// ==============================
type RedisServer struct {
	// 指定的处理程序
	handler CommandHandler
	// 缓存处理函数
	methodCache map[string]reflect.Value
}

// 创建服务实例
func NewRedisServer() (server *RedisServer) {
	server = &RedisServer{}
	server.Init()
	return
}

func (server *RedisServer) Init() {
	server.methodCache = make(map[string]reflect.Value)
}

func (server *RedisServer) SetHanlder(handler CommandHandler) {
	server.handler = handler
}

func (server *RedisServer) On(name string, cmd *Command) (reply *Reply) {
	reply = ErrorReply("Not Supported: " + name)
	return
}

/**
 * 开始监听主机端口
 * @param host "localhost:6379"
 */
func (server *RedisServer) Listen(host string) {
	listener, e1 := net.Listen("tcp", host)
	if e1 != nil {
		panic(e1)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("[goredis] accepted error", err)
			continue
		}
		fmt.Println("[goredis] connection accepted from", conn.RemoteAddr())
		// gogogo
		go server.handleConnection(newSession(conn))
	}
}

// 处理一个客户端连接
func (server *RedisServer) handleConnection(session *Session) {
	// 不断从一个连接中获取命令，并处理，返回
	for {
		cmd, e1 := ReadCommand(session.reader)
		// 常见的error是:
		// 1) io.EOF
		// 2) read tcp 127.0.0.1:51863: connection reset by peer
		if e1 != nil {
			fmt.Println("[goredis] end connection", e1, session.conn.RemoteAddr())
			session.Close()
			return
		}
		// 初始化
		cmd.session = session
		cmdName := strings.ToUpper(cmd.Name())
		// 从Cache取出处理函数
		method, exists := server.methodCache[cmdName]
		if !exists {
			method = reflect.ValueOf(server.handler).MethodByName("On" + cmdName)
			server.methodCache[cmdName] = method
		}
		if method.IsValid() {
			callResult := method.Call([]reflect.Value{reflect.ValueOf(cmd)})
			reply := callResult[0].Interface().(*Reply)
			session.Reply(reply)
		} else {
			reply := server.handler.On(cmdName, cmd)
			session.Reply(reply)
		}
	}
}
