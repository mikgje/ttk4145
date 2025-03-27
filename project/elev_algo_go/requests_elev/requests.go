package requests_elev

import (
	"main/elev_algo_go/elevator"
	"main/elevio"
	"main/utilities"
)

type Dirn_behaviour_pair struct {
	Dirn      elevio.MotorDirection
	Behaviour elevator.Elevator_behaviour
}

func requests_above(e elevator.Elevator) bool {
	for f := e.Floor + 1; f < utilities.N_FLOORS; f++ {
		for btn := 0; btn < utilities.N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_below(e elevator.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < utilities.N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_here(e elevator.Elevator) bool {
	for btn := 0; btn < utilities.N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func Requests_choose_direction(e elevator.Elevator) Dirn_behaviour_pair {
	switch e.Dirn {
	case elevio.MD_Up:
		if requests_above(e) {
			return Dirn_behaviour_pair{elevio.MD_Up, elevator.EB_Moving}
		} else if requests_here(e) {
			return Dirn_behaviour_pair{elevio.MD_Down, elevator.EB_DoorOpen}
		} else if requests_below(e) {
			return Dirn_behaviour_pair{elevio.MD_Down, elevator.EB_Moving}
		} else {
			return Dirn_behaviour_pair{elevio.MD_Stop, elevator.EB_Idle}
		}

	case elevio.MD_Down:
		if requests_below(e) {
			return Dirn_behaviour_pair{elevio.MD_Down, elevator.EB_Moving}
		} else if requests_here(e) {
			return Dirn_behaviour_pair{elevio.MD_Up, elevator.EB_DoorOpen}
		} else if requests_above(e) {
			return Dirn_behaviour_pair{elevio.MD_Up, elevator.EB_Moving}
		} else {
			return Dirn_behaviour_pair{elevio.MD_Stop, elevator.EB_Idle}
		}

	case elevio.MD_Stop:
		if requests_here(e) {
			return Dirn_behaviour_pair{elevio.MD_Stop, elevator.EB_DoorOpen}
		} else if requests_above(e) {
			return Dirn_behaviour_pair{elevio.MD_Up, elevator.EB_Moving}
		} else if requests_below(e) {
			return Dirn_behaviour_pair{elevio.MD_Down, elevator.EB_Moving}
		} else {
			return Dirn_behaviour_pair{elevio.MD_Stop, elevator.EB_Idle}
		}

	default:
		return Dirn_behaviour_pair{elevio.MD_Stop, elevator.EB_Idle}
	}
}
func Requests_should_stop(e elevator.Elevator) bool {
	switch e.Dirn {
	case elevio.MD_Down:
		return e.Requests[e.Floor][elevio.BT_HallDown] ||
			e.Requests[e.Floor][elevio.BT_Cab] ||
			!requests_below(e)

	case elevio.MD_Up:
		return e.Requests[e.Floor][elevio.BT_HallUp] ||
			e.Requests[e.Floor][elevio.BT_Cab] ||
			!requests_above(e)

	case elevio.MD_Stop:
		fallthrough
	default:
		return true
	}
}

func Requests_should_clear_immediately(e elevator.Elevator, btnFloor int, btnType elevio.ButtonType) bool {
	switch e.Config.Clear_request_variant {
	case elevator.CV_All:
		return e.Floor == btnFloor

	case elevator.CV_InDirn:
		return e.Floor == btnFloor && ((e.Dirn == elevio.MD_Up && btnType == elevio.BT_HallUp) ||
			(e.Dirn == elevio.MD_Down && btnType == elevio.BT_HallDown) ||
			(e.Dirn == elevio.MD_Stop) ||
			(btnType == elevio.BT_Cab))

	default:
		return false
	}
}

func Requests_clear_at_current_floor(e elevator.Elevator) elevator.Elevator {
	switch e.Config.Clear_request_variant {
	case elevator.CV_All:
		for btn := 0; btn < utilities.N_BUTTONS; btn++ {
			e.Requests[e.Floor][btn] = false
		}

	case elevator.CV_InDirn:
		e.Requests[e.Floor][elevio.BT_Cab] = false

		switch e.Dirn {
		case elevio.MD_Up:
			if !requests_above(e) && !(e.Requests[e.Floor][elevio.BT_HallUp]) {
				e.Requests[e.Floor][elevio.BT_HallDown] = false
			}
			e.Requests[e.Floor][elevio.BT_HallUp] = false

		case elevio.MD_Down:
			if !requests_below(e) && !(e.Requests[e.Floor][elevio.BT_HallDown]) {
				e.Requests[e.Floor][elevio.BT_HallUp] = false
			}
			e.Requests[e.Floor][elevio.BT_HallDown] = false

		case elevio.MD_Stop:
			fallthrough
		default:
			e.Requests[e.Floor][elevio.BT_HallUp] = false
			e.Requests[e.Floor][elevio.BT_HallDown] = false
		}
	default:
	}

	return e
}
