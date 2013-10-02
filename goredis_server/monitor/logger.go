package monitor

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Formater interface {
	Padding() int
	Title() string
	Text() string
}

/*
用于实现mongostat的效果
insert  query update delete getmore command flushes mapped  vsize    res faults  locked db idx miss %     qr|qw   ar|aw  netIn netOut  conn repl       time
    *6    112    *49    *12       0    12|0       0   126g   269g  25.5g     27    .:20.3%          0       0|0     0|0    15k    22k 10096  SLV   00:25:23
    *6     97    *39     *6       0    10|0       0   126g   269g  25.5g     18    .:10.5%          0       0|0     1|1    12k    30k 10095  SLV   00:25:24
    *5    118    *22     *7       0    12|0       0   126g   269g  25.5g      6     .:3.7%          0       0|0     0|0    15k    20k 10094  SLV   00:25:25
*/
type StatusLogger struct {
	logfile             *log.Logger
	items               []Formater
	skipFirstLine       bool // 第一行数据统计不完全，需要跳过
	TitleOutputInterval int  // 输出标题的间隔
}

func NewStatusLogger(path string) (logger *StatusLogger) {
	logger = &StatusLogger{}
	logger.items = make([]Formater, 0)
	logger.skipFirstLine = false
	logger.TitleOutputInterval = 10
	file1, e1 := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if e1 != nil {
		panic(e1)
	}
	logger.logfile = log.New(file1, "", 0)
	return
}

func (s *StatusLogger) Add(formater Formater) {
	s.items = append(s.items, formater)
}

func (s *StatusLogger) Start() {
	ticker := time.NewTicker(time.Millisecond * 1000)
	go func() {
		// 当然打印间隔
		printInterval := 0
		for _ = range ticker.C {
			// 跳过第一行错误数据，并用于打印title
			if !s.skipFirstLine {
				s.skipFirstLine = true
				s.printTitle()
				continue
			}
			if printInterval >= s.TitleOutputInterval {
				s.printTitle()
				printInterval = 0
			}
			s.printLine()
			printInterval++
		}
	}()
}

func (s *StatusLogger) printTitle() {
	buf := bytes.Buffer{}
	for _, item := range s.items {
		fmtStr := "%" + strconv.Itoa(item.Padding()) + "s"
		buf.WriteString(fmt.Sprintf(fmtStr, item.Title()))
	}
	s.logfile.Println(buf.String())
}

func (s *StatusLogger) printLine() {
	buf := bytes.Buffer{}
	for _, item := range s.items {
		fmtStr := "%" + strconv.Itoa(item.Padding()) + "s"
		buf.WriteString(fmt.Sprintf(fmtStr, item.Text()))
	}
	s.logfile.Println(buf.String())
}
