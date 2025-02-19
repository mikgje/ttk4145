package main

import (
	"fmt"
	"os/exec"
	"time"
	"net"
	"sync"
	"strings"
	"strconv"
)

var timer_end time.Time
var timer time.Time
var timeout float64
var duration float64
var reply []byte
var wg sync.WaitGroup
var counter int


func main() {
	fmt.Println("Backup has been launched")

	duration = 15
	timeout = 3
	timer = time.Now()

	my_addr, err := net.ResolveUDPAddr("udp", "localhost:22222")
	error_check(err)
	read_conn, err := net.ListenUDP("udp", my_addr)
	error_check(err)

	fmt.Println("Listen started")

	timer_end = time.Now().Add(time.Duration(duration)*time.Second)

	quit := make(chan bool)
	wg.Add(1)
	go receive(read_conn, quit)

	for {	
		if(timer.Add(time.Duration(timeout)*time.Second).Compare(time.Now()) != 1) {
			fmt.Println("Timeout")
			break
		}

	}
	quit <- true
	wg.Wait()
		
	fmt.Println("Assuming primary role")

	_, err = exec.Command("go", "run", "primary.go", strconv.Itoa(counter+1)).Output()
	error_check(err)
}

func error_check(err error) {
	if err != nil {
		panic(fmt.Sprintf("Error: ", err))
	}
}

func receive(read_conn net.Conn, quit chan bool) {
	reply := make([]byte, 1000)

	for {
		select {
		case <- quit:
			read_conn.Close()
			wg.Done()
			return
		default:
			_, err := read_conn.Read(reply) 
			read_conn.SetReadDeadline(time.Now().Add(3*time.Second))
			sreply := string(reply)
			counter,_ = strconv.Atoi(sreply[:strings.IndexRune(sreply,'␝')])
			if err == nil {
				fmt.Println("Received:", string(reply))
			}
			timer = time.Now()
		}
		time.Sleep(2*time.Second)
	}
}
