// Go版RedisServer
// @author latermoon
// @since 2013-08-14

package goredis

import (
	"fmt"
	"net"
	"strings"
)

const (
	CR   = '\r'
	LF   = '\n'
	CRLF = "\r\n"
)

// ==============================
// RedisServer只实现最基本的Redis协议
// 提供On接口处理传入的各种指令，使用session返回数据
/*

*/
// ==============================
type RedisServer interface {
	Listen(host string)
	On(commandName string, fn func(cmd *Command) (reply *Reply))
}

// Implemented
type SimpleRedisServer struct {
	// 存放指令处理函数
	handlers map[string](func(cmd *Command) (reply *Reply))
}

// 创建服务实例
func NewRedisServer() (server RedisServer, err error) {
	simpleServer := &SimpleRedisServer{}
	err = nil
	simpleServer.handlers = make(map[string](func(cmd *Command) (reply *Reply)))
	server = simpleServer
	return
}

/**
 * 开始监听主机端口
 * @param host "localhost:6379"
 */
func (server *SimpleRedisServer) Listen(host string) {
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
		session := newSession(conn)
		go server.handleConnection(session)
	}
}

// 添加指令处理函数
func (server *SimpleRedisServer) On(commandName string, fn func(cmd *Command) (reply *Reply)) {
	name := strings.ToUpper(commandName)
	server.handlers[name] = fn
}

// 处理一个客户端连接
func (server *SimpleRedisServer) handleConnection(session *Session) {
	// 不断从一个连接中获取命令，并处理，返回
	for {
		cmd, e1 := session.ReadCommand()
		if e1 != nil {
			fmt.Println("[goredis] end connection", e1, session.conn.RemoteAddr())
			session.Close()
			return
		}
		// 取出处理函数
		handler, exists := server.handlers[strings.ToUpper(cmd.Name())]
		if exists {
			reply := handler(cmd)
			session.Reply(reply)
		} else {
			session.Reply(ErrorReply("Not Supported"))
		}
	}
}
