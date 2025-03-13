package main_controller

// Interface between controller network and elevator
import (
	"fmt"
	"main/controll/controller_modes"
	"main/elev_algo_go/elevator"
	"main/elevio"
	"main/utilities"
	//For testing
)

type State int

const (
	state_normal State = iota
	state_backup
	state_primary
	state_disconnected
)

var (
	// controller_id               int
	// is_elevator_healthy         bool  = true //Used to signal to network, the current elevator is not to be considered for new orders
	state State = 0
	// augmented_requests          [utilities.N_FLOORS][utilities.N_BUTTONS]bool
	// current_elevator            elevator.Elevator
)

func Main_controller(controller_id int,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent) {
	/* Placeholder for network routines */
	// go TEMP_transmit_network(network_send_chan)
	// go TEMP_receive_network(network_receive_order_chan)
	/* End of placeholder */

	controller_state_machine(state, controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan)

}

func controller_state_machine(state State,
	controller_id int,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent) {

	switch state {
	case state_normal:
		fmt.Println("Starting normal controller")
		controller_modes.Normal_controller(controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan)
	case state_backup:
		fmt.Println("Starting backup controller")
		// 	controller_modes.Backup_controller(controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, network_receive_chan, network_send_chan)
		// case state_primary:
		// 	fmt.Println("Starting primary controller")
		// 	controller_modes.Primary_controller(controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, network_receive_chan, network_send_chan)
		// case state_disconnected:
		// 	fmt.Println("Starting disconnected controller")
		// 	controller_modes.Disconnected_controller(controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, network_receive_chan, network_send_chan)
	}
}
