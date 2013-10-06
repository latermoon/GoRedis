/*
logger.SetFormat("%time %target %line %txt")
logger.Println(txt)


*/
package golog

import (
	"io"
	"sync"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	out   io.Writer
	mu    sync.Mutex
	Level level
}

func New(out io.Writer, lvl Level) (logger *Logger) {
	logger = &Logger{}
	logger.Level = lvl
	return
}

func (l *Logger) Log(lvl Level, v ...interface{}) {

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
