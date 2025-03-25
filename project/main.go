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
	elev_to_ctrl_button_chan   = make(chan elevio.ButtonEvent)
	elev_to_ctrl_chan          = make(chan elevator.Elevator, 1)
	ctrl_to_elev_chan          = make(chan utilities.ControllerToElevatorMessage, 1)
	ctrl_to_elev_cab_chan      = make(chan elevio.ButtonEvent, 1)
	obstruction_timer_duration int
)

func main() {
	flag.Parse()

	if *utilities.Debug {
		fmt.Println("Debug mode enabled")
		obstruction_timer_duration = 5
	} else {
		fmt.Println("Running in normal mode")
		obstruction_timer_duration = 30
	}

	fmt.Println("Starting controller and elevetor")

	go controller.Start(elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan)
	go single_elevator.Start(obstruction_timer_duration, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan)
	for {
	}

}
