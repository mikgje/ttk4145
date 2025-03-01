package main

import (
	"main/network/broadcast"
	"fmt"
	"net"
	"time"
	"os"
	"strconv"
)

/* Message format
 * [LOCAL_ID : STATUS : TARGET_ID : INSTRUCTION]
 *
 */

var timer_end time.Time
var read_err error

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please specify a local port and a name")
		return
	}

	port := os.Args[1]
	name := os.Args[2]

	port_int, _ := strconv.Atoi(port)
	myconn := broadcast.DialBroadcastUDP(port_int)
	remote_addr, _ := net.ResolveUDPAddr("udp", "255.255.255.255:"+port)

	status := "slave"

	reply := make([]byte, 1000)

	timer_end = time.Now()

	for {
		message := name+":"+status
		myconn.SetReadDeadline(time.Now().Add(time.Second))
		_, write_err := myconn.WriteTo([]byte(message), remote_addr)
		if write_err != nil {
			fmt.Println("I lost connection")
		}
//		error_check(err)
//		_, _, read_err := myconn.ReadFrom(reply) // read_err on timeout/deadline
		read_err = nil
		for read_err == nil { // read all available messages _, _, read_err =  myconn.ReadFrom(reply) // NB: ReadFrom only overwrites the size of the received messages, it will not change characters outisde that
			fmt.Println(string(reply))
				
			if name != string(reply[:len(name)]) {
				timer_end = time.Now()
			}
		}
		if time.Now().Compare(timer_end.Add(3*time.Second)) != -1 { // time out if time since last message from node is > 3s
//			fmt.Println("I timed out")
			status = "master"
		}

		if name == "a" {
//		fmt.Println("Siste melding jeg sendte:",message)
		}
		time.Sleep(1000*time.Millisecond)
	}
}

//func receiver(message []byte) {
//	remote_id := 


func error_check(err error) {
	if err != nil {
		panic(fmt.Sprintf("Error: %v", err))
	}
}
