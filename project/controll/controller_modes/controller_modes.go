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

func base_controller(
	just_booted *bool,
	status_message *utilities.StatusMessage,
	current_elevator *elevator.Elevator,
	controller_id *int,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- utilities.ControllerToElevatorMessage,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	ctrl_to_network_chan chan<- utilities.StatusMessage,
	network_to_ctrl_chan <-chan utilities.OrderDistributionMessage,
	cab_call_from_network_chan <-chan [utilities.N_FLOORS]bool,
	kill_base_ctrl_chan <-chan bool,
) {

	var augmented_requests [utilities.N_FLOORS][utilities.N_BUTTONS]bool

	for {
		select {
		case msg := <-elev_to_ctrl_chan:
			*current_elevator = msg
			*status_message = utilities.StatusMessage{Controller_id: *controller_id, Behaviour: elevator.EB_to_string[current_elevator.Behaviour],
				Floor: current_elevator.Floor, Direction: elevator.Dirn_to_string[current_elevator.Dirn], Node_orders: current_elevator.Requests}
		case msg := <-elev_to_ctrl_button_chan:
			new_order_floor := msg.Floor
			new_order_button := msg.Button
			new_order := elevio.ButtonEvent{Floor: new_order_floor, Button: new_order_button}
			if new_order_button == elevio.BT_Cab {
				ctrl_to_elev_cab_chan <- new_order
			}

			augmented_requests = controller_tools.Augment_request_array(current_elevator.Requests, new_order)

			*status_message = utilities.StatusMessage{Controller_id: *controller_id, Behaviour: elevator.EB_to_string[current_elevator.Behaviour],
				Floor: current_elevator.Floor, Direction: elevator.Dirn_to_string[current_elevator.Dirn], Node_orders: augmented_requests}
			fmt.Println("Status message: ", *status_message)

		case msg := <-network_to_ctrl_chan /*The channel that supplies the ODM*/ :
			new_orders := controller_tools.Extract_orderline(*controller_id, msg)
			fmt.Println("New orders: ", new_orders)

			for floor := 0; floor < utilities.N_FLOORS; floor++ {
				for btn := 0; btn < utilities.N_BUTTONS-1; btn++ {
					status_message.Node_orders[floor][btn] = new_orders[floor][btn]
				}
			}

			other_orderlines := controller_tools.Extract_other_orderlines(*controller_id, msg)
			ctrl_to_elev_chan <- utilities.ControllerToElevatorMessage{Orderline: new_orders, Other_orderlines: other_orderlines}
		//TODO: Set up channel to transport cab calls from network to base controller in case of restart
		case msg := <-cab_call_from_network_chan:
			if *just_booted {
				for i := 0; i < utilities.N_FLOORS; i++ {
					if msg[i] {
						ctrl_to_elev_cab_chan <- elevio.ButtonEvent{i, elevio.BT_Cab}
					}
				}
				*just_booted = false
			}
		case <-kill_base_ctrl_chan:
			return
		}
		if current_elevator.Floor != -1 {
			// fmt.Println("Sending status message")
			ctrl_to_network_chan <- *status_message
		}
	}
}

func Slave(
	prev_odm *utilities.OrderDistributionMessage,
	connected_elevators_status *map[int]utilities.StatusMessage,
	has_ever_connected *bool,
	just_booted *bool,
	state *utilities.State,
	current_elevator *elevator.Elevator,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- utilities.ControllerToElevatorMessage,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	ctrl_to_network_chan chan<- utilities.StatusMessage,
	network_to_ctrl_chan <-chan utilities.OrderDistributionMessage,
	other_elevators_status_chan <-chan utilities.StatusMessage,
	dropped_peer_chan <-chan utilities.StatusMessage,
	cab_call_from_network_chan <-chan [utilities.N_FLOORS]bool,
	net *network.Network,
) {

	kill_base_ctrl_chan := make(chan bool)

	var status_message = utilities.StatusMessage{Controller_id: net.Ctrl_id, Behaviour: elevator.EB_to_string[current_elevator.Behaviour],
		Floor: current_elevator.Floor, Direction: elevator.Dirn_to_string[current_elevator.Dirn], Node_orders: current_elevator.Requests}

	go base_controller(just_booted, &status_message, current_elevator, &net.Ctrl_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan, network_to_ctrl_chan, cab_call_from_network_chan, kill_base_ctrl_chan)
	for {
		if !net.Connection {
			*state = utilities.State_disconnected
			kill_base_ctrl_chan <- true
			fmt.Println("Switching to disconnected mode")
			return
		}
		if !*has_ever_connected {
			*has_ever_connected = true
			fmt.Println("Confirmed connection")
		}
		if net.Master {
			*state = utilities.State_master
			kill_base_ctrl_chan <- true
			fmt.Println("Switching to master mode")
			return
		}
		select {
		case msg := <-other_elevators_status_chan:
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
	prev_odm *utilities.OrderDistributionMessage,
	connected_elevators_status *map[int]utilities.StatusMessage,
	just_booted *bool,
	state *utilities.State,
	current_elevator *elevator.Elevator,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- utilities.ControllerToElevatorMessage,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	ctrl_to_network_chan chan<- utilities.StatusMessage,
	network_to_ctrl_chan <-chan utilities.OrderDistributionMessage,
	ODM_to_network_chan chan<- utilities.OrderDistributionMessage,
	other_elevators_status_chan <-chan utilities.StatusMessage,
	dropped_peer_chan <-chan utilities.StatusMessage,
	cab_call_from_network_chan <-chan [utilities.N_FLOORS]bool,
	net *network.Network,
) {

	var kill_base_ctrl_chan = make(chan bool)
	var dropped_peer bool = false
	var status_message = utilities.StatusMessage{Controller_id: net.Ctrl_id, Behaviour: elevator.EB_to_string[current_elevator.Behaviour],
		Floor: current_elevator.Floor, Direction: elevator.Dirn_to_string[current_elevator.Dirn], Node_orders: current_elevator.Requests}

	go base_controller(just_booted, &status_message, current_elevator, &net.Ctrl_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan, network_to_ctrl_chan, cab_call_from_network_chan, kill_base_ctrl_chan)

	for {
		select {

		case msg := <-other_elevators_status_chan:

			if msg.Controller_id != utilities.Default_id {
				(*connected_elevators_status)[msg.Controller_id] = msg
				// fmt.Println("Controller id: ", msg.Controller_id)
			}
			
			outer_loop:
			for i := 0; i < utilities.N_ELEVS; i++ {
				select {
				case msg := <-other_elevators_status_chan:
					if msg.Controller_id != utilities.Default_id {
						(*connected_elevators_status)[msg.Controller_id] = msg
						// fmt.Println("Controller id outside if: ", msg.Controller_id)
					}
				default:

					break outer_loop
				}
			}


			status_slice := make([]utilities.StatusMessage, 0, len((*connected_elevators_status)))

			for _, status := range *connected_elevators_status {
				status_slice = append(status_slice, status)

			}

			status_to_order_handler := make([]utilities.StatusMessage, 0, len((*connected_elevators_status)))
			for _, status := range *connected_elevators_status {

				if status.Behaviour == elevator.EB_to_string[elevator.EB_Disconnected] {
					// fmt.Println("Received disconnected elevator")
					status.Behaviour = elevator.EB_to_string[elevator.EB_Obstructed]
					status_to_order_handler = append(status_to_order_handler, status)
					// fmt.Println("Connected_elevators_status before delete: ", *connected_elevators_status)
					delete((*connected_elevators_status), status.Controller_id)
					// fmt.Println("Connected_elevators_status after delete: ", *connected_elevators_status)
					// fmt.Println("Status to order handler: ", status_to_order_handler)

				} else {
					status_to_order_handler = append(status_to_order_handler, status)
					if dropped_peer {
						// fmt.Println("Status to order handler: ", status_to_order_handler)
					}
				}
			}

			var new_odm utilities.OrderDistributionMessage
			if len(status_to_order_handler) > 0 {
				new_odm = order_handler.Order_handler(status_to_order_handler)
			} else {
				new_odm = *prev_odm
			}
			if new_odm != *prev_odm {
				fmt.Println("Status_to_order_handler: ", status_to_order_handler)
				fmt.Println("New ODM: ", new_odm)
				fmt.Println("Connected elevators status: ", *connected_elevators_status, "\n\n\n\n")

				ODM_to_network_chan <- new_odm
				*prev_odm = new_odm

				for i := 0; i < 50; i++ {
					ODM_to_network_chan <- new_odm
				}

				// BREAK GLASS IN CASE OF EMEGENCY
				controller_tools.Flush_status_messages(other_elevators_status_chan)
				dropped_peer = false
			}

		case msg := <-dropped_peer_chan:
			dropped_peer = true
			fmt.Println("Dropped peer")
			disconnected_peer_status := msg
			disconnected_peer_status.Controller_id = net.N_nodes
			disconnected_peer_status.Behaviour = elevator.EB_to_string[elevator.EB_Disconnected]
			(*connected_elevators_status)[net.N_nodes] = disconnected_peer_status
			// BREAK GLASS IN CASE OF EMEGENCY
			controller_tools.Flush_status_messages(other_elevators_status_chan)
			fmt.Println("Disconnected peer status: ", disconnected_peer_status)
			fmt.Println("Connected_elevators_status in dropped peer: ", *connected_elevators_status)


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
	has_ever_connected *bool,
	just_booted *bool,
	state *utilities.State,
	current_elevator *elevator.Elevator,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- utilities.ControllerToElevatorMessage,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	ctrl_to_network_chan chan<- utilities.StatusMessage,
	net *network.Network,
) {

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
				var orderline [utilities.N_FLOORS][utilities.N_BUTTONS-1]bool
				orderline[new_order_floor][new_order_button] = true
				ctrl_to_elev_chan <- utilities.ControllerToElevatorMessage{Label: elevator.EB_to_string[elevator.EB_Disconnected], Orderline: orderline}
			}
		default:
			if net.Connection {
				if *has_ever_connected {
					*just_booted = false
				}
				*state = utilities.State_slave
				fmt.Println("Switching to slave mode")
				return
			}
		}
	}
}
