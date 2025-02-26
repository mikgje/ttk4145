package main

import (
	"fmt"
	"net"
)

func main() {
	server_addr, err := net.ResolveUDPAddr("udp", ":25555")
//	server_addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:25555")
	error_check(err)

	conn, err := net.ListenUDP("udp", server_addr)
	error_check(err)

	reply := make([]byte, 1000)
	for {
		// Read incoming message
		n, _, err := conn.ReadFromUDP(reply)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			continue
		}

		fmt.Println("I received from client:", string(reply))

//		reply_addr, err := net.ResolveUDPAddr("udp", "10.149.224.185"+string(reply))
		reply_addr, err := net.ResolveUDPAddr("udp", "10.149.224.185:20000")
		error_check(err)

		// Reply to client
		fmt.Println("Server will attempt to reply to", reply_addr)
//		_, err = conn.WriteToUDP([]byte(fmt.Sprintf("Hei %s!", string(reply[:n]))), remote_addr)
		_, err = conn.WriteToUDP([]byte(fmt.Sprintf("Hei %s!", string(reply[:n]))), reply_addr)
		error_check(err)
		fmt.Println("Server has written to", reply_addr)
	}
}

func error_check(err error) {
	if err != nil {
		panic(fmt.Sprintf("Error: %v", err))
	}
}
