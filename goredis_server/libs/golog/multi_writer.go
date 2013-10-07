package golog

import (
	//"bufio"
	"io"
	"os"
)

// 一个io.Writer接口，写多个io.Writer
type MultiWriter struct {
	io.Writer
	writers []io.Writer
}

// 同时写入Console和file的Writer
func NewConsoleAndFileWriter(path string) (m *MultiWriter) {
	m = &MultiWriter{}
	wr, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic("bad log path:" + path)
	}
	m.SetWriters(os.Stdout, (wr))
	return
}

// 绑定多个writer
func NewMultiWriter(writers ...io.Writer) (m *MultiWriter) {
	m = &MultiWriter{}
	m.writers = writers
	return
}

func (m *MultiWriter) SetWriters(writers ...io.Writer) {
	m.writers = writers
}

func (m *MultiWriter) Write(p []byte) (n int, err error) {
	for _, writer := range m.writers {
		n, err = writer.Write(p)
	}
	return
}
