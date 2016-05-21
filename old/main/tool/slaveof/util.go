package slaveof

import (
	"fmt"
)

func bytesInHuman(size int64) string {
	f := float64(size)
	if f > 1024*1024*1024*1024 {
		return fmt.Sprintf("%0.1fT", f/1024/1024/1024/1024)
	}
	if f > 1024*1024*1024 {
		return fmt.Sprintf("%0.1fG", f/1024/1024/1024)
	}
	if f > 1024*1024 {
		return fmt.Sprintf("%0.1fM", f/1024/1024)
	}
	if f > 1024 {
		return fmt.Sprintf("%0.1fK", f/1024)
	}
	return fmt.Sprintf("%dB", size)
}
