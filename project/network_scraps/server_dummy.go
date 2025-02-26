package main

import (
	"fmt"
	"net"
)

func main() {
//	server_addr, err := net.ResolveUDPAddr("udp", "10.149.224.185:25555")
	server_addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:25555")
	error_check(err)

	conn, err := net.ListenUDP("udp", server_addr)
	error_check(err)

	reply := make([]byte, 1000)
	for {
		fmt.Println("Server will read")
		// Read incoming message
		_, _, err := conn.ReadFromUDP(reply)
		error_check(err)
		fmt.Println("I have received:", string(reply))
	}
}

func error_check(err error) {
	if err != nil {
		panic(fmt.Sprintf("Error: %v", err))
	}
}

/*
Må implementere logikk for å velge master
For dette vil man trenge en timeout på å lese broadcast
Kan vi ha en broadcast port, her 25555, og så dynamisk allokere resten
*/
