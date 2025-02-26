package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
//	server_addr, err := net.ResolveUDPAddr("udp", "10.149.224.185:25555")
	server_addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:25555")
	error_check(err)

	conn, err := net.DialUDP("udp", nil, server_addr)
	error_check(err)
	fmt.Println("Client address:", conn.LocalAddr())

	for {
		_, err = conn.Write([]byte(fmt.Sprintf("%s", conn.LocalAddr())))
		error_check(err)
		time.Sleep(time.Second)
	}
}

func error_check(err error) {
	if err != nil {
		panic(fmt.Sprintf("Error: %v", err))
	}
}
