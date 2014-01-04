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
	"os"
	"sync"
)

var (
	stdlogger Logger            // 默认Logger
	caches    map[string]Logger // 缓存
	mu        sync.Mutex
	prefix    func() string // 默认前缀函数
)

func init() {
	caches = make(map[string]Logger)
	stdlogger = Log("") // 默认
}

// 获取一个指定的Logger
func Log(name string) (l Logger) {
	mu.Lock()
	defer mu.Unlock()
	var ok bool
	l, ok = caches[name]
	if !ok {
		l = NewSimpleLogger(os.Stdout) // 默认输出到os.Stdout
		if prefix != nil {
			l.SetPrefix(prefix)
		}
		caches[name] = l
	}
	return
}

func SetPrefix(fn func() string) {
	mu.Lock()
	defer mu.Unlock()
	stdlogger.SetPrefix(fn)
	// 如果存在没初始化的Prefix，统一设置
	for _, l := range caches {
		if l.GetPrefix() == nil {
			l.SetPrefix(fn)
		}
	}
}

func Println(v ...interface{}) {
	stdlogger.Println(v...)
}

func Printf(format string, v ...interface{}) {
	stdlogger.Printf(format, v...)
}

func Print(v ...interface{}) {
	stdlogger.Print(v...)
}
