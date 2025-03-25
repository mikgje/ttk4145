package utilities

import (
	"time"
	"flag"
)

type StatusMessage struct {
	Controller_id int
	Behaviour     string
	Floor         int
	Direction     string
	Node_orders   [N_FLOORS][N_BUTTONS]bool
}

type OrderDistributionMessage struct {
	Orderlines [N_ELEVS][N_FLOORS][N_BUTTONS - 1]bool
}

type ControllerToElevatorMessage struct {
	Orderline [N_FLOORS][N_BUTTONS - 1]bool
	Other_orderlines [][N_FLOORS][N_BUTTONS - 1]bool
}

type State int

const (
	State_slave State = iota
	State_master
	State_disconnected
)

/*========================================PROJECT CONSTANTS========================================*/

const (
	N_FLOORS  	int = 4
	N_BUTTONS 	int = 3
	N_ELEVS 	int = 3
)

const (
	Default_id			int 	= -1
	Default_behaviour 	string 	= "idle"
	Default_direction 	string 	= "stop"
)

const (
	Elevator_update_rate_ms time.Duration = 100*time.Millisecond
)

const (
	Network_prefix string = "peer-G49"
)

var Debug	= flag.Bool("debug", false, "Enable debug mode")
var Id		= flag.String("id", "", "Set id for node")
var Elevio 	= flag.Int("elevio", 15657, "Set port used for elevio")
var Peers 	= flag.Int("peers", 15647, "Set port used for peers")
var Bcast 	= flag.Int("bcast", 16569, "Set port used for bcast")
