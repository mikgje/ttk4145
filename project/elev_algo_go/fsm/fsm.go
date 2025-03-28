package fsm

import (
	"main/elev_algo_go/elevator"
	"main/elev_algo_go/requests_elev"
	"main/elev_algo_go/timer"
	"main/elevio"
	"main/utilities"
)

var Elevator_cab elevator.Elevator
var kill_timer_channel = make(chan bool)

func Init_fsm() {
	Elevator_cab = elevator.Uninitialised_elevator()
	utilities.Load_cab_calls(&Elevator_cab.Requests, elevio.BT_Cab, utilities.Cab_calls_file_name)
	Set_all_lights(Elevator_cab)
	elevio.SetDoorOpenLamp(false)
}

func Return_elevator() elevator.Elevator {
	return Elevator_cab
}

func Overwrite_hall_orders(orders [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool, door_timer_chan chan<- bool, elevator_stuck_chan chan bool, kill_stuck_timer_chan chan bool) {
	for floor := 0; floor < utilities.N_FLOORS; floor++ {
		for btn := 0; btn < utilities.N_BUTTONS-1; btn++ {
			if orders[floor][btn] {
				On_request_button_press(floor, elevio.ButtonType(btn), door_timer_chan, elevator_stuck_chan, kill_stuck_timer_chan)
			} else {
				Elevator_cab.Requests[floor][btn] = false
			}
		}
	}
}

func Set_all_lights(es elevator.Elevator) {
	// Set lights for local service orders
	for floor := 0; floor < utilities.N_FLOORS; floor++ {
		for btn := 0; btn < utilities.N_BUTTONS; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, es.Requests[floor][btn])
		}
	}
	// Add lights for all service orders on network
	for _, orderlines := range es.Other_orderlines {
		for floor := 0; floor < utilities.N_FLOORS; floor++ {
			for btn := 0; btn < utilities.N_BUTTONS-1; btn++ {
				if es.Requests[floor][btn] || orderlines[floor][btn] {
					elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
				}
			}
		}
	}
}

func Set_other_orderlines(other_orderlines [][utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool) {
	Elevator_cab.Other_orderlines = other_orderlines
	Set_all_lights(Elevator_cab)
}

func On_init_between_floors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	Elevator_cab.Dirn = elevio.MD_Down
	Elevator_cab.Behaviour = elevator.EB_Moving
}

func On_request_button_press(btn_floor int, btn_type elevio.ButtonType, door_timer_chan chan<- bool, elevator_stuck_chan chan bool, kill_stuck_timer_chan chan bool) {
	switch Elevator_cab.Behaviour {
	case elevator.EB_DoorOpen:
		if requests_elev.Requests_should_clear_immediately(Elevator_cab, btn_floor, btn_type) {
			go timer.Timer_start(Elevator_cab.Config.Door_open_duration_s, door_timer_chan, kill_timer_channel)
		} else {
			Elevator_cab.Requests[btn_floor][btn_type] = true
			if btn_type == elevio.BT_Cab {
				utilities.Save_cab_calls(Elevator_cab.Requests,elevio.BT_Cab,utilities.Cab_calls_file_name)
			}
		}
	case elevator.EB_Moving:
		Elevator_cab.Requests[btn_floor][btn_type] = true
		if btn_type == elevio.BT_Cab {
			utilities.Save_cab_calls(Elevator_cab.Requests,elevio.BT_Cab,utilities.Cab_calls_file_name)
		}
	case elevator.EB_Idle:
		Elevator_cab.Requests[btn_floor][btn_type] = true
		if btn_type == elevio.BT_Cab {
			utilities.Save_cab_calls(Elevator_cab.Requests,elevio.BT_Cab,utilities.Cab_calls_file_name)
		}
		pair := requests_elev.Requests_choose_direction(Elevator_cab)
		Elevator_cab.Dirn = pair.Dirn
		Elevator_cab.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			go timer.Timer_start(Elevator_cab.Config.Door_open_duration_s, door_timer_chan, kill_timer_channel)
			Elevator_cab = requests_elev.Requests_clear_at_current_floor(Elevator_cab)
		case elevator.EB_Moving:
			elevio.SetMotorDirection(Elevator_cab.Dirn)
			Start_stuck_check(elevator_stuck_chan, kill_stuck_timer_chan)
		case elevator.EB_Idle:
		}
	}
	Set_all_lights(Elevator_cab)
}

func On_floor_arrival(new_floor int, door_timer_chan chan<- bool, kill_stuck_timer_chan chan bool) {

	select{
		case kill_stuck_timer_chan <- true:
		default:
			if Elevator_cab.Behaviour == elevator.EB_Obstructed {
				On_init_between_floors()
				elevio.SetStopLamp(false)
			}
	}

    Elevator_cab.Floor = new_floor
    elevio.SetFloorIndicator(Elevator_cab.Floor)
    switch Elevator_cab.Behaviour {
    case elevator.EB_Moving:
        if requests_elev.Requests_should_stop(Elevator_cab) {
            elevio.SetMotorDirection(elevio.MD_Stop)
            elevio.SetDoorOpenLamp(true)
            Elevator_cab = requests_elev.Requests_clear_at_current_floor(Elevator_cab)
			utilities.Save_cab_calls(Elevator_cab.Requests, elevio.BT_Cab, utilities.Cab_calls_file_name)            
            go timer.Timer_start(Elevator_cab.Config.Door_open_duration_s, door_timer_chan, kill_timer_channel)
            Set_all_lights(Elevator_cab)
            Elevator_cab.Behaviour = elevator.EB_DoorOpen
        }
	case elevator.EB_Idle:
		elevio.SetDoorOpenLamp(true)
		Elevator_cab = requests_elev.Requests_clear_at_current_floor(Elevator_cab)
		go timer.Timer_start(Elevator_cab.Config.Door_open_duration_s, door_timer_chan, kill_timer_channel)
		Set_all_lights(Elevator_cab)
		Elevator_cab.Behaviour = elevator.EB_DoorOpen
    default:
    }
}

func On_door_timeout(door_timer_chan chan<- bool, elevator_stuck_chan chan bool, kill_stuck_timer_chan chan bool) {
	switch Elevator_cab.Behaviour {
	case elevator.EB_DoorOpen:
		pair := requests_elev.Requests_choose_direction(Elevator_cab)
		Elevator_cab.Dirn = pair.Dirn
		Elevator_cab.Behaviour = pair.Behaviour

		switch Elevator_cab.Behaviour {
		case elevator.EB_DoorOpen:
			go timer.Timer_start(Elevator_cab.Config.Door_open_duration_s, door_timer_chan, kill_timer_channel)
			Elevator_cab = requests_elev.Requests_clear_at_current_floor(Elevator_cab)
			Set_all_lights(Elevator_cab)
		case elevator.EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(Elevator_cab.Dirn)
			Start_stuck_check(elevator_stuck_chan, kill_stuck_timer_chan)
		case elevator.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(Elevator_cab.Dirn)
			Start_stuck_check(elevator_stuck_chan, kill_stuck_timer_chan)
		}
	default:
		break
	}
}

func Start_stuck_check(elevator_stuck_chan chan<- bool, kill_stuck_timer_chan chan bool) {
	if Elevator_cab.Dirn != elevio.MD_Stop {
		go timer.Timer_start(float64(utilities.Obstruction_timer_duration), elevator_stuck_chan, kill_stuck_timer_chan)
	}
}