package main

import (
	"fmt"
	"net"
)

func main() {
	dialConn, err := net.Dial("tcp", "10.100.23.204:34933")
	if err != nil {
		fmt.Println("Error dialing: ", err)
		return
	}

	go listen(dialConn)
	go send(dialConn)

	select {}
}

func listen(conn net.Conn) {
	reply := make([]byte, 1024)
	conn.Read(reply)
	fmt.Println("Reply: ", string(reply))
}

func send(conn net.Conn) {
	_, err := conn.Write([]byte("test\000"))
	if err != nil {
		fmt.Println("Error writing: ", err)
		return
	}
}
