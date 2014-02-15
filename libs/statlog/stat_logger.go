package statlog

/*
用于实现状态输出的日志组件
insert  query update delete    res faults  conn repl       time
    *6    112    *49    *12  25.5g     27 10096  SLV   00:25:23
    *6     97    *39     *6  25.5g     18 10095  SLV   00:25:24
    *5    118    *22     *7  25.5g      6 10094  SLV   00:25:25

slog := statlog.NewStatLogger(os.Stdout)
opt := &statlog.Opt{Padding: 8}

slog.Add(statlog.TimeItem("time"))
slog.Add(statlog.Item("total", func() interface{} {
	return "10"
}, opt))
slog.Add(statlog.Item("buffer", func() interface{} {
	return 10342
}, opt))

slog.Start()
*/
import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"
)

// ==============================
// 状态日志打印
// ==============================
type StatLogger struct {
	wr                  io.Writer
	items               []*StatItem
	skipFirstLine       bool // 第一行数据统计不完全，需要跳过
	beforeFunc          func()
	afterFunc           func()
	TitleOutputInterval int // 输出标题的间隔, 需要的话增加setter/getter
	shouldStop          bool
}

func NewStatLogger(wr io.Writer) (s *StatLogger) {
	s = &StatLogger{}
	s.wr = wr
	s.items = make([]*StatItem, 0, 25)
	s.skipFirstLine = false
	s.TitleOutputInterval = 10
	return
}

func (s *StatLogger) BeforePrint(fn func()) {
	s.beforeFunc = fn
}

func (s *StatLogger) AfterPrint(fn func()) {
	s.afterFunc = fn
}

func (s *StatLogger) Add(item *StatItem) {
	s.items = append(s.items, item)
}

func (s *StatLogger) Start() {
	s.shouldStop = false
	ticker := time.NewTicker(time.Millisecond * 1000)
	// 当然打印间隔
	printInterval := 0
	for _ = range ticker.C {
		if s.shouldStop {
			break
		}
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

		if s.beforeFunc != nil {
			s.beforeFunc()
		}
		// 输出数据
		s.printLine()
		if s.afterFunc != nil {
			s.afterFunc()
		}
		printInterval++
	}
}

func (s *StatLogger) Stop() {
	s.shouldStop = true
}

func (s *StatLogger) printTitle() {
	buf := bytes.Buffer{}
	for _, item := range s.items {
		fmtStr := "%" + strconv.Itoa(item.Option.Padding) + "s"
		buf.WriteString(fmt.Sprintf(fmtStr, item.Name))
	}
	buf.WriteByte('\n')
	io.WriteString(s.wr, buf.String())
}

func (s *StatLogger) printLine() {
	buf := bytes.Buffer{}
	for _, item := range s.items {
		fmtStr := "%" + strconv.Itoa(item.Option.Padding) + "s"
		buf.WriteString(fmt.Sprintf(fmtStr, fmt.Sprint(item.Text())))
	}
	buf.WriteByte('\n')
	io.WriteString(s.wr, buf.String())
}
