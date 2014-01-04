package stdlog

import (
	"fmt"
	"io"
)

type Logger interface {
	SetPrefix(fn func() string)
	GetPrefix() func() string
	Println(v ...interface{})
	Printf(format string, v ...interface{})
	Print(v ...interface{})
}

type SimpleLogger struct {
	Logger
	out    io.Writer
	prefix func() string
}

func NewSimpleLogger(out io.Writer) (l *SimpleLogger) {
	l = &SimpleLogger{}
	l.out = out
	return
}

// 设置函数前缀，典型使用场景是返回当前时间字符串
func (l *SimpleLogger) SetPrefix(fn func() string) {
	l.prefix = fn
}

func (l *SimpleLogger) GetPrefix() (fn func() string) {
	return l.prefix
}

// 输出函数
func (l *SimpleLogger) ouput(s string) {
	if l.prefix != nil {
		io.WriteString(l.out, l.prefix())
	}
	io.WriteString(l.out, s)
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
