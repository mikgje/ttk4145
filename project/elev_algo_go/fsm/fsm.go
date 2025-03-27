package fsm

import (
	"main/elev_algo_go/elevator"
	"main/elev_algo_go/requests_elev"
	"main/elev_algo_go/timer"
	"main/elevio"
	"main/utilities"

	"fmt"
)

var Elevator_cab elevator.Elevator
var kill_timer_channel = make(chan bool)

func Fsm_return_elevator() elevator.Elevator {
	return Elevator_cab
}

func Fsm_overwrite_hall_orders(orders [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool, timer_channel chan<- bool) {
	elevator.Clear_elevator_requests(&Elevator_cab)
	for floor := 0; floor < utilities.N_FLOORS; floor++ {
		for btn := 0; btn < utilities.N_BUTTONS-1; btn++ {
			if orders[floor][btn] {
				Fsm_on_request_button_press(floor, elevio.ButtonType(btn), timer_channel)
			}
		}
	}
	// elevator.Elevator_print(Elevator_cab)
}

func Fsm_init() {
	Elevator_cab = elevator.Elevator_uninitialised()

	if err := utilities.Load_cab_calls(&Elevator_cab.Requests, elevio.BT_Cab, utilities.Cab_calls_file_name); err != nil {
		fmt.Println("Load error:", err)
	}
	
	Fsm_set_all_lights(Elevator_cab)
}

func Fsm_set_all_lights(es elevator.Elevator) {
	// Set lights for local calls
	for floor := 0; floor < utilities.N_FLOORS; floor++ {
		for btn := 0; btn < utilities.N_BUTTONS; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, es.Requests[floor][btn])
		}
	}

	for _, orderlines := range es.Other_orderlines {

		for floor := 0; floor < utilities.N_FLOORS; floor++ {
			for btn := 0; btn < utilities.N_BUTTONS-1; btn++ {
				if es.Requests[floor][btn] || orderlines[floor][btn] {

					elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
				} else {

				}
			}
		}
	}

}

func Fsm_set_other_orderlines(other_orderlines [][utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool) {
	Elevator_cab.Other_orderlines = other_orderlines
	Fsm_set_all_lights(Elevator_cab)
}

func Fsm_on_init_between_floors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	Elevator_cab.Dirn = elevio.MD_Down
	Elevator_cab.Behaviour = elevator.EB_Moving
}

func Fsm_on_request_button_press(btn_floor int, btn_type elevio.ButtonType, timer_channel chan<- bool) {
	switch Elevator_cab.Behaviour {
	case elevator.EB_DoorOpen:
		if requests_elev.Requests_should_clear_immediately(Elevator_cab, btn_floor, btn_type) {
			go timer.Timer_start(Elevator_cab.Config.DoorOpenDuration_s, timer_channel, kill_timer_channel)
		} else {
			Elevator_cab.Requests[btn_floor][btn_type] = true

			//____change____//
			// If it's a cab call, save the updated state (Requests) by passing the 3 args:
			if btn_type == elevio.BT_Cab {
				if err := utilities.Save_cab_calls(Elevator_cab.Requests,elevio.BT_Cab,utilities.Cab_calls_file_name,); err != nil {
					fmt.Println("Error saving cab calls:", err)
				}
			}
			//____change____//
		}

	case elevator.EB_Moving:
		Elevator_cab.Requests[btn_floor][btn_type] = true

		//____change____//
		// If it's a cab call, save the updated state (Requests).
		if btn_type == elevio.BT_Cab {
			if err := utilities.Save_cab_calls(Elevator_cab.Requests,elevio.BT_Cab,utilities.Cab_calls_file_name,); err != nil {
				fmt.Println("Error saving cab calls:", err)
			}
		}
		//____change____//

	case elevator.EB_Idle:
		Elevator_cab.Requests[btn_floor][btn_type] = true

		//____change____//
		// If it's a cab call, save the updated state (Requests).
		if btn_type == elevio.BT_Cab {
			if err := utilities.Save_cab_calls(Elevator_cab.Requests,elevio.BT_Cab,utilities.Cab_calls_file_name,); err != nil {
				fmt.Println("Error saving cab calls:", err)
			}
		}
		//____change____//

		pair := requests_elev.Requests_choose_direction(Elevator_cab)
		Elevator_cab.Dirn = pair.Dirn
		Elevator_cab.Behaviour = pair.Behaviour
		switch pair.Behaviour {
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			go timer.Timer_start(Elevator_cab.Config.DoorOpenDuration_s, timer_channel, kill_timer_channel)
			Elevator_cab = requests_elev.Requests_clear_at_current_floor(Elevator_cab)
		case elevator.EB_Moving:
			elevio.SetMotorDirection(Elevator_cab.Dirn)
		case elevator.EB_Idle:
			// Nothing additional to do.
		}
	}

	// Update all button lamps
	Fsm_set_all_lights(Elevator_cab)
}





func Fsm_on_floor_arrival(new_floor int, timer_channel chan<- bool) {
    Elevator_cab.Floor = new_floor
    elevio.SetFloorIndicator(Elevator_cab.Floor)

    switch Elevator_cab.Behaviour {
    case elevator.EB_Moving:
        if requests_elev.Requests_should_stop(Elevator_cab) {
            elevio.SetMotorDirection(elevio.MD_Stop)
            elevio.SetDoorOpenLamp(true)

            Elevator_cab = requests_elev.Requests_clear_at_current_floor(Elevator_cab)
            
            //____change____//
            // After clearing the requests for the current floor, save the updated cab calls (column = BT_Cab).
            if err := utilities.Save_cab_calls(Elevator_cab.Requests, elevio.BT_Cab, utilities.Cab_calls_file_name); err != nil {
                fmt.Println("Error saving cab calls:", err)
            }
            //____change____//
            
            go timer.Timer_start(Elevator_cab.Config.DoorOpenDuration_s, timer_channel, kill_timer_channel)
            Fsm_set_all_lights(Elevator_cab)
            Elevator_cab.Behaviour = elevator.EB_DoorOpen
        }
    default:
        // No-op for other states
    }

    // fmt.Println("\nNew state:")
    // elevator.Elevator_print(Elevator_cab)
}




func Fsm_on_door_timeout(timer_channel chan<- bool) {
	// elevator.Elevator_print(Elevator_cab)

	elevio.SetDoorOpenLamp(false) // for å cleare lampen ved timeout

	switch Elevator_cab.Behaviour {
	case elevator.EB_DoorOpen:
		pair := requests_elev.Requests_choose_direction(Elevator_cab)
		Elevator_cab.Dirn = pair.Dirn
		Elevator_cab.Behaviour = pair.Behaviour

		switch Elevator_cab.Behaviour {
		case elevator.EB_DoorOpen:
			go timer.Timer_start(Elevator_cab.Config.DoorOpenDuration_s, timer_channel, kill_timer_channel)
			Elevator_cab = requests_elev.Requests_clear_at_current_floor(Elevator_cab)
			Fsm_set_all_lights(Elevator_cab)
		case elevator.EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(Elevator_cab.Dirn)
		case elevator.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(Elevator_cab.Dirn)
		}

	default:
		break
	}

	// fmt.Print("\nNew state:")
	// elevator.Elevator_print(Elevator_cab)
}
