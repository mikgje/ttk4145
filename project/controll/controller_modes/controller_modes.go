package controller_modes

import (
	"fmt"
	"main/controll/controller_tools"
	order_handler "main/controll/hall_order_assigner"
	"main/elev_algo_go/elevator"
	"main/elevio"
	"main/network"
	"main/utilities"
)

func base_controller(status_message *utilities.StatusMessage, current_elevator *elevator.Elevator, controller_id int,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	ctrl_to_network_chan chan<- utilities.StatusMessage,
	network_to_ctrl_chan <-chan utilities.OrderDistributionMessage,
	kill_base_ctrl_chan <-chan bool) {

	var augmented_requests [utilities.N_FLOORS][utilities.N_BUTTONS]bool

	for {
		select {
		case msg := <-elev_to_ctrl_chan:
			*current_elevator = msg
			// fmt.Println("Controller modes: ", *current_elevator)
			*status_message = utilities.StatusMessage{Controller_id: controller_id, Behaviour: elevator.EB_to_string[current_elevator.Behaviour],
				Floor: current_elevator.Floor, Direction: elevator.Dirn_to_string[current_elevator.Dirn], Node_orders: augmented_requests}
		case msg := <-elev_to_ctrl_button_chan:
			new_order_floor := msg.Floor
			new_order_button := msg.Button
			new_order := elevio.ButtonEvent{Floor: new_order_floor, Button: new_order_button}
			fmt.Printf("New order from local elevator: Floor %d, Button %s\n", new_order_floor, elevator.Button_to_string[new_order_button])
			if new_order_button == elevio.BT_Cab {
				ctrl_to_elev_cab_chan <- new_order
			}

			augmented_requests = controller_tools.Augment_request_array(current_elevator.Requests, new_order)
			*status_message = utilities.StatusMessage{Controller_id: controller_id, Behaviour: elevator.EB_to_string[current_elevator.Behaviour],
				Floor: current_elevator.Floor, Direction: elevator.Dirn_to_string[current_elevator.Dirn], Node_orders: augmented_requests}

		case msg := <-network_to_ctrl_chan /*The channel that supplies the ODM*/ :
			new_orders := controller_tools.Extract_orderline(controller_id, msg)
			ctrl_to_elev_chan <- new_orders
		case <-kill_base_ctrl_chan:
			return
		}
		ctrl_to_network_chan <- *status_message
	}
}

func Slave(state *utilities.State, current_elevator *elevator.Elevator, controller_id int,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	ctrl_to_network_chan chan<- utilities.StatusMessage,
	network_to_ctrl_chan <-chan utilities.OrderDistributionMessage,
	net *network.Network) {

	kill_base_ctrl_chan := make(chan bool)

	var status_message = utilities.StatusMessage{Controller_id: controller_id, Behaviour: elevator.EB_to_string[current_elevator.Behaviour],
		Floor: current_elevator.Floor, Direction: elevator.Dirn_to_string[current_elevator.Dirn], Node_orders: current_elevator.Requests}

	go base_controller(&status_message, current_elevator, controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan, network_to_ctrl_chan, kill_base_ctrl_chan)
	for {
		if net.Master {
			*state = utilities.State_master
			kill_base_ctrl_chan <- true
			fmt.Println("Switching to master mode")
			return
		} else if !net.Connection {
			*state = utilities.State_disconnected
			kill_base_ctrl_chan <- true
			fmt.Println("Switching to disconnected mode")
			return
		}

	}

}

func Master(state *utilities.State, current_elevator *elevator.Elevator, controller_id int,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	ctrl_to_network_chan chan<- utilities.StatusMessage,
	network_to_ctrl_chan <-chan utilities.OrderDistributionMessage,
	ODM_to_network_chan chan<- utilities.OrderDistributionMessage,
	other_elevators_status <-chan utilities.StatusMessage,
	net *network.Network) {

	var prev_odm utilities.OrderDistributionMessage

	var healthy_elevators_status = make(map[int]utilities.StatusMessage)
	var unhealty_elevators = make(map[int]bool)
	var kill_base_ctrl_chan = make(chan bool)
	var status_message = utilities.StatusMessage{Controller_id: controller_id, Behaviour: elevator.EB_to_string[current_elevator.Behaviour],
		Floor: current_elevator.Floor, Direction: elevator.Dirn_to_string[current_elevator.Dirn], Node_orders: current_elevator.Requests}

	go base_controller(&status_message, current_elevator, controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan, network_to_ctrl_chan, kill_base_ctrl_chan)

	for {
		select {
		// TODO: Function only runs assigner if other elevators have request orders? I.e. hall orders can be serviced by itself.
		// M has temporarily changed other_elevators_status to be of size 3 and write_statuses to include local node.
		case msg := <-other_elevators_status:
			if msg.Controller_id == utilities.Default_id {
				continue
			}
			if unhealty_elevators[msg.Controller_id] && (msg.Behaviour != elevator.EB_to_string[elevator.EB_Unhealthy]) {
				delete(unhealty_elevators, msg.Controller_id)
				healthy_elevators_status[msg.Controller_id] = msg

			} else if msg.Behaviour == elevator.EB_to_string[elevator.EB_Unhealthy] {
				unhealty_elevators[msg.Controller_id] = true
				delete(healthy_elevators_status, msg.Controller_id)

			} else {
				healthy_elevators_status[msg.Controller_id] = msg
			}

			status_to_order_handler := make([]utilities.StatusMessage, 0, len(healthy_elevators_status))
			
			status_to_order_handler = append(status_to_order_handler, status_message)

			for _, status := range healthy_elevators_status {
				status_to_order_handler = append(status_to_order_handler, status)
			}
			new_odm := order_handler.Order_handler(status_to_order_handler)
			if new_odm != prev_odm {
				fmt.Println("Sending new ODM to network")
				ODM_to_network_chan <- new_odm
				prev_odm = new_odm
			}

		default:
			status_to_order_handler := make([]utilities.StatusMessage, 0, 1)
			status_to_order_handler = append(status_to_order_handler, status_message)
			new_odm := order_handler.Order_handler(status_to_order_handler)
			if new_odm != prev_odm {
				fmt.Println("Sending new ODM to network")
				ODM_to_network_chan <- new_odm
				prev_odm = new_odm
			}

			if !net.Connection {
				*state = utilities.State_disconnected
				kill_base_ctrl_chan <- true
				fmt.Println("Switching to disconnected mode")
				return
			}
		}
	}
}

func Disconnected(state *utilities.State, current_elevator *elevator.Elevator, controller_id int,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	ctrl_to_network_chan chan<- utilities.StatusMessage,
	net *network.Network) {

	for {
		select {
		case msg := <-elev_to_ctrl_chan:
			*current_elevator = msg
		case msg := <-elev_to_ctrl_button_chan:
			new_order_floor := msg.Floor
			new_order_button := msg.Button
			new_order := elevio.ButtonEvent{Floor: new_order_floor, Button: new_order_button}
			fmt.Printf("New order from local elevator: Floor %d, Button %s\n", new_order_floor, elevator.Button_to_string[new_order_button])
			if new_order_button == elevio.BT_Cab {
				ctrl_to_elev_cab_chan <- new_order
			} else {
				fmt.Println("Disconnected, cannot send orders to network")
			}
		default:
			if net.Connection {
				*state = utilities.State_slave
				fmt.Println("Switching to slave mode")
				return
			}
		}
	}
}
