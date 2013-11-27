// Copyright (c) 2013, Latermoon <lptmoon@gmail.com>
// All rights reserved.
//
// Go版RedisServer
// @author latermoon
// @since 2013-08-14
// @last 2013-09-07

package goredis

import (
	"fmt"
	"net"
	"reflect"
	"runtime/debug"
	"strings"
)

const (
	CR   = '\r'
	LF   = '\n'
	CRLF = "\r\n"
)

// 命令处理接口
type CommandHandler interface {
	// 首先搜索"On+大写NAME"格式的函数，存在则调用，不存在则调用On
	// OnGET(cmd *Command) (reply *Reply)
	// OnGET(session *Session, cmd *Command) (reply *Reply)
	OnUndefined(session *Session, cmd *Command) (reply *Reply)
	On(session *Session, cmd *Command)
}

// 一个空的默认命令处理对象
type emptyCommandHandler struct {
	CommandHandler
}

func (s *emptyCommandHandler) On(session *Session, cmd *Command) {
}

func (s *emptyCommandHandler) OnUndefined(session *Session, cmd *Command) (reply *Reply) {
	return ErrorReply("Not Supported: " + cmd.String())
}

// ==============================
// RedisServer只实现最基本的Redis协议
// 提供On接口处理传入的各种指令，使用session返回数据
// ==============================
type RedisServer struct {
	// 指定的处理程序
	handler CommandHandler
	// 缓存处理函数，减少relect次数
	methodCache map[string]reflect.Value
}

func NewRedisServer(handler CommandHandler) (server *RedisServer) {
	server = &RedisServer{}
	server.SetHandler(handler)
	return
}

func (server *RedisServer) SetHandler(handler CommandHandler) {
	server.handler = handler
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

	// init
	server.methodCache = make(map[string]reflect.Value)
	if server.handler == nil {
		server.SetHandler(&emptyCommandHandler{})
	}

	// run loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("[goredis] accepted error", err)
			continue
		}
		fmt.Println("[goredis] connection accepted from", conn.RemoteAddr())
		// go
		go server.handleConnection(NewSession(conn))
	}
}

func (server *RedisServer) InvokeCommandHandler(session *Session, cmd *Command) (reply *Reply) {
	cmdName := strings.ToUpper(cmd.Name())
	// 从Cache取出处理函数
	method, exists := server.methodCache[cmdName]
	if !exists {
		method = reflect.ValueOf(server.handler).MethodByName("On" + cmdName)
		server.methodCache[cmdName] = method
	}

	// 通用调用
	server.handler.On(session, cmd)
	if method.IsValid() {
		// 可以调用两种接口
		// method = OnXXX(cmd *Command) (reply *Reply)
		// method = OnXXX(session *Session, cmd *Command) (reply *Reply)
		var in []reflect.Value
		if method.Type().NumIn() == 1 {
			in = []reflect.Value{reflect.ValueOf(cmd)}
		} else {
			in = []reflect.Value{reflect.ValueOf(session), reflect.ValueOf(cmd)}
		}
		callResult := method.Call(in)
		reply = callResult[0].Interface().(*Reply)
	} else {
		reply = server.handler.OnUndefined(session, cmd)
	}
	return
}

// 处理一个客户端连接
func (server *RedisServer) handleConnection(session *Session) {
	// 异常处理
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(fmt.Sprintf("Error %s %s", session.conn.RemoteAddr(), err))
			debug.PrintStack()
			session.Close()
		}
	}()
	for {
		cmd, e1 := session.ReadCommand()
		// 常见的error是:
		// 1) io.EOF
		// 2) read tcp 127.0.0.1:51863: connection reset by peer
		if e1 != nil {
			panic(fmt.Sprintf("end connection %s", e1))
		}
		// 处理
		reply := server.InvokeCommandHandler(session, cmd)
		if reply != nil {
			session.Reply(reply)
		}
	}
}
