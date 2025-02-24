package main

import (
	"main/network/broadcast"
	"fmt"
	"net"
	"time"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please specify a local port and a remote port")
		return
	}

	local_port := os.Args[1]
	remote_port := os.Args[2]

	local_port_int, _ := strconv.Atoi(local_port)
	myconn := broadcast.DialBroadcastUDP(local_port_int)
	remote_addr, _ := net.ResolveUDPAddr("udp", "255.255.255.255:"+remote_port)

	status := "slave"

	reply := make([]byte, 1000)
	for {
		message := status+myconn.LocalAddr().String()
		myconn.SetReadDeadline(time.Now().Add(time.Second))
		myconn.WriteTo([]byte(message), remote_addr)
		_, _, err := myconn.ReadFrom(reply)
		fmt.Println("My message:", message, "I received:", string(reply))

		if err != nil {
			status = "master"
		}

		time.Sleep(100*time.Millisecond)
	}
}
