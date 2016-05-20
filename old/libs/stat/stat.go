package stat

import (
	"fmt"
	"io"
	"time"
)

func New(out io.Writer) (w *Writer) {
	w = &Writer{
		wr:            out,
		items:         make([]Item, 0, 20),
		titleInterval: 10,
	}
	return
}

func TimeString() string {
	t := time.Now()
	return fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
}
