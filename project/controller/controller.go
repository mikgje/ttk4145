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
	state       				utilities.State = utilities.State_slave
	prev_odm 					utilities.Order_distribution_message
	node_statuses_chan			= make(chan utilities.Status_message, utilities.N_ELEVS)
	current_elevator			elevator.Elevator = elevator.Uninitialised_elevator()
	net							network.Network
	connected_elevators_status	= make(map[int]utilities.Status_message)

	node_status_chan			= make(chan utilities.Status_message, 1)
	service_orders_chan			= make(chan utilities.Order_distribution_message, 1)
	send_orders_chan			= make(chan utilities.Order_distribution_message, 1)
	dropped_peer_chan			= make(chan utilities.Status_message, 1)
)

func Start(
	elevator_status_chan <-chan elevator.Elevator,
	button_event_chan <-chan elevio.ButtonEvent,
	hall_orders_chan chan<- utilities.Controller_to_elevator_message,
	cab_orders_chan chan<- elevio.ButtonEvent,
	) {
	go network.Network_run(&net, service_orders_chan, send_orders_chan, node_status_chan, node_statuses_chan, dropped_peer_chan)
	controller_state_machine(elevator_status_chan, button_event_chan, hall_orders_chan, cab_orders_chan, node_status_chan, &net)

}

func controller_state_machine(
	elevator_status_chan <-chan elevator.Elevator,
	button_event_chan <-chan elevio.ButtonEvent,
	hall_orders_chan chan<- utilities.Controller_to_elevator_message,
	cab_orders_chan chan<- elevio.ButtonEvent,
	node_status_chan chan<- utilities.Status_message,
	net *network.Network,
	) {
	fmt.Println("Starting controller state machine")
	for {
		switch state {
		case utilities.State_slave:
			controller_modes.Slave(&prev_odm, &connected_elevators_status, &state, &current_elevator, elevator_status_chan, button_event_chan, hall_orders_chan, cab_orders_chan, node_status_chan, send_orders_chan, node_statuses_chan, dropped_peer_chan, net)
		case utilities.State_master:
			controller_modes.Master(&prev_odm, &connected_elevators_status, &state, &current_elevator, elevator_status_chan, button_event_chan, hall_orders_chan, cab_orders_chan, node_status_chan, send_orders_chan, service_orders_chan, node_statuses_chan, dropped_peer_chan, net)
		case utilities.State_disconnected:
			controller_modes.Disconnected(&state, &current_elevator, elevator_status_chan, button_event_chan, hall_orders_chan, cab_orders_chan, node_status_chan, net)
		}
	}
}
