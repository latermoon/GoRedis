// Copyright 2013 Latermoon. All rights reserved.

package stdlog

import (
	"fmt"
	"io"
	"sync"
)

type Logger interface {
	SetPrefix(fn func() string)
	SetOutput(w io.Writer)
	// Print
	Println(v ...interface{})
	Printf(format string, v ...interface{})
	Print(v ...interface{})
}

type SimpleLogger struct {
	Logger
	prefix func() string
	out    io.Writer
	mu     sync.Mutex
}

// 设置函数前缀，典型使用场景是返回当前时间字符串
func (l *SimpleLogger) SetPrefix(fn func() string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = fn
}

func (l *SimpleLogger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

// 输出函数，如果没有设置Prefix和Output，则使用全局配置
func (l *SimpleLogger) ouput(s string) {
	// prefix函数时间不可控，先执行
	var p string
	if l.prefix != nil {
		p = l.prefix()
	} else if Prefix() != nil {
		p = Prefix()()
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	w := l.out
	if l.out == nil {
		w = Output()
	}
	// write
	if len(p) > 0 {
		io.WriteString(w, p+s)
	} else {
		io.WriteString(w, s)
	}
}

func (l *SimpleLogger) Println(v ...interface{}) {
	l.ouput(fmt.Sprintln(v...))
}

func (l *SimpleLogger) Printf(format string, v ...interface{}) {
	l.ouput(fmt.Sprintf(format, v...))
}

func (l *SimpleLogger) Print(v ...interface{}) {
	l.ouput(fmt.Sprint(v...))
}
