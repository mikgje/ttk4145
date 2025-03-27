package controller

/*-------------------------------------*/
// INPUT:
// elevator_status_chan: Elevator status from local elevator
// button_event_chan: Button events from local elevator
// send_orders_chan: Service orders from network
// *net: Network struct containing master status for local controller, connectivity, number of nodes on the network and controller ID.
// node_statuses_chan: Status messages from all controllers on network
/*-------------------------------------*/
// OUTPUT:
// hall_orders_chan: Hall orders to local elevator
// cab_orders_chan: Cab orders to local elevator
// node_status_chan: Status messages to network from local controller
// IF master, service_orders_chan: Order distribution message to network from local controller
/*-------------------------------------*/

// Interface between controller network and elevator
import (
	"fmt"
	"main/controller/controller_modes"
	"main/elev_algo_go/elevator"
	"main/elevio"
	"main/network"
	"main/utilities"
)

var (
	state                       utilities.State   = utilities.State_slave
	prev_odm 					utilities.Order_distribution_message
	ctrl_to_network_chan                          = make(chan utilities.Status_message, 1)
	ODM_to_network_chan                           = make(chan utilities.Order_distribution_message, 1)
	bcast_sorders_chan                            = make(chan utilities.Order_distribution_message, 1)
	dropped_peer_chan                             = make(chan utilities.Status_message, 1)
	cab_call_from_network_chan 					  = make(chan [utilities.N_FLOORS]bool, 1)
	other_elevators_status_chan                   = make(chan utilities.Status_message, utilities.N_ELEVS)
	current_elevator            elevator.Elevator = elevator.Uninitialised_elevator()
	net                         network.Network
	just_booted				 	bool
	has_ever_connected 			bool
	connected_elevators_status = make(map[int]utilities.Status_message)
)

func Start(
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- utilities.Controller_to_elevator_message,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	) {
	just_booted = true
	go network.Network_run(&net, ODM_to_network_chan, bcast_sorders_chan, ctrl_to_network_chan, other_elevators_status_chan, dropped_peer_chan)
	controller_state_machine(elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan, &net)

}

func controller_state_machine(
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- utilities.Controller_to_elevator_message,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	ctrl_to_network_chan chan<- utilities.Status_message,
	net *network.Network,
	) {

	fmt.Println("Starting controller state machine")

	for {
		switch state {
		case utilities.State_slave:
			controller_modes.Slave(&prev_odm, &connected_elevators_status, &has_ever_connected, &just_booted, &state, &current_elevator, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan, bcast_sorders_chan, other_elevators_status_chan, dropped_peer_chan, cab_call_from_network_chan, net)
		case utilities.State_master:
			controller_modes.Master(&prev_odm, &connected_elevators_status, &just_booted, &state, &current_elevator, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan, bcast_sorders_chan, ODM_to_network_chan, other_elevators_status_chan, dropped_peer_chan, cab_call_from_network_chan, net)
		case utilities.State_disconnected:
			controller_modes.Disconnected(&has_ever_connected, &just_booted, &state, &current_elevator, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan, net)
		}
	}
}
