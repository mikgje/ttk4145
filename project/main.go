package main
// The main node to run elevator and controller

import (
	"fmt"
)

var(
	elev_to_ctrl_chan = make(chan string)
	ctrl_to_elev_chan = make(chan string)
)


func main(){

	fmt.Println("Starting controller and elevetor")

	go main_controller()
	go main_elevator()

}