/*
简易日志，通用接口，以后可以直接替换成其他成熟logger
支持log(fmt string)和log(func() string)
*/
package golog

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

type Level int

// 调试级别
const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

// 要输出的字段
const (
	F_Time = 1 << iota
	F_Level
	F_FileLine
	F_FuncLine
	F_MSG
)

// 日志类
type Logger struct {
	out   io.Writer
	mu    sync.Mutex
	level Level
}

func New(out io.Writer, lvl Level) (logger *Logger) {
	logger = &Logger{}
	logger.out = out
	logger.level = lvl
	return
}

func (l *Logger) SetLevel(lvl Level) {
	l.level = lvl
}

func (l *Logger) Level() Level {
	return l.level
}

// http://zh.wikipedia.org/wiki/%E5%90%84%E5%9C%B0%E6%97%A5%E6%9C%9F%E5%92%8C%E6%97%B6%E9%97%B4%E8%A1%A8%E7%A4%BA%E6%B3%95
func (l *Logger) formatTime(t *time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

// v[0]可以为func()string
func (l *Logger) Log(lvl Level, v ...interface{}) {
	if lvl < l.level {
		return
	}
	buf := bytes.Buffer{}

	now := time.Now()
	buf.WriteString("[")
	buf.WriteString(l.formatTime(&now))
	buf.WriteString("] ")

	firstArg := v[0]
	switch firstArg.(type) {
	case string:
		msg := fmt.Sprintf(firstArg.(string), v[1:]...)
		buf.WriteString(msg)
	case func() string:
		msg := firstArg.(func() string)()
		buf.WriteString(msg)
	default:
		buf.WriteString("bad log args")
	}

	io.WriteString(l.out, buf.String())
	io.WriteString(l.out, "\n")
}

func (l *Logger) Debug(v ...interface{}) {
	l.Log(DEBUG, v...)
}

func (l *Logger) Info(v ...interface{}) {
	l.Log(INFO, v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.Log(WARN, v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.Log(ERROR, v...)
}
