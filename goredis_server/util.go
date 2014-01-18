package goredis_server

import (
	"strconv"
)

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'g', 12, 64)
}
