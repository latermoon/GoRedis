package stdlog

/*
减少侵入、通用、灵活的日志，适用于嵌入类库里，只提供stdout和stderr

绑定到stdout和stderr的日志输出，通过修改os.Stdout和os.Stderr的文件指向，实现全局日志统一输出
stdlog.Println("...")
stdlog.Errorln("...")

增加函数参数支持，实现灵活的时间和前缀输出，函数接口为func()string
timefmt := func() string {
	return fmt.Sprintf("[%d-%02d-%02d %02d:%02d:%02d]", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}
stdlog.Println(timeprefix, "init ...")
Output:
	[2013-12-08 21:50:29] init ...
*/
import (
	"bytes"
	"fmt"
	"os"
)

// 原始方法
func ouput(f *os.File, s string) {
	f.WriteString(s)
}

// 对于函数和字符串混合的情况，统一格式化
func sprint(v ...interface{}) string {
	hasfunc := false
	for _, e := range v {
		switch e.(type) {
		case func() string:
			hasfunc = true
			break
		}
	}
	if !hasfunc {
		return fmt.Sprint(v...)
	} else {
		buf := &bytes.Buffer{}
		count := len(v)
		for i := 0; i < count; i++ {
			switch v[i].(type) {
			case func() string:
				buf.WriteString(v[i].(func() string)())
			default:
				buf.WriteString(fmt.Sprint(v[i]))
			}
			if i < count-1 {
				buf.WriteString(" ")
			}
		}
		return buf.String()
	}
}

func sprintln(v ...interface{}) string {
	return sprint(v...) + "\n"
}

func Print(v ...interface{}) {
	ouput(os.Stdout, sprint(v...))
}

func Printf(format string, v ...interface{}) {
	ouput(os.Stdout, fmt.Sprintf(format, v...))
}

func Println(v ...interface{}) {
	ouput(os.Stdout, sprintln(v...))
}

func Error(v ...interface{}) {
	ouput(os.Stderr, fmt.Sprint(v...))
}

func Errorf(format string, v ...interface{}) {
	ouput(os.Stderr, fmt.Sprintf(format, v...))
}

func Errorln(v ...interface{}) {
	ouput(os.Stderr, fmt.Sprintln(v...))
}
