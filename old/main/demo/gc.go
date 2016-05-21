package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	cmdline := fmt.Sprintf("pmap %d", os.Getpid())
	out, err := exec.Command(cmdline).Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(out))
}
