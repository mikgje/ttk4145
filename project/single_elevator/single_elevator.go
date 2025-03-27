package single_elevator

/*-------------------------------------*/
// INPUT:
// hall_orders_chan: Hall orders from master controller on network
// cab_orders_chan: Cab orders from elevator via controller
/*-------------------------------------*/
// OUTPUT:
// elevator_status_chan: Elevator status to controller (floor, direction, behaviour, requests)
// button_event_chan: Button events from elevator (floor, button type) to controller
/*-------------------------------------*/
import (
	"main/elev_algo_go/elevator"
	"main/elev_algo_go/fsm"
	"main/elev_algo_go/timer"
	"main/elevio"
	"main/utilities"

	"fmt"
	"time"
)

func Run_single_elevator(
	elevator_status_chan chan<- elevator.Elevator,
	button_event_chan chan<- elevio.ButtonEvent,
	hall_orders_chan <-chan utilities.ControllerToElevatorMessage,
	cab_orders_chan <-chan elevio.ButtonEvent,
) {

	var elevator_obstructed    bool = false
	
	floor_chan                  := make(chan int)
	button_io_chan              := make(chan elevio.ButtonEvent)
	obstruction_io_chan         := make(chan bool)
	door_timer_chan             := make(chan bool)
	obstruction_timer_chan      := make(chan bool)
	abort_timer_chan            := make(chan bool)

	elevio.Init(fmt.Sprintf("localhost:%d", *utilities.Elevio), utilities.N_FLOORS)
	fsm.Init_fsm()

	if elevio.GetFloor() == -1 {
		fsm.On_init_between_floors()
	} else {
		fsm.On_floor_arrival(elevio.GetFloor(), door_timer_chan)
	}

	go elevio.PollButtons(button_io_chan)
	go elevio.PollFloorSensor(floor_chan)
	go elevio.PollObstructionSwitch(obstruction_io_chan)

	// Periodically update controller with elevator status
	go func() {
		for {
			elevator_status_chan <- fsm.Return_elevator()
			time.Sleep(utilities.Elevator_update_rate_ms)
		}
	}()

	for {
		select {
		case button := <-button_io_chan:
			button_event_chan <- button
		case floor := <-floor_chan:
			fsm.On_floor_arrival(floor, door_timer_chan)
		case obstruction := <-obstruction_io_chan:
			elevator_obstructed = obstruction
			if elevator_obstructed {
				fmt.Println("Obstruction detected, starting timer")
				go timer.Obstruction_timer(utilities.Obstruction_timer_duration, obstruction_timer_chan, abort_timer_chan)
			} else {
				if fsm.Elevator_cab.Behaviour == elevator.EB_Obstructed { //Check if unhealthy and handle this only to prevent blocking in abort_timer_chan
					fsm.Elevator_cab.Behaviour = elevator.EB_Idle
					elevio.SetStopLamp(false)
				} else {
					fmt.Println("Obstruction removed, stopping timer")
					abort_timer_chan <- false
				}
			}
		case <-obstruction_timer_chan: //The obstruction timer has fired, the elevator is inoperable and communicates this to the controller
			fsm.Elevator_cab.Behaviour = elevator.EB_Obstructed
			elevio.SetStopLamp(true)
		case <-door_timer_chan:
			if !(elevator_obstructed) {
				fsm.Fsm_on_door_timeout(door_timer_chan)
			} else {
				go timer.Timer_start(3, door_timer_chan, nil)
			}
		case msg := <-hall_orders_chan:
			fsm.Overwrite_hall_orders(msg.Orderline, door_timer_chan)
			fsm.Set_other_orderlines(msg.Other_orderlines)
		case msg := <-cab_orders_chan:
			fsm.On_request_button_press(msg.Floor, msg.Button, door_timer_chan)
		}
	}
}
