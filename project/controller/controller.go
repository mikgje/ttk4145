package controller

/*-------------------------------------*/
// INPUT:
// Elevator copy
// Request orders from elevator
// ODM from network
// Master status from network
// Controller ID from network
// IF master: other node status messages
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
	state                       utilities.State   = utilities.State_slave
	ctrl_to_network_chan                          = make(chan utilities.StatusMessage, 1)
	ODM_to_network_chan                           = make(chan utilities.OrderDistributionMessage, 1)
	bcast_sorders_chan                            = make(chan utilities.OrderDistributionMessage, 1)
	dropped_peer_chan                             = make(chan utilities.StatusMessage, 1)
	other_elevators_status_chan                   = make(chan utilities.StatusMessage, utilities.N_ELEVS)
	current_elevator            elevator.Elevator = elevator.Elevator_uninitialised()
	net                         network.Network
)

func Start(
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- utilities.ControllerToElevatorMessage,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	) {
	go network.Network_master(&net, ODM_to_network_chan, bcast_sorders_chan, ctrl_to_network_chan, other_elevators_status_chan)
	controller_state_machine(elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan, &net)

}

func controller_state_machine(
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- utilities.ControllerToElevatorMessage,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	ctrl_to_network_chan chan<- utilities.StatusMessage,
	net *network.Network,
	) {

	fmt.Println("Starting controller state machine")

	for {
		switch state {
		case utilities.State_slave:
			controller_modes.Slave(&state, &current_elevator, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan, bcast_sorders_chan, net)
		case utilities.State_master:
			controller_modes.Master(&state, &current_elevator, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan, bcast_sorders_chan, ODM_to_network_chan, other_elevators_status_chan, dropped_peer_chan, net)
		case utilities.State_disconnected:
			controller_modes.Disconnected(&state, &current_elevator, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan, net)
		}
	}
}
