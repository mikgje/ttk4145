package single_elevator

import (
	// "main/elev_algo_go/elevator"
	"main/elev_algo_go/elevator"
	"main/elev_algo_go/fsm"
	"main/elev_algo_go/timer"
	"main/elevio"
	"main/utilities"

	// "time"
	"fmt"
)

var current_elevator elevator.Elevator

func Start(obstruction_timer_duration int,
	elev_to_ctrl_chan chan<- elevator.Elevator,
	elev_to_ctrl_button_chan chan<- elevio.ButtonEvent,
	ctrl_to_elev_chan <-chan [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool,
	ctrl_to_elev_cab_chan <-chan elevio.ButtonEvent) {

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
	door_timer_channel := make(chan bool)
	obstruction_timer_channel := make(chan bool)
	abort_timer_channel := make(chan bool)

	go elevio.PollButtons(button_channel)
	go elevio.PollFloorSensor(floor_channel)
	go elevio.PollObstructionSwitch(obstruction_channel)

	for {
		select {
		case button := <-button_channel:
			elev_to_ctrl_chan <- fsm.Fsm_return_elevator()
			elev_to_ctrl_button_chan <- button
			// fsm.Fsm_on_request_button_press(button.Floor, button.Button, door_timer_channel)

		case floor := <-floor_channel:
			fsm.Fsm_on_floor_arrival(floor, door_timer_channel)
			elev_to_ctrl_chan <- fsm.Fsm_return_elevator()

		case obstruction := <-obstruction_channel:
			is_elevator_obstructed = obstruction
			if obstruction {
				fmt.Println("Obstruction detected, starting timer")
				go timer.Obstruction_timer(obstruction_timer_duration, obstruction_timer_channel, abort_timer_channel) //Start watchdog for obstruction switch
			} else {
				if fsm.Elevator_cab.Behaviour == elevator.EB_Unhealthy { //Check if unhealthy and handle this only to prevent blocking in abort_timer_channel
					fsm.Elevator_cab.Behaviour = elevator.EB_Idle
					elevio.SetStopLamp(false)
					elev_to_ctrl_chan <- fsm.Fsm_return_elevator()
				} else {
					fmt.Println("Obstruction removed, stopping timer")
					abort_timer_channel <- false
				}
			}

		case <-obstruction_timer_channel: //The obstruction timer has fired, the elevator is inoperable and communicates this to the controller
			fsm.Elevator_cab.Behaviour = elevator.EB_Unhealthy
			elevio.SetStopLamp(true)
			elev_to_ctrl_chan <- fsm.Fsm_return_elevator()

		case <-door_timer_channel:
			// fmt.Println("Status: ", is_elevator_healthy)
			if !(is_elevator_obstructed) {
				fsm.Fsm_on_door_timeout(door_timer_channel)
			} else {
				// fmt.Println("Starting timer again")
				go timer.Timer_start2(3, door_timer_channel)
			}

		case msg := <-ctrl_to_elev_chan:
			fsm.Fsm_overwrite_hall_orders(msg)
		case msg := <-ctrl_to_elev_cab_chan:
			fsm.Fsm_on_request_button_press(msg.Floor, msg.Button, door_timer_channel)
		default:

		}
	}
}
