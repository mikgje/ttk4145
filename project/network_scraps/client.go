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

	local_addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:20000")
	error_check(err)

	conn, err := net.DialUDP("udp", nil, server_addr)
	error_check(err)
	fmt.Println("Dial address:", conn.LocalAddr())

	read_conn, err := net.ListenUDP("udp", local_addr)
	error_check(err)
//	fmt.Println("Listen address:", read_conn.LocalAddr())

	reply := make([]byte, 1000)
	for {
		fmt.Println("skal skrive", ":20000")
		_, err = conn.Write([]byte(fmt.Sprintf("%s", ":20000")))
		error_check(err)
		time.Sleep(time.Second)

		fmt.Println("skal lese fra", conn.RemoteAddr())
		read_conn.SetReadDeadline(time.Now().Add(1*time.Second))
//		conn.ReadFromUDP(reply)
		n, _, _ := read_conn.ReadFromUDP(reply)
//		error_check(err)
		fmt.Println("On my address", conn.LocalAddr(), "Received:", string(reply[:n]))
	}
}

func error_check(err error) {
	if err != nil {
		panic(fmt.Sprintf("Error: %v", err))
	}
}

// FUNGERER IKKE Å BRUKE 255.255.255.255?? Dette gjelder både client og server
