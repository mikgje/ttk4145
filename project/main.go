package main

// The main node to run elevator and controller

import (
	"flag"
	"fmt"
	"main/elevio"
)

var (
	elev_to_ctrl_chan          = make(chan elevio.ButtonEvent)
	ctrl_to_elev_chan          = make(chan elevio.ButtonEvent)
	obstruction_timer_duration int
)

func main() {
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

	go main_controller()
	go main_elevator()
	for {
	}

}
