package fsm

import (
	"fmt"
	"elevator"
	"requests"
	"timer"
	"elevio"
)

var elevator_cab elevator.Elevator

func fsm_init() {
	elevator_cab = elevator.Elevator_uninitialised();
}

func fsm_set_all_lights(es elevator.Elevator) {
	for floor:=0; floor < elevator.N_FLOORS; floor++ {
		for btn:=0; btn < elevator.N_BUTTONS; btn++ {
			// outputDevice.requestButtonLight(floor, btn, es.Requests[floor][btn])
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, bool(es.Requests[floor][btn]))
		}
	}
}

func  fsm_on_init_between_floors() {
	// outputDevice.motorDirection(elevio.MD_Down)
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator_cab.Dirn = elevio.MD_Down
	elevator_cab.Behaviour = elevator.EB_Moving
}

func fsm_on_request_button_press(btn_floor int, btn_type elevio.ButtonType) {
	elevator.Elevator_print(elevator_cab)

	switch elevator_cab.Behaviour {
	case elevator.EB_DoorOpen:
		if requests.Requests_should_clear_immediately(elevator_cab, btn_floor, btn_type) {
			timer.Timer_start(elevator_cab.Config.DoorOpenDuration_s)
		} else {
			elevator_cab = requests.Requests_clear_at_current_floor(elevator_cab)
		}
		break
	
	case elevator.EB_Moving:
		elevator_cab.Requests[btn_floor][btn_type] = true
		break
	
	case elevator.EB_Idle:
		elevator_cab.Requests[btn_floor][btn_type] = true
		pair := requests.Requests_choose_direction(elevator_cab)
		elevator_cab.Dirn = pair.Dirn
		elevator_cab.Behaviour = pair.Behaviour
		switch pair.Behaviour {
		case elevator.EB_DoorOpen:
			// outputDevice.doorLight(1)
			elevio.SetDoorOpenLamp(true)
			timer.Timer_start(elevator_cab.Config.DoorOpenDuration_s)
			elevator_cab = requests.Requests_clear_at_current_floor(elevator_cab)
			break

		case elevator.EB_Moving:
			// outputDevice.motorDirection(elevator_cab.Dirn)
			elevio.SetMotorDirection(elevator_cab.Dirn)
			break

		case elevator.EB_Idle:
			break
		}
	}

	fsm_set_all_lights(elevator_cab)

	fmt.Println("\nNew state:")
	elevator.Elevator_print(elevator_cab)
}

func fsm_on_floor_arrival(new_floor int) {
	elevator.Elevator_print(elevator_cab)
	elevator_cab.Floor = new_floor
	// outputDevice.floorIndicator(elevator_cab.Floor)
	elevio.SetFloorIndicator(elevator_cab.Floor)
	
	switch elevator_cab.Behaviour {
	case elevator.EB_Moving:
		if requests.Requests_should_stop(elevator_cab) {
			// outputDevice.motorDirection(elevio.MD_Stop)
			elevio.SetMotorDirection(elevio.MD_Stop)
			// outputDevice.doorLight(1)
			elevio.SetDoorOpenLamp(true)
			elevator_cab = requests.Requests_clear_at_current_floor(elevator_cab)
			timer.Timer_start(elevator_cab.Config.DoorOpenDuration_s)
			fsm_set_all_lights(elevator_cab)
			elevator_cab.Behaviour = elevator.EB_DoorOpen
		}
		break
	default:
		break;
	}

	fmt.Println("\nNew state:")
	elevator.Elevator_print(elevator_cab)
}

func fsm_on_door_timeout() {
	elevator.Elevator_print(elevator_cab)

	switch elevator_cab.Behaviour {
	case elevator.EB_DoorOpen:
		pair := requests.Requests_choose_direction(elevator_cab)
		elevator_cab.Dirn = pair.Dirn
		elevator_cab.Behaviour = pair.Behaviour

		switch elevator_cab.Behaviour {
		case elevator.EB_DoorOpen:
			timer.Timer_start(elevator_cab.Config.DoorOpenDuration_s)
			elevator_cab = requests.Requests_clear_at_current_floor(elevator_cab)
			fsm_set_all_lights(elevator_cab)
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

