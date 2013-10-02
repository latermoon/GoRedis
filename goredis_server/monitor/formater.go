package monitor

import (
	"fmt"
	"strconv"
	"time"
)

type CommonFormater struct {
	Formater
	title   string
	padding int
}

func (f *CommonFormater) Padding() int {
	return f.padding
}

func (f *CommonFormater) Title() string {
	return f.title
}

func (f *CommonFormater) Text() string {
	return ""
}

// ==========CountFormater==========
type CountFormater struct {
	CommonFormater
	counter *Counter
}

func NewCountFormater(counter *Counter, title string, padding int) (f *CountFormater) {
	f = &CountFormater{}
	f.counter = counter
	f.title = title
	f.padding = padding
	return
}

func (f *CountFormater) Text() string {
	chg := f.counter.ChangedCount()
	return strconv.Itoa(chg)
}

// ==========TimeFormater==========
type TimeFormater struct {
	CommonFormater
}

func NewTimeFormater(title string, padding int) (f *TimeFormater) {
	f = &TimeFormater{}
	f.title = title
	f.padding = padding
	return
}

func (f *TimeFormater) Text() (s string) {
	t := time.Now()
	s = fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
	//s = fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	return
}
