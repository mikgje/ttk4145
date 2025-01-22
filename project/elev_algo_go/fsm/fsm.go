package fsm

import (
	"fmt"
	"elevator"
	"requests"
	"timer"
)

func fsm_init() {
	elevator = elevator.elevator_unitialized();
}

func fsm_set_all_lights(Elevator es) {
	for floor:=0; floor < N_FLOORS; floor++ {
		for btn:=0; btn < N_BUTTONS; btn++ {
			outputDevice.requestButtonLight(floor, btn, es.requests[floor][btn])
		}
	}
}

func  fsm_on_init_between_floors() {
	outputDevice.motorDirection(D_Down)
	elevator.dirn = D_Down
	elevator.behaviour = EB_Moving
}

func fsm_on_request_button_press(btn_floor int, btn_type Button) {
	elevaotr.elevator_print(elevator)

	switch elevator.behaviour {
	case EB_DoorOpen:
		if requests_should_clear_immediately(elevator, btn_floor, btn_type) {
			timer_start(elevator.config.doorOpenDuration_s)
		} else {
			elevator = requests_clear_at_current_floor(elevator)
		}
		break
	
	case EB_Moving:
		elevator.requests[btn_floor][btn_type] = 1
		break
	
	case EB_Idle:
		elevator.requests[btn_floor][btn_type] = 1
		DirnBehaviourPair pair = requests_choose_direction(elevator)
		elevator.dirn = pair.dirn
		elevator.behaviour = pair.behaviour
		switch pair.behaviour {
		case EB_DoorOpen:
			outputDevice.doorLight(1)
			timer_start(elevator.config.doorOpenDuration_s)
			elevator = requests_clear_at_current_floor(elevator)
			break

		case EB_Moving:
			outputDevice.motorDirection(elevator.dirn)
			break

		case EB_Idle:
			break
		}
	}

	fsm_set_all_lights(elevator)

	fsm.Println("\nNew state:\n")
	elevator_print(elevator)
}

func fsm_on_floor_arrival(new_floor int) {
	elevator_print(elevator)
	elevator.floor = new_floor
	outputDevice.floorIndicator(elevator.floor)
	
	switch elevator.behaviour {
	case EB_Moving:
		if requests_should_stop(elevator) {
			outputDevice.motorDirection(D_Stop)
			outputDevice.doorLight(1)
			elevator = requests_clear_at_current_floor(elevator)
			timer_start(elevator.config.doorOpenDuration_s)
			fsm_set_all_lights(elevator)
			elevator.behaviour = EB_DoorOpen
		}
		break
	default:
		break;
	}

	fmt.Println("\nNew state:\n")
	elevator_print(elevator)
}

func fsm_on_door_timeout() {
	elevator_print(elevator)

	switch elevator.behaviour {
	case EB_DoorOpen:
		DirnBehaviourPair pair = requests_choose_direction(elevator)
		elevator.dirn = pair.dirn
		elevator.behaviour = pair.behaviour

		switch elevator.behaviour {
		case EB_DoorOpen:
			timer_start(elevator.config.doorOpenDuration_s)
			elevator = requests_clear_at_current_floor(elevator)
			fsm_set_all_lights(elevator)
			break
		case EB_Moving:
			break
		case EB_Idle:
			outputDevice.doorLight(0)
			outputDevice.motorDirection(elevator.dirn)
			break
		}

		break
	default:
		break
	}

	fmt.Print("\nNew state:\n")
	elevator_print(elevator)
}

