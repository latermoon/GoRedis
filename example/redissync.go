package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	sync("localhost:6379")
}

func sync(host string) {
	conn, e1 := net.Dial("tcp", host)
	if e1 != nil {
		panic(e1)
	}
	reader := bufio.NewReader(conn)

	fmt.Println("SYNC...")
	conn.Write([]byte("SYNC\r\n"))

	for {
		c, err := reader.ReadByte()
		if err != nil {
			panic(err)
		}
		if c >= ' ' && c < 127 {
			fmt.Print(string(c))
		} else if c == '\r' {
			fmt.Print("\\r")
		} else if c == '\n' {
			fmt.Println("\\n")
		} else {
			//fmt.Printf("[%02X]", c)
			fmt.Printf("[%d]", c)
		}
	}

}
