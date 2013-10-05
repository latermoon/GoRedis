/*
logger.SetFormat("%time %target %line %txt")
logger.Println(txt)


*/
package golog

import (
	"io"
	"sync"
)

type DebugLevel int

const (
	DEBUG = 0
	INFO  = 1
	WARN  = 2
	ERROR = 4
)

type Logger struct {
	out io.Writer
	mu  sync.Mutex
}

func New(out io.Writer) (logger *Logger) {
	logger = &Logger{}
	return
}

func (l *Logger) Println(v ...interface{}) {

}

func (l *Logger) Debug(v ...interface{}) {

}

func (l *Logger) Info(v ...interface{}) {

}

func (l *Logger) Warn(v ...interface{}) {

}

func (l *Logger) Error(v ...interface{}) {

}
