package elevator
// Necessary datatypes and functions for running a single elevator
import (
	"main/elevio"
	"main/utilities"
)

type Elevator_behaviour int
type Clear_request_variant int
type Direction int
type Button int

type Elevator struct {
	Floor     int
	Dirn      elevio.MotorDirection
	Requests  [utilities.N_FLOORS][utilities.N_BUTTONS]bool
	Other_orderlines [][utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool
	Behaviour Elevator_behaviour

	Config struct {
		Clear_request_variant Clear_request_variant
		Door_open_duration_s  float64
	}	
}

const (
	CV_All Clear_request_variant = iota
	CV_InDirn
)

const (
	EB_Idle Elevator_behaviour = iota
	EB_DoorOpen
	EB_Moving
	EB_Obstructed
	EB_Disconnected
)

var EB_to_string = map[Elevator_behaviour]string{
	EB_Idle:     "idle",
	EB_DoorOpen: "doorOpen",
	EB_Moving:   "moving",
	EB_Obstructed: "obstructed",
	EB_Disconnected: "disconnected",
}

var Dirn_to_string = map[elevio.MotorDirection]string{
	elevio.MD_Up:   "up",
	elevio.MD_Down: "down",
	elevio.MD_Stop: "stop",
}

var Button_to_string = map[elevio.ButtonType]string{
	elevio.BT_HallUp:   "B_hallUp",
	elevio.BT_HallDown: "B_hallDown",
	elevio.BT_Cab:      "B_cab",
}

var CV_to_string = map[Clear_request_variant]string{
	CV_All:    "CV_all",
	CV_InDirn: "CV_inDirn",
}

func Uninitialised_elevator() Elevator {
	uninitialised_elevator := Elevator{
		Floor:     -1,
		Dirn:      elevio.MD_Stop,
		Behaviour: EB_Idle,
		Config: struct {
			Clear_request_variant Clear_request_variant
			Door_open_duration_s  float64
		}{
			Clear_request_variant: CV_InDirn,
			Door_open_duration_s:  3.0,
		},
	}
	return uninitialised_elevator
}

func Clear_elevator_requests(elevator *Elevator) {
	for floor := 0; floor < utilities.N_FLOORS; floor++ {
		for btn := 0; btn < utilities.N_BUTTONS-1; btn++ {
			elevator.Requests[floor][btn] = false
		}
	}
}
