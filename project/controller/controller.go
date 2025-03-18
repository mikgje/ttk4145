package controller

// Interface between controller network and elevator
import (
	"fmt"
	"main/controll/controller_modes"
	"main/elev_algo_go/elevator"
	"main/elevio"
	"main/utilities"
	"main/network"
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
	state State = 0
	ctrl_to_network_chan = make(chan utilities.StatusMessage, 1)
	assign_chan = make(chan utilities.OrderDistributionMessage, 1)
	bcast_sorders_chan = make(chan utilities.OrderDistributionMessage, 1)
	current_elevator elevator.Elevator
)

func Start(controller_id int,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent) {
	/* Placeholder for network routines */
	// go TEMP_transmit_network(network_send_chan)
	// go TEMP_receive_network(network_receive_order_chan)
	/* End of placeholder */
	
	go network.Network(assign_chan, bcast_sorders_chan, ctrl_to_network_chan)
	controller_state_machine(state, controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan)

}

func controller_state_machine(state State,
	controller_id int,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	ctrl_to_network_chan chan<- utilities.StatusMessage) {

	for {
		switch state {
		case state_normal:
			fmt.Println("Starting normal controller")
			controller_modes.Normal(&current_elevator, controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan)
		case state_backup:
			fmt.Println("Starting backup controller")
			controller_modes.Backup(&current_elevator, controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan)
		case state_primary:
			fmt.Println("Starting primary controller")
			controller_modes.Primary(&current_elevator, controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan)
		case state_disconnected:
			fmt.Println("Starting disconnected controller")
			controller_modes.Disconnected(&current_elevator, controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan)
		}
}
}
