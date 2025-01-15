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
	localAddr, err := net.ResolveUDPAddr("udp", "10.100.23.16:0")
	if err != nil {
		fmt.Println("Error resolving local address: ", err)
		return
	}
	conn, err := net.DialUDP("udp", localAddr, serverAddr)
	if err != nil {
		fmt.Println("Error creating socket: ", err)
		return
	}
	readConn, readErr := net.ListenUDP("udp", serverAddr)
	if readErr != nil {
		fmt.Println("Error creating listen socket: ", readErr)
		return
	}
	defer conn.Close()
	locAddr := conn.LocalAddr()
	remAddr := conn.RemoteAddr()
	fmt.Println("Local: ", locAddr, "Remote: ", remAddr)
	data := "Test UDP desk 6"
	fmt.Println("Buffer: ", data)
	_, err = conn.Write([]byte(data))
	if err != nil {
		fmt.Println("Error writing: ", err)
		return
	}
	reply := make([]byte, len(data))
	_, err = readConn.Read(reply)
	if err != nil {
		fmt.Println("Error reading: ", err)
		return
	}
	fmt.Println("Print reply:\r\n")
	fmt.Println(string(reply))
}
