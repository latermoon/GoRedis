// Go版RedisServer
// @author latermoon
// @since 2013-08-14

package goredis

import (
	"bufio"
	"github.com/op/go-logging"
	"net"
	"strconv"
	"strings"
)

const (
	CR   = '\r'
	LF   = '\n'
	CRLF = "\r\n"
)

var logger = logging.MustGetLogger("goredis")

// ==============================
// RedisServer只实现最基本的Redis协议
// 提供On接口处理传入的各种指令，使用session返回数据
/*
	server, _ := goredis.NewRedisServer()

	// KeyValue
	kvCache := make(map[string]interface{})
	// Set操作的写锁
	chanSet := make(chan int, 1)

	server.On("GET", func(session *goredis.Session, cmd *goredis.Command) (err error) {
		err = nil
		key, _ := cmd.StringAtIndex(1)
		value := kvCache[key]
		session.ReplyBulk(value)
		return
	})

	server.On("SET", func(session *goredis.Session, cmd *goredis.Command) (err error) {
		key, _ := cmd.StringAtIndex(1)
		value, _ := cmd.StringAtIndex(2)
		chanSet <- 0
		kvCache[key] = value
		<-chanSet
		session.ReplyStatus("OK")
		return
	})

	server.Listen(":8002")
*/
// ==============================
type RedisServer interface {
	Listen(host string)
	On(commandName string, fn func(session *Session, cmd *Command) (err error))
}

// Implemented
type SimpleRedisServer struct {
	// 存放指令处理函数
	handlers map[string](func(session *Session, cmd *Command) (err error))
}

// 创建服务实例
func NewRedisServer() (server RedisServer, err error) {
	simpleServer := &SimpleRedisServer{}
	err = nil
	simpleServer.handlers = make(map[string](func(session *Session, cmd *Command) (err error)))
	server = simpleServer
	return
}

func (server *SimpleRedisServer) Init() {
	server.handlers = make(map[string](func(session *Session, cmd *Command) (err error)))
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
			logger.Warning("[goredis] accepted error %s", err)
			continue
		}
		logger.Info("[goredis] connection accepted from %s", conn.RemoteAddr())
		session := newSession(conn)
		go server.handleConnection(session)
	}
}

// 添加指令处理函数
func (server *SimpleRedisServer) On(commandName string, fn func(session *Session, cmd *Command) (err error)) {
	name := strings.ToUpper(commandName)
	server.handlers[name] = fn
}

// 处理一个客户端连接
func (server *SimpleRedisServer) handleConnection(session *Session) {
	reader := bufio.NewReader(session.conn)

	// 不断从一个连接中获取命令，并处理，返回
	for {
		cmd, e1 := readCommand(reader)
		if e1 != nil {
			logger.Info("[goredis] end connection %s %s", e1, session.conn.RemoteAddr())
			session.Close()
			return
		}
		// 取出处理函数
		fn, ok := server.handlers[strings.ToUpper(cmd.Name())]
		if ok {
			e2 := fn(session, cmd)
			if e2 != nil {
				logger.Warning("[goredis] e2: %s", e2)
			}
		}
	}
}

/*
// 从客户端连接获取指令
// (下面读取过程，线上应用前需要增加错误校验，数据大小限制)
*<number of arguments> CR LF
$<number of bytes of argument 1> CR LF
<argument data> CR LF
...
$<number of bytes of argument N> CR LF
<argument data> CR LF
*/
func readCommand(reader *bufio.Reader) (cmd *Command, err error) {
	cmd = &Command{}
	err = nil

	// Read ( *<number of arguments> CR LF )
	_, err = reader.ReadBytes('*')
	if err != nil { // EOF
		return
	}
	line, e1 := reader.ReadBytes(CR)
	if e1 != nil {
		err = e1
		return
	}
	// number of arguments
	argCount, _ := strconv.Atoi(string(line[:len(line)-1]))
	_, err = reader.ReadBytes(LF)
	if err != nil {
		return
	}

	cmd.Args = make([][]byte, argCount)
	for i := 0; i < argCount; i++ {
		// Read ( $<number of bytes of argument 1> CR LF )
		_, _ = reader.ReadBytes('$')
		line, e2 := reader.ReadBytes(CR)
		if e2 != nil {
			err = e2
			return
		}
		argSize, _ := strconv.Atoi(string(line[:len(line)-1]))
		_, err = reader.ReadBytes(LF)

		// Read ( <argument data> CR LF )
		cmd.Args[i] = make([]byte, argSize)
		_, err = reader.Read(cmd.Args[i])
		_, err = reader.ReadBytes(CR)
		_, err = reader.ReadBytes(LF)
		if err != nil {
			return
		}
	}

	return
}
