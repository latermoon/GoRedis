package main

import (
	. "../goredis"
	//"bufio"
	"fmt"
	"net"
)

var redis006 = "10.80.100.193:6348"

func main() {
	innerConn, e1 := net.Dial("tcp", redis006)
	if e1 != nil {
		panic(e1)
	}

	sess := NewSession(innerConn)
	fmt.Println("sync...")
	cmdSync := NewCommand([]byte("SYNC"))
	sess.WriteCommand(cmdSync)

	for {
		var c byte
		var err error
		if c, err = sess.PeekByte(); err != nil {
			panic(err)
		}
		//fmt.Println("char:", string(c))
		if c == '*' {
			if cmd, e2 := sess.ReadCommand(); e2 != nil {
				panic(e2)
			} else {
				fmt.Println(cmd.Name())
			}
		} else if c == '$' {
			fmt.Println("skip rdb...")
			if e3 := sess.ReadRDB(); e3 != nil {
				panic(e3) 
			} else {
				fmt.Println("skip finish")
			}
		} else {
			panic(fmt.Sprintf("Bad first byte: %s", c))
			break
		}
	}
}
