package main
// The main node to run elevator and controller

import (
	"fmt"
	"main/elevio"
)


var(
	elev_to_ctrl_chan = make(chan elevio.ButtonEvent)
	ctrl_to_elev_chan = make(chan elevio.ButtonEvent)
)


func main(){

	fmt.Println("Starting controller and elevetor")

	go main_controller()
	go main_elevator()
	for {}

}