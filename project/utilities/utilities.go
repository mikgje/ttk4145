package utilities

// Supporting functions and data structures for entire elevator project

import (
	"flag"
	"time"
	"encoding/json"
	"os"
)

type Status_message struct {
	Controller_id int
	Behaviour     string
	Floor         int
	Direction     string
	Node_orders   [N_FLOORS][N_BUTTONS]bool
}

type Order_distribution_message struct {
	Orderlines [N_ELEVS][N_FLOORS][N_BUTTONS - 1]bool
}

type Controller_to_elevator_message struct {
	Label 			 string
	Orderline        [N_FLOORS][N_BUTTONS - 1]bool
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
	N_FLOORS  int = 4
	N_BUTTONS int = 3
	N_ELEVS   int = 3
)

const (
	Default_id        int    = -1
	Default_behaviour string = "idle"
	Default_direction string = "stop"
)

const (
	Elevator_update_rate_ms    time.Duration = 100 * time.Millisecond
	Obstruction_timer_duration int           = 10
)

const Network_prefix string = "peer-G49"

// Flags for configuring nodes on startup
var Id = flag.String("id", "", "Set id for node")
var Elevio = flag.Int("elevio", 15657, "Set port used for elevio")
var Peers = flag.Int("peers", 40000, "Set port used for peers")
var Bcast = flag.Int("bcast", 50000, "Set port used for bcast")

/*======================================== Save and load from disk functions ========================================*/

const Cab_calls_file_name = "cab_calls.json"

func Save_cab_calls(requests [N_FLOORS][N_BUTTONS]bool, column int, file_name string) error {
	cab_calls := make([]bool, len(requests))
	for i, row := range requests {
		cab_calls[i] = row[column]
	}
	data, err := json.Marshal(cab_calls)
	if err != nil {
		return err
	}
	return os.WriteFile(file_name, data, 0644)
}

func Load_cab_calls(requests *[N_FLOORS][N_BUTTONS]bool, column int, file_name string) error {
	data, err := os.ReadFile(file_name)
	if err != nil {
		return err
	}
	var cab_calls []bool
	if err := json.Unmarshal(data, &cab_calls); err != nil {
		return err
	}
	for i, cab_call := range cab_calls {
		requests[i][column] = cab_call
	}
	return nil
}

/*========================================UTILITY FUNCTIONS========================================*/

func Remove_from_slice(slice []string, item string) []string {
    for i, element := range slice {
        if element == item {
            return append(slice[:i], slice[i+1:]...)
        }
    }
    return slice
}