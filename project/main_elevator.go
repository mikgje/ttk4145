package main

import (
	// "main/elev_algo_go/elevator"
	"main/elev_algo_go/fsm"
	"main/elev_algo_go/timer"
	"main/elevio"
	// "time"
	// "fmt"
)

func main_elevator() {

	numFloors := 4

	is_elevator_obstructed := false

	elevio.Init("localhost:15657", numFloors)
	fsm.Fsm_init()

	if elevio.GetFloor() == -1 {
		fsm.Fsm_on_init_between_floors()
	}

	floor_channel := make(chan int)
	button_channel := make(chan elevio.ButtonEvent)
	obstruction_channel := make(chan bool)
	timer_channel := make(chan bool)
	
	go elevio.PollButtons(button_channel)
	go elevio.PollFloorSensor(floor_channel)
	go elevio.PollObstructionSwitch(obstruction_channel)

	for {
		select {
		case button := <- button_channel:
			fsm.Fsm_on_request_button_press(button.Floor, button.Button, timer_channel)
		case floor := <- floor_channel:
			fsm.Fsm_on_floor_arrival(floor, timer_channel)
		case obstruction := <- obstruction_channel:
			is_elevator_obstructed = obstruction
		case <- timer_channel:
			if !(is_elevator_obstructed) {
				fsm.Fsm_on_door_timeout(timer_channel)
			} else {
				go timer.Timer_start2(3, timer_channel)
			}
		}
	}
}
