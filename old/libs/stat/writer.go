package stat

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"
)

type Item interface {
	Title() string
	Padding() int
	Value() interface{}
}

type Writer struct {
	wr            io.Writer
	items         []Item
	skipFirstLine bool // 第一行数据统计不完全，需要跳过
	titleInterval int  // 输出标题的间隔, 需要的话增加setter/getter
	stop          bool
}

func (w *Writer) Add(item Item) {
	w.items = append(w.items, item)
}

func (w *Writer) Start() {
	if w.stop {
		return
	}
	ticker := time.NewTicker(time.Millisecond * 1000)
	// 打印间隔
	printInterval := 0
	for _ = range ticker.C {
		if w.stop {
			break
		}
		// 跳过第一行不准确的数据，并用于打印title
		if !w.skipFirstLine {
			w.skipFirstLine = true
			w.printTitle()
			continue
		}
		if printInterval >= w.titleInterval {
			w.printTitle()
			printInterval = 0
		}
		w.printLine() // output
		printInterval++
	}
	ticker.Stop()
}

func (w *Writer) Close() {
	w.stop = true
}

func (w *Writer) printTitle() {
	buf := bytes.Buffer{}
	for _, item := range w.items {
		fmtStr := "%" + strconv.Itoa(item.Padding()) + "s"
		buf.WriteString(fmt.Sprintf(fmtStr, item.Title()))
	}
	buf.WriteByte('\n')
	io.WriteString(w.wr, buf.String())
}

func (w *Writer) printLine() {
	buf := bytes.Buffer{}
	for _, item := range w.items {
		fmtStr := "%" + strconv.Itoa(item.Padding()) + "s"
		obj := item.Value()
		switch obj.(type) {
		case func() interface{}:
			obj = obj.(func() interface{})()
		case func() string:
			obj = obj.(func() string)()
		}
		buf.WriteString(fmt.Sprintf(fmtStr, fmt.Sprint(obj)))
	}
	buf.WriteByte('\n')
	io.WriteString(w.wr, buf.String())
}
