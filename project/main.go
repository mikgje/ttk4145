package main

// The main node to run elevator and controller

import (
	"flag"
	"fmt"
	"main/controller"
	"main/elev_algo_go/elevator"
	"main/elevio"
	"main/single_elevator"
	"main/utilities"
)

var (
	button_event_chan    = make(chan elevio.ButtonEvent)
	elevator_status_chan = make(chan elevator.Elevator, 1)
	hall_orders_chan = make(chan utilities.ControllerToElevatorMessage, 1)
	cab_orders_chan   = make(chan elevio.ButtonEvent, 1)
)

func main() {
	flag.Parse()

	fmt.Println("Starting controller and elevetor")

	go controller.Start(elevator_status_chan, button_event_chan, hall_orders_chan, cab_orders_chan)
	go single_elevator.Run_single_elevator(elevator_status_chan, button_event_chan, hall_orders_chan, cab_orders_chan)
	for {}
}
