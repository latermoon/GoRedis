package stdlog

import (
	"io"
)

// 把一个io.Writer接口绑定到多个writer输出
// 常用于输出到文件的同时，也输出到os.Stdout, NewMultiWriter(os.Stdout, logfile)
type MultiWriter struct {
	io.Writer
	writers []io.Writer
}

func NewMultiWriter(w ...io.Writer) (m *MultiWriter) {
	m = &MultiWriter{}
	m.writers = w
	return
}

func (m *MultiWriter) Write(p []byte) (n int, err error) {
	for _, writer := range m.writers {
		n, err = writer.Write(p)
	}
	return
}
