package elevator

import (
	"fmt"
	"main/elevio"
)

const (
	N_FLOORS  int = 4
	N_BUTTONS int = 3
)

type ElevatorBehaviour int
type ClearRequestVariant int
type Direction int
type Button int

type Elevator struct {
	Floor     int
	Dirn      elevio.MotorDirection
	Requests  [N_FLOORS][N_BUTTONS]bool
	Behaviour ElevatorBehaviour

	Config struct {
		ClearRequestVariant ClearRequestVariant
		DoorOpenDuration_s  float64
	}
}

// Enum for clear request variant
const (
	CV_All ClearRequestVariant = iota
	CV_InDirn
)

// Enum for elevator behaviour
const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

// map elevator behaviour to strings for printing
var eb_to_string = map[ElevatorBehaviour]string{
	EB_Idle:     "EB_idle",
	EB_DoorOpen: "EB_doorOpen",
	EB_Moving:   "EB_moving",
}

// map elevator direction to strings
var dirn_to_string = map[elevio.MotorDirection]string{
	elevio.MD_Up:   "MD_up",
	elevio.MD_Down: "MD_down",
	elevio.MD_Stop: "MD_stop",
}

// map buttons to strings
var button_to_string = map[elevio.ButtonType]string{
	elevio.BT_HallUp:   "B_hallUp",
	elevio.BT_HallDown: "B_hallDown",
	elevio.BT_Cab:      "B_cab",
}

// map request variant to strings
var cv_to_string = map[ClearRequestVariant]string{
	CV_All:    "CV_all",
	CV_InDirn: "CV_inDirn",
}

// Returns an uninitialised elevator object to be used
// and configured in the main loop and fsm
func Elevator_uninitialised() Elevator {
	uninitialised_elevator := Elevator{
		Floor:     -1,
		Dirn:      elevio.MD_Stop,
		Behaviour: EB_Idle,
		Config: struct {
			ClearRequestVariant ClearRequestVariant
			DoorOpenDuration_s  float64
		}{
			ClearRequestVariant: CV_All,
			DoorOpenDuration_s:  3.0,
		},
	}

	return uninitialised_elevator
}

// Prints all the stats of the elevator
func Elevator_print(elevator Elevator) {
	fmt.Println("Floor:", elevator.Floor)
	fmt.Println("Direction:", dirn_to_string[elevator.Dirn])
	fmt.Println("Behaviour:", eb_to_string[elevator.Behaviour])
	fmt.Println("Clear request variant:", cv_to_string[elevator.Config.ClearRequestVariant])
	fmt.Println("Door open duration:", elevator.Config.DoorOpenDuration_s)
	fmt.Printf("\n") // Add padding in terminal
}
