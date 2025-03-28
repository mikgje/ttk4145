package controller_modes

//Controller modes for the controller state machine

import (
	"fmt"
	"main/controller/controller_tools"
	"main/hall_order_assigner"
	"main/elev_algo_go/elevator"
	"main/elevio"
	"main/network"
	"main/utilities"
)

func base_controller(
	status_message *utilities.Status_message,
	current_elevator *elevator.Elevator,
	controller_id *int,
	elevator_status_chan <-chan elevator.Elevator,
	button_event_chan <-chan elevio.ButtonEvent,
	hall_orders_chan chan<- utilities.Controller_to_elevator_message,
	cab_orders_chan chan<- elevio.ButtonEvent,
	node_status_chan chan<- utilities.Status_message,
	send_orders_chan <-chan utilities.Order_distribution_message,
	kill_base_ctrl_chan <-chan bool,
) {
	var augmented_requests [utilities.N_FLOORS][utilities.N_BUTTONS]bool
	for {
		select {
		case msg := <-elevator_status_chan:
			*current_elevator = msg
			*status_message = utilities.Status_message{Controller_id: *controller_id, Behaviour: elevator.EB_to_string[current_elevator.Behaviour],
				Floor: current_elevator.Floor, Direction: elevator.Dirn_to_string[current_elevator.Dirn], Node_orders: current_elevator.Requests}
		case msg := <-button_event_chan:
			new_order_floor := msg.Floor
			new_order_button := msg.Button
			new_order := elevio.ButtonEvent{Floor: new_order_floor, Button: new_order_button}

			if new_order_button == elevio.BT_Cab {
				cab_orders_chan <- new_order
			}
			augmented_requests = controller_tools.Augment_request_array(current_elevator.Requests, new_order)
			*status_message = utilities.Status_message{Controller_id: *controller_id, Behaviour: elevator.EB_to_string[current_elevator.Behaviour],
				Floor: current_elevator.Floor, Direction: elevator.Dirn_to_string[current_elevator.Dirn], Node_orders: augmented_requests}
		case msg := <-send_orders_chan /*The channel that supplies the ODM*/ :
			new_orders := controller_tools.Extract_orderline(*controller_id, msg)

			for floor := 0; floor < utilities.N_FLOORS; floor++ {
				for btn := 0; btn < utilities.N_BUTTONS-1; btn++ {
					status_message.Node_orders[floor][btn] = new_orders[floor][btn]
				}
			}
			other_orderlines := controller_tools.Extract_other_orderlines(*controller_id, msg)
			hall_orders_chan <- utilities.Controller_to_elevator_message{Orderline: new_orders, Other_orderlines: other_orderlines}
		case <-kill_base_ctrl_chan:
			return
		}

		if current_elevator.Floor != -1 {
			node_status_chan <- *status_message
		}
	}
}

func Slave(
	prev_odm *utilities.Order_distribution_message,
	connected_elevators_status *map[int]utilities.Status_message,
	state *utilities.State,
	current_elevator *elevator.Elevator,
	elevator_status_chan <-chan elevator.Elevator,
	button_event_chan <-chan elevio.ButtonEvent,
	hall_orders_chan chan<- utilities.Controller_to_elevator_message,
	cab_orders_chan chan<- elevio.ButtonEvent,
	node_status_chan chan<- utilities.Status_message,
	send_orders_chan <-chan utilities.Order_distribution_message,
	node_statuses_chan <-chan utilities.Status_message,
	dropped_peer_chan <-chan utilities.Status_message,
	net *network.Network,
) {
	kill_base_ctrl_chan := make(chan bool)
	var status_message = utilities.Status_message{Controller_id: net.Ctrl_id, Behaviour: elevator.EB_to_string[current_elevator.Behaviour],
		Floor: current_elevator.Floor, Direction: elevator.Dirn_to_string[current_elevator.Dirn], Node_orders: current_elevator.Requests}
	go base_controller(&status_message, current_elevator, &net.Ctrl_id, elevator_status_chan, button_event_chan, hall_orders_chan, cab_orders_chan, node_status_chan, send_orders_chan, kill_base_ctrl_chan)
	
	for {

		if !net.Connection {
			*state = utilities.State_disconnected
			kill_base_ctrl_chan <- true
			fmt.Println("Switching to disconnected mode")
			return
		}

		if net.Master {
			*state = utilities.State_master
			kill_base_ctrl_chan <- true
			fmt.Println("Switching to master mode")
			return
		}

		select {
		case msg := <-node_statuses_chan:
			if msg.Controller_id != utilities.Default_id {
				(*connected_elevators_status)[msg.Controller_id] = msg
			}
		case msg := <-dropped_peer_chan:
			disconnected_peer_status := msg
			disconnected_peer_status.Behaviour = elevator.EB_to_string[elevator.EB_Disconnected]
			(*connected_elevators_status)[msg.Controller_id] = disconnected_peer_status
		default:
		}
	}
}

func Master(
	prev_odm *utilities.Order_distribution_message,
	connected_elevators_status *map[int]utilities.Status_message,
	state *utilities.State,
	current_elevator *elevator.Elevator,
	elevator_status_chan <-chan elevator.Elevator,
	button_event_chan <-chan elevio.ButtonEvent,
	hall_orders_chan chan<- utilities.Controller_to_elevator_message,
	cab_orders_chan chan<- elevio.ButtonEvent,
	node_status_chan chan<- utilities.Status_message,
	send_orders_chan <-chan utilities.Order_distribution_message,
	service_orders_chan chan<- utilities.Order_distribution_message,
	node_statuses_chan <-chan utilities.Status_message,
	dropped_peer_chan <-chan utilities.Status_message,
	net *network.Network,
) {
	var kill_base_ctrl_chan = make(chan bool)
	var status_message = utilities.Status_message{Controller_id: net.Ctrl_id, Behaviour: elevator.EB_to_string[current_elevator.Behaviour],
		Floor: current_elevator.Floor, Direction: elevator.Dirn_to_string[current_elevator.Dirn], Node_orders: current_elevator.Requests}
	go base_controller(&status_message, current_elevator, &net.Ctrl_id, elevator_status_chan, button_event_chan, hall_orders_chan, cab_orders_chan, node_status_chan, send_orders_chan, kill_base_ctrl_chan)

	for {
		select {
		case msg := <-node_statuses_chan:

			if msg.Controller_id != utilities.Default_id {
				(*connected_elevators_status)[msg.Controller_id] = msg
			}
			outer_loop:
			for i := 0; i < utilities.N_ELEVS; i++ {
				select {
				case msg := <-node_statuses_chan:
					if msg.Controller_id != utilities.Default_id {
						(*connected_elevators_status)[msg.Controller_id] = msg
					}
				default:
					break outer_loop
				}
			}
			(*connected_elevators_status)[net.Ctrl_id] = status_message
			status_to_order_handler := make([]utilities.Status_message, 0, len((*connected_elevators_status)))

			for _, status := range *connected_elevators_status {
				if status.Behaviour == elevator.EB_to_string[elevator.EB_Disconnected] {
					status.Behaviour = elevator.EB_to_string[elevator.EB_Obstructed]
					status_to_order_handler = append(status_to_order_handler, status)
					delete((*connected_elevators_status), status.Controller_id)
				} else {
					status_to_order_handler = append(status_to_order_handler, status)
				}
			}

			var new_odm utilities.Order_distribution_message
			if len(status_to_order_handler) > 0 {
				new_odm = order_handler.Order_handler(status_to_order_handler)
			} else {
				new_odm = *prev_odm
			}
		
			if new_odm != *prev_odm {
				service_orders_chan <- new_odm
				*prev_odm = new_odm
				for i := 0; i < 50; i++ {
					service_orders_chan <- new_odm
				}
				controller_tools.Flush_status_messages(node_statuses_chan)
			}

		case msg := <-dropped_peer_chan:
			fmt.Println("Dropped peer")
			disconnected_peer_status := msg
			disconnected_peer_status.Controller_id = net.N_nodes
			disconnected_peer_status.Behaviour = elevator.EB_to_string[elevator.EB_Disconnected]
			(*connected_elevators_status)[net.N_nodes] = disconnected_peer_status
			controller_tools.Flush_status_messages(node_statuses_chan)
		default:
		}

		if !net.Connection {
			*state = utilities.State_disconnected
			kill_base_ctrl_chan <- true
			fmt.Println("Switching to disconnected mode")
			return
		}

		if !net.Master {
			*state = utilities.State_slave
			kill_base_ctrl_chan <- true
			fmt.Println("Switching to slave mode")
			return
		}
	}
}

func Disconnected(
	state *utilities.State,
	current_elevator *elevator.Elevator,
	elevator_status_chan <-chan elevator.Elevator,
	button_event_chan <-chan elevio.ButtonEvent,
	hall_orders_chan chan<- utilities.Controller_to_elevator_message,
	cab_orders_chan chan<- elevio.ButtonEvent,
	net *network.Network,
) {

	for {
		select {
		case msg := <-elevator_status_chan:
			*current_elevator = msg
		case msg := <-button_event_chan:
			new_order_floor := msg.Floor
			new_order_button := msg.Button
			new_order := elevio.ButtonEvent{Floor: new_order_floor, Button: new_order_button}
			if new_order_button == elevio.BT_Cab {
				cab_orders_chan <- new_order
			} else {
				var orderline [utilities.N_FLOORS][utilities.N_BUTTONS-1]bool
				orderline[new_order_floor][new_order_button] = true
				hall_orders_chan <- utilities.Controller_to_elevator_message{Label: elevator.EB_to_string[elevator.EB_Disconnected], Orderline: orderline}
			}
		default:
		}

		if net.Connection {
			*state = utilities.State_slave
			fmt.Println("Switching to slave mode")
			return
		}
	}
}
