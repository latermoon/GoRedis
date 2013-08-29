package api

import (
	"fmt"
)

func MyAPIFunc() string {
	return "laterm"
}

func init() {
	fmt.Println("api.init()")
}
