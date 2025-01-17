package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	// Dial (connect) to the server
	dialConn, err := net.Dial("tcp", "10.100.23.204:34933")
	if err != nil {
		fmt.Println("Error dialing: ", err)
		return
	}

	// Listen (create a local server)
	ln, err := net.Listen("tcp", "10.100.23.16:7887")
	_ = ln
	if err != nil {
		fmt.Println("Error listening: ", err)
		return
	}
	fmt.Println("I am listening on: ", ln.Addr())

	// Instruct the remote server to connect to the local server
	_, err = dialConn.Write([]byte("Connect to: 10.100.23.16:7887\000"))
	if err != nil {
		fmt.Println("Error writing: ", err)
		return
	}

	// Wait for the remote server to connect to the local server
	listenConn, err := ln.Accept()
	_ = listenConn
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	// Read from and write to the remote server
	go read(dialConn)
	go send(dialConn)

	
	// Read from and write to the local server
	go read(listenConn)
	go send(listenConn)

	select {}
}

func read(conn net.Conn) {
	for {
		reply := make([]byte, 1024)
		conn.Read(reply)
		fmt.Println("Reply: ", string(reply))
		time.Sleep(time.Second)
	}
}

func send(conn net.Conn) {
	for {
		_, err := conn.Write([]byte("Test\000"))
		if err != nil {
			fmt.Println("Error writing: ", err)
			return
		}
	}
}
