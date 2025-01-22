// Use `go run foo.go` to run your program

package main

import (
    . "fmt"
    "runtime"
    //"time"
)

var i = 0

func incrementing(ch chan int, fin chan int) {
    //TODO: increment i 1000000 times
	for j:=0; j<1000000; j++ {
		ch <- 1
    }
    fin <- 0
}

func decrementing(ch chan int, fin chan int) {
    //TODO: decrement i 1000000 times
	for j:=0; j<1000000; j++ {
		ch <- 2
    }
	fin <- 0
}

func main() {
    // What does GOMAXPROCS do? What happens if you set it to 1?
    runtime.GOMAXPROCS(2)

	ch1 := make(chan int)
	ch2 := make(chan int)
	fin := make(chan int)
	
    // TODO: Spawn both functions as goroutines

	go func() {
		for {
		    select {
		    case <-ch1:
		    	i++
		    case <-ch2:
			    i--
	        }
        }
    }()
	go incrementing(ch1, fin)
	go decrementing(ch2, fin)

    // We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
    // We will do it properly with channels soon. For now: Sleep.
    //time.Sleep(500*time.Millisecond)
	for k := 0; k < 2; k++ {
		<-fin
    }
	Println("The magic number is:", i)
}
