package statlog

import (
	"fmt"
	"time"
)

// 打印选项
type Opt struct {
	Padding int
}

// ==============================
// 单条打印数据
// ==============================
type StatItem struct {
	Name   string
	Text   func() interface{}
	Option *Opt
}

func Item(name string, text func() interface{}, opt *Opt) (item *StatItem) {
	item = &StatItem{}
	item.Name = name
	item.Text = text
	item.Option = opt
	return
}

// 时间字段
func TimeItem(name string) (item *StatItem) {
	return Item(name, func() interface{} {
		t := time.Now()
		return fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
	}, &Opt{Padding: 8})
}
