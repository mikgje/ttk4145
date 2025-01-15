package main

import (
	"fmt"
	//"runtime"
	//"time"
	"net"
)

func main() {
	serverAddr, err := net.ResolveUDPAddr("udp", ":20006")
	if err != nil {
		fmt.Println("Error resolving server address: ", err)
		return
	}
	for {
		readConn, readErr := net.ListenUDP("udp", serverAddr)
		if readErr != nil {
			fmt.Println("Error creating listen socket: ", readErr)
			return
		}
		reply := make([]byte, 1024)
		_, err = readConn.Read(reply)
		if err != nil {
			fmt.Println("Error reading: ", err)
			return
		}
		fmt.Println("Print reply:\r\n")
		fmt.Println(string(reply))
	}
}
