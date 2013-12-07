package stat

import (
	// "fmt"
	"github.com/latermoon/GoRedis/libs/statlog"
	"github.com/latermoon/redigo/redis"
	"os"
	"strings"
	"time"
)

type Field struct {
	Name          string
	DisplayName   string
	UseByteFormat bool
}

type StatClient struct {
	pool        *redis.Pool
	host        string
	section     string // 要获取的info section
	fields      []*Field
	fieldMap    map[string]*Field
	fieldValues map[string]string // 从info获取到的内容
}

func NewStatClient(host string, infoSection string) (s *StatClient) {
	s = &StatClient{}
	s.host = host
	s.section = infoSection
	s.fieldValues = make(map[string]string, 0)
	return
}

func (s *StatClient) SetFields(fieldAndNames ...string) {
	count := len(fieldAndNames)
	if count%2 != 0 {
		panic("bad fields pairs")
	}
	s.fields = make([]string, 0, count/2)
	s.fieldNames = make(map[string]string, count/2)
	for i := 0; i < count; i += 2 {
		field := fieldAndNames[i]
		name := fieldAndNames[i+1]
		s.fields = append(s.fields, field)
		s.fieldNames[field] = name
	}
}

func (s *StatClient) initRedisPool() {
	s.pool = &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", s.host)
			return c, err
		},
	}
}

func (s *StatClient) Connect() {
	s.initRedisPool()

	s.updateFieldValues()

	l := statlog.NewStatLogger(os.Stdout)

	l.BeforePrint(func() {
		go s.updateFieldValues()
	})

	l.Add(statlog.TimeItem("time"))

	for _, field := range s.fields {
		name := s.fieldNames[field]
		// copy local variable
		fn := func(field string) func() interface{} {
			return func() interface{} {
				return s.fieldValues[field]
			}
		}(field)
		padding := 8
		if len(name) > 8 {
			padding = len(name) + 1
		}
		opt := &statlog.Opt{Padding: padding}
		l.Add(statlog.Item(name, fn, opt))
	}

	l.Start()
}

func (s *StatClient) updateFieldValues() {
	conn := s.pool.Get()
	reply, err := conn.Do("INFO", s.section)
	if err != nil {
		conn.Close()
		panic(err)
	}
	info := string(reply.([]byte))
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		// fmt.Println(line[:idx], line[idx+1:])
		s.fieldValues[line[:idx]] = line[idx+1:]
	}
	conn.Close()
}
