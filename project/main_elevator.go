package main

import (
	// "main/elev_algo_go/elevator"
	"main/elev_algo_go/fsm"
	"main/elev_algo_go/timer"
	"main/elevio"

	// "time"
	"fmt"
)

var healthy bool = true

const UNHEALTHY_FLAG int = -1
const HEALTHY_FLAG int = 100

func main_elevator() {

	// elevio.SetStopLamp(false)
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
	obstruction_timer_channel := make(chan bool)
	abort_timer_channel := make(chan bool)

	go elevio.PollButtons(button_channel)
	go elevio.PollFloorSensor(floor_channel)
	go elevio.PollObstructionSwitch(obstruction_channel)

	for {
		select {
		case button := <-button_channel:
			elev_to_ctrl_chan <- button
			// fsm.Fsm_on_request_button_press(button.Floor, button.Button, timer_channel)

		case floor := <-floor_channel:
			fsm.Fsm_on_floor_arrival(floor, timer_channel)

		case obstruction := <-obstruction_channel:
			is_elevator_obstructed = obstruction
			if obstruction {
				fmt.Println("Obstruction detected, starting timer")
				go timer.Obstruction_timer(obstruction_timer_channel, abort_timer_channel) //Start watchdog for obstruction switch
			} else {
				if !healthy { //Check if unhealthy and handle this only to prevent blocking in abort_timer_channel
					healthy = true
					elevio.SetStopLamp(false)
					elev_to_ctrl_chan <- elevio.ButtonEvent{Floor: HEALTHY_FLAG, Button: elevio.BT_Cab}
				} else {
					fmt.Println("Obstruction removed, stopping timer")
					abort_timer_channel <- false
				}
			}

		case <-obstruction_timer_channel: //The obstruction timer has fired, the elevator is inoperable and communicates this to the controller
			healthy = false
			elevio.SetStopLamp(true)
			elev_to_ctrl_chan <- elevio.ButtonEvent{Floor: UNHEALTHY_FLAG, Button: elevio.BT_Cab}

		case <-timer_channel:
			// fmt.Println("Status: ", is_elevator_healthy)
			if !(is_elevator_obstructed) {
				fsm.Fsm_on_door_timeout(timer_channel)
			} else {
				// fmt.Println("Starting timer again")
				go timer.Timer_start2(3, timer_channel)
			}

		case msg := <-ctrl_to_elev_chan:
			fsm.Fsm_on_request_button_press(msg.Floor, msg.Button, timer_channel)

		}
	}
}
