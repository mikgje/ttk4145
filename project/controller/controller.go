package controller

/*-------------------------------------*/
// INPUT:
// Elevator copy
// Request orders from elevator
// ODM from network
// Master status from network
// Controller ID from network
// IF master: other node status messages
// IF master: ODM from network
/*-------------------------------------*/
// OUTPUT:
// Service orders to the elevator
// Status to network
// IF master: ODM to network
// IF master: other node status messages to order assigner
/*-------------------------------------*/


// Interface between controller network and elevator
import (
	"fmt"
	"main/controll/controller_modes"
	"main/elev_algo_go/elevator"
	"main/elevio"
	"main/network"
	"main/utilities"
	//For testing
)

var (
	state                  utilities.State = utilities.State_slave
	ctrl_to_network_chan                   = make(chan utilities.StatusMessage, 1)
	ODM_to_network_chan                    = make(chan utilities.OrderDistributionMessage, 1)
	bcast_sorders_chan                     = make(chan utilities.OrderDistributionMessage, 1)
	current_elevator       elevator.Elevator
	controller_id          int
	other_elevators_status = make(map[int]utilities.StatusMessage) //Map of other elevators based on ID
)

func Start(
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent) {
	/* Placeholder for network routines */
	// go TEMP_transmit_network(network_send_chan)
	// go TEMP_receive_network(network_receive_order_chan)
	/* End of placeholder */

	go network.Network(ODM_to_network_chan, bcast_sorders_chan, ctrl_to_network_chan)
	controller_state_machine(elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan)

}

func controller_state_machine(
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	ctrl_to_network_chan chan<- utilities.StatusMessage) {

	for {
		switch state {
		case utilities.State_slave:
			fmt.Println("Starting normal controller")
			controller_modes.Slave(&state, &current_elevator, controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan)
		case utilities.State_master:
			fmt.Println("Starting primary controller")
			controller_modes.Master(&state, &current_elevator, controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan, ODM_to_network_chan)
		case utilities.State_disconnected:
			fmt.Println("Starting disconnected controller")
			controller_modes.Disconnected(&state, &current_elevator, controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan)
		}
	}
}
