package util

import (
	"os"
)

func FileExist(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
