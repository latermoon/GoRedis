package stdlog

/*
绑定到stdout和stderr的日志输出，通过修改os.Stdout和os.Stderr的文件指向，实现全局日志统一输出
stdlog.Println("...")
stdlog.Errorln("...")
*/
import (
	"fmt"
	"os"
)

func ouput(f *os.File, s string) {
	f.WriteString(s)
}

func Print(v ...interface{}) {
	ouput(os.Stdout, fmt.Sprint(v...))
}

func Printf(format string, v ...interface{}) {
	ouput(os.Stdout, fmt.Sprintf(format, v...))
}

func Println(v ...interface{}) {
	ouput(os.Stdout, fmt.Sprintln(v...))
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
