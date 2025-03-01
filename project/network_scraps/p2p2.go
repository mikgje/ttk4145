package main

import (
	"main/network/broadcast"
	"fmt"
	"net"
	"time"
)

func main() {
	myconn := broadcast.DialBroadcastUDP(33333)

	remote_addr, _ := net.ResolveUDPAddr("udp4", "255.255.255.255:22222")
	
	reply := make([]byte, 1000)
	for {
		fmt.Println("skal skrive")
		myconn.WriteTo([]byte("hei"), remote_addr)
		fmt.Println("skal lese")
		myconn.SetReadDeadline(time.Now().Add(time.Second))
		n,_,_ := myconn.ReadFrom(reply)
		fmt.Println("ferdig med å lese, n:",n)
		fmt.Println(string(reply))
	}
}
