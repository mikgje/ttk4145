package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
	"net"
	"strconv"
)

func main() {
	fmt.Println("Primary has been started")

	err := exec.Command("gnome-terminal", "--", "go", "run", "backup.go").Run()
	error_check(err)

	time.Sleep(3*time.Second)

	target_addr, err := net.ResolveUDPAddr("udp", "localhost:22222")
	error_check(err)

	write_conn, err := net.DialUDP("udp", nil, target_addr)
	error_check(err)

	fmt.Println("Local: ", write_conn.LocalAddr(), "Remote: ", write_conn.RemoteAddr())

	defer write_conn.Close()

	i := 0
	if len(os.Args[1:]) > 0 {
		i,_ = strconv.Atoi(os.Args[1])
	}

	for {
		_, err = write_conn.Write([]byte(fmt.Sprintf("%d%s", i, "␝")))
		error_check(err)
		i += 1
		time.Sleep(2*time.Second)
	}
}

func error_check(err error) {
	if err != nil {
		panic(fmt.Sprintf("Error: ", err))
	}
}
