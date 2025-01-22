package elevator

import(
	"fmt"
)

const(
	N_FLOORS int = 4
	N_BUTTONS int = 3 
)

type ElevatorBehaviour int
type ClearRequestVariant int
type Direction int
type Button int

type Elevator struct{
	floor int
	direction Direction
	requests [N_FLOORS][N_BUTTONS]int
	behaviour ElevatorBehaviour

	config struct{
		clear_request_variant ClearRequestVariant
		door_open_duration_s float64
	}
}

// Enum for clear request variant
const(
	CV_all ClearRequestVariant = iota
	CV_inDirn
)

// Enum for elevator behaviour
const(
	EB_idle ElevatorBehaviour = iota
	EB_doorOpen
	EB_moving
)

// Enum for elevator direction
const(
	D_down Direction = -1
	D_stop Direction = 0
	D_up Direction = 1
)

// Enum for button type
const(
	B_hallUp Button = iota
	B_hallDown
	B_cab
)

// map elevator behaviour to strings for printing
var eb_to_string = map[ElevatorBehaviour]string{
	EB_idle: "EB_idle",
	EB_doorOpen: "EB_doorOpen",
	EB_moving: "EB_moving",
}

// map elevator direction to strings
var dirn_to_string = map[Direction]string{
	D_down: "D_down",
	D_stop: "D_stop",
	D_up: "D_up",
}

// map buttons to strings
var button_to_string = map[Button]string{
	B_hallUp: "B_hallUp",
	B_hallDown: "B_hallDown",
	B_cab: "B_cab",
}

// map request variant to strings
var cv_to_string = map[ClearRequestVariant]string{
	CV_all: "CV_all",
	CV_inDirn: "CV_inDirn",
}

// Returns an uninitialised elevator object to be used
// and configured in the main loop and fsm
func elevator_uninitialised() Elevator{
	uninitialised_elevator := Elevator{
		floor: -1,
		direction: D_stop,
		behaviour: EB_idle,
		config: struct {
			clear_request_variant ClearRequestVariant
			door_open_duration_s float64
		}{
			clear_request_variant: CV_all,
			door_open_duration_s: 3.0,
		},
	}

	return uninitialised_elevator
}

// Prints all the stats of the elevator
func print_elevator_stats(elevator Elevator){
	fmt.Println("Floor:", elevator.floor)
	fmt.Println("Direction:", dirn_to_string[elevator.direction])
	fmt.Println("Behaviour:", eb_to_string[elevator.behaviour])
	fmt.Println("Clear request variant:", cv_to_string[elevator.config.clear_request_variant])
	fmt.Println("Door open duration:", elevator.config.door_open_duration_s)
	fmt.Printf("\n") // Add padding in terminal
}

func main(){
	elevator := elevator_uninitialised()
	print_elevator_stats(elevator)
}
