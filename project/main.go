package main

// The main node to run elevator and controller

import (
	"flag"
	"fmt"
	"main/elev_algo_go/elevator"
	"main/elevio"
	"main/main_controller"
	"main/main_elevator"
	"main/utilities"
)

var (
	elev_to_ctrl_button_chan   = make(chan elevio.ButtonEvent)
	elev_to_ctrl_chan          = make(chan elevator.Elevator)
	ctrl_to_elev_chan          = make(chan [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool)
	ctrl_to_elev_cab_chan      = make(chan elevio.ButtonEvent)
	obstruction_timer_duration int
	controller_id              int
)

func _main() {
	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	if *debug {
		fmt.Println("Debug mode enabled")
		obstruction_timer_duration = 5
	} else {
		fmt.Println("Running in normal mode")
		obstruction_timer_duration = 30
	}

	fmt.Println("Starting controller and elevetor")

	go main_controller.Main_controller(controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan)
	go main_elevator.Main_elevator(obstruction_timer_duration, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan)
	for {
	}

}
