package fsm

import (
	"fmt"
	"main/elev_algo_go/elevator"
	"main/elev_algo_go/requests_elev"
	"main/elev_algo_go/timer"
	"main/elevio"
)

var elevator_cab elevator.Elevator

func Fsm_init() {
	elevator_cab = elevator.Elevator_uninitialised()
}

func Fsm_set_all_lights(es elevator.Elevator) {
	for floor := 0; floor < elevator.N_FLOORS; floor++ {
		for btn := 0; btn < elevator.N_BUTTONS; btn++ {
			// outputDevice.requestButtonLight(floor, btn, es.Requests[floor][btn])
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, bool(es.Requests[floor][btn]))
		}
	}
}

func Fsm_on_init_between_floors() {
	// outputDevice.motorDirection(ecom/mikgje/ttk4145/elevio"levio.MD_Down)
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator_cab.Dirn = elevio.MD_Down
	elevator_cab.Behaviour = elevator.EB_Moving
}

func Fsm_on_request_button_press(btn_floor int, btn_type elevio.ButtonType, timer_channel chan<- bool) {
	elevator.Elevator_print(elevator_cab)

	switch elevator_cab.Behaviour {
	case elevator.EB_DoorOpen:
		if requests_elev.Requests_should_clear_immediately(elevator_cab, btn_floor, btn_type) {
			go timer.Timer_start2(elevator_cab.Config.DoorOpenDuration_s, timer_channel)
		} else {
			elevator_cab.Requests[btn_floor][btn_type] = true
		}
		break

	case elevator.EB_Moving:
		elevator_cab.Requests[btn_floor][btn_type] = true
		break

	case elevator.EB_Idle:
		elevator_cab.Requests[btn_floor][btn_type] = true
		pair := requests_elev.Requests_choose_direction(elevator_cab)
		elevator_cab.Dirn = pair.Dirn
		elevator_cab.Behaviour = pair.Behaviour
		switch pair.Behaviour {
		case elevator.EB_DoorOpen:
			// outputDevice.doorLight(1)
			elevio.SetDoorOpenLamp(true)
			go timer.Timer_start2(elevator_cab.Config.DoorOpenDuration_s, timer_channel)
			elevator_cab = requests_elev.Requests_clear_at_current_floor(elevator_cab)
			break

		case elevator.EB_Moving:
			// outputDevice.motorDirection(elevator_cab.Dirn)
			elevio.SetMotorDirection(elevator_cab.Dirn)
			break

		case elevator.EB_Idle:
			break
		}
	}

	Fsm_set_all_lights(elevator_cab)

	fmt.Println("\nNew state:")
	elevator.Elevator_print(elevator_cab)
}

func Fsm_on_floor_arrival(new_floor int, timer_channel chan<- bool) {
	elevator.Elevator_print(elevator_cab)
	elevator_cab.Floor = new_floor
	// outputDevice.floorIndicator(elevator_cab.Floor)
	elevio.SetFloorIndicator(elevator_cab.Floor)

	switch elevator_cab.Behaviour {
	case elevator.EB_Moving:
		if requests_elev.Requests_should_stop(elevator_cab) {
			// outputDevice.motorDirection(elevio.MD_Stop)
			elevio.SetMotorDirection(elevio.MD_Stop)
			// outputDevice.doorLight(1)
			elevio.SetDoorOpenLamp(true)
			elevator_cab = requests_elev.Requests_clear_at_current_floor(elevator_cab)
			go timer.Timer_start2(elevator_cab.Config.DoorOpenDuration_s, timer_channel)
			Fsm_set_all_lights(elevator_cab)
			elevator_cab.Behaviour = elevator.EB_DoorOpen
		}
		break
	default:
		break
	}

	fmt.Println("\nNew state:")
	elevator.Elevator_print(elevator_cab)
}

func Fsm_on_door_timeout(timer_channel chan<- bool) {
	elevator.Elevator_print(elevator_cab)

	elevio.SetDoorOpenLamp(false) // for å cleare lampen ved timeout

	switch elevator_cab.Behaviour {
	case elevator.EB_DoorOpen:
		pair := requests_elev.Requests_choose_direction(elevator_cab)
		elevator_cab.Dirn = pair.Dirn
		elevator_cab.Behaviour = pair.Behaviour

		switch elevator_cab.Behaviour {
		case elevator.EB_DoorOpen:
			go timer.Timer_start2(elevator_cab.Config.DoorOpenDuration_s, timer_channel)
			elevator_cab = requests_elev.Requests_clear_at_current_floor(elevator_cab)
			Fsm_set_all_lights(elevator_cab)
			break
		case elevator.EB_Moving:
			break
		case elevator.EB_Idle:
			// outputDevice.doorLight(0)
			elevio.SetDoorOpenLamp(false)
			// outputDevice.motorDirection(elevator.Dirn)
			elevio.SetMotorDirection(elevator_cab.Dirn)
			break
		}

		break
	default:
		break
	}

	fmt.Print("\nNew state:")
	elevator.Elevator_print(elevator_cab)
}
