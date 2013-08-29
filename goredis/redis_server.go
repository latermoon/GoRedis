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
	server := goredis.NewRedisServer()

	// KeyValue
	kvCache := make(map[string]interface{})
	// Set操作的写锁
	chanSet := make(chan int, 1)

	server.On("GET", func(cmd *goredis.Command) (reply *goredis.Reply) {
		key := cmd.StringAtIndex(1)
		value := kvCache[key]
		reply = goredis.BulkReply(value)
		return
	})

	server.On("SET", func(cmd *goredis.Command) (reply *goredis.Reply) {
		key := cmd.StringAtIndex(1)
		value := cmd.StringAtIndex(2)
		chanSet <- 0
		kvCache[key] = value
		<-chanSet
		reply = goredis.StatusReply("OK")
		return
	})

	server.On("INFO", func(cmd *goredis.Command) (reply *goredis.Reply) {
		reply = goredis.BulkReply("GoRedis 0.1 by latermoon\n")
		return
	})

	// 开始监听端口
	fmt.Println("Listen :8002")
	server.Listen(":8002")
*/
// ==============================
type RedisServer struct {
	// 存放指令处理函数
	handlers map[string](func(cmd *Command) (reply *Reply))
}

// 创建服务实例
func NewRedisServer() (server *RedisServer) {
	server = &RedisServer{}
	server.handlers = make(map[string](func(cmd *Command) (reply *Reply)))
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

		session := newSession(conn)

		go server.handleConnection(session)
	}
}

// 添加指令处理函数
func (server *RedisServer) On(commandName string, handler func(cmd *Command) (reply *Reply)) {
	name := strings.ToUpper(commandName)
	server.handlers[name] = handler
}

// 处理一个客户端连接
func (server *RedisServer) handleConnection(session *Session) {
	// 不断从一个连接中获取命令，并处理，返回
	for {
		cmd, e1 := session.ReadCommand()
		// 常见的error是:
		// 1) io.EOF
		// 2) read tcp 127.0.0.1:51863: connection reset by peer
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
