package controller_modes

import (
	"fmt"
	"main/controll/controller_tools"
	"main/controll/hall_order_assigner"
	"main/elev_algo_go/elevator"
	"main/elevio"
	"main/network"
	"main/utilities"
)

func base_controller(
	status_message *utilities.StatusMessage, 
	current_elevator *elevator.Elevator, 
	controller_id *int,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- utilities.ControllerToElevatorMessage,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	ctrl_to_network_chan chan<- utilities.StatusMessage,
	network_to_ctrl_chan <-chan utilities.OrderDistributionMessage,
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

		case msg := <-network_to_ctrl_chan /*The channel that supplies the ODM*/ :
			new_orders := controller_tools.Extract_orderline(*controller_id, msg)
			if *controller_id != 0 {
				fmt.Println("New orders: ", new_orders)
			}

			for floor := 0; floor < utilities.N_FLOORS; floor++ {
				for btn := 0; btn < utilities.N_BUTTONS-1; btn++ {
					status_message.Node_orders[floor][btn] = new_orders[floor][btn]
				}
			}

			other_orderlines := controller_tools.Extract_other_orderlines(*controller_id, msg)
			ctrl_to_elev_chan <- utilities.ControllerToElevatorMessage{Orderline: new_orders, Other_orderlines: other_orderlines}
		case <-kill_base_ctrl_chan:
			return
		}
		//TODO: Fix order_handler error with unsupported floor
		if current_elevator.Floor != -1 {
			ctrl_to_network_chan <- *status_message
		}
	}
}

func Slave(
	state *utilities.State, 
	current_elevator *elevator.Elevator,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- utilities.ControllerToElevatorMessage,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	ctrl_to_network_chan chan<- utilities.StatusMessage,
	network_to_ctrl_chan <-chan utilities.OrderDistributionMessage,
	net *network.Network,
	) {

	kill_base_ctrl_chan := make(chan bool)

	var status_message = utilities.StatusMessage{Controller_id: net.Ctrl_id, Behaviour: elevator.EB_to_string[current_elevator.Behaviour],
		Floor: current_elevator.Floor, Direction: elevator.Dirn_to_string[current_elevator.Dirn], Node_orders: current_elevator.Requests}
		
		go base_controller(&status_message, current_elevator, &net.Ctrl_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan, network_to_ctrl_chan, kill_base_ctrl_chan)
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

func Master(
	state *utilities.State, 
	current_elevator *elevator.Elevator,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- utilities.ControllerToElevatorMessage,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	ctrl_to_network_chan chan<- utilities.StatusMessage,
	network_to_ctrl_chan <-chan utilities.OrderDistributionMessage,
	ODM_to_network_chan chan<- utilities.OrderDistributionMessage,
	other_elevators_status <-chan utilities.StatusMessage,
	dropped_peer_chan <-chan utilities.StatusMessage,
	net *network.Network,
	) {

	var prev_odm utilities.OrderDistributionMessage
	var button_confirmations = make([][utilities.N_FLOORS][utilities.N_BUTTONS]bool, 0, utilities.N_ELEVS)
	var node_confirmations = make([]bool, 0, utilities.N_ELEVS)

	var connected_elevators_status = make(map[int]utilities.StatusMessage)
	var kill_base_ctrl_chan = make(chan bool)
	var status_message = utilities.StatusMessage{Controller_id: net.Ctrl_id, Behaviour: elevator.EB_to_string[current_elevator.Behaviour],
		Floor: current_elevator.Floor, Direction: elevator.Dirn_to_string[current_elevator.Dirn], Node_orders: current_elevator.Requests}

	go base_controller(&status_message, current_elevator, &net.Ctrl_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, ctrl_to_elev_cab_chan, ctrl_to_network_chan, network_to_ctrl_chan, kill_base_ctrl_chan)

	for {
		select {
		// TODO: 

		case msg := <-other_elevators_status:

			if msg.Controller_id != utilities.Default_id {
				connected_elevators_status[msg.Controller_id] = msg
			}

			outer_loop:
			for i := 0; i < utilities.N_ELEVS; i++ {
				select{
				case msg := <-other_elevators_status:
					if msg.Controller_id != utilities.Default_id {
						connected_elevators_status[msg.Controller_id] = msg
					}
				default:

					break outer_loop
				}
			}
			
			status_slice := make([]utilities.StatusMessage, 0, len(connected_elevators_status))			
			
			for _, status := range connected_elevators_status {
				status_slice = append(status_slice, status)

			}
			//TODO: must be reset
			button_confirmations, node_confirmations = controller_tools.Update_confirmation(button_confirmations, prev_odm, status_slice)
//			fmt.Println(button_confirmations)

			status_to_order_handler := make([]utilities.StatusMessage, 0, len(connected_elevators_status))
			
			for _, status := range connected_elevators_status {
				if status.Controller_id < len(node_confirmations) {
					if node_confirmations[status.Controller_id] {
						status_to_order_handler = append(status_to_order_handler, status)
					} else {
						status.Behaviour = elevator.EB_to_string[elevator.EB_Unhealthy]
						status_to_order_handler = append(status_to_order_handler, status)
					}
				}
			}
			//fmt.Println("status_to_order_handler")
			//fmt.Println(status_to_order_handler)
			all_unhealthy := true
			for i:=0; i<len(status_to_order_handler); i++ {
				all_unhealthy = all_unhealthy && (status_to_order_handler[i].Behaviour == elevator.EB_to_string[elevator.EB_Unhealthy])
			}
			all_conf := true
			for i:=0; i<len(node_confirmations); i++ {
				all_conf = all_conf && node_confirmations[i]
			}
			//fmt.Println("AllUn", all_unhealthy)
			if !all_unhealthy && all_conf {
			new_odm := order_handler.Order_handler(status_to_order_handler)
			if new_odm != prev_odm {
				fmt.Println("I am resetting")
//				fmt.Println(new_odm)
//				fmt.Println("------------------------")
//				fmt.Println(prev_odm)
//				fmt.Println("------------------------")
				button_confirmations = make([][utilities.N_FLOORS][utilities.N_BUTTONS]bool, 0, utilities.N_ELEVS)

				ODM_to_network_chan <- new_odm
				for  _, status := range connected_elevators_status {
					if node_confirmations[status.Controller_id] {
						prev_odm.Orderlines[status.Controller_id] = new_odm.Orderlines[status.Controller_id]
					}
				}
			}
				// BREAK GLASS IN CASE OF EMEGENCY
//				controller_tools.Flush_status_messages(other_elevators_status)
			}
		case msg := <-dropped_peer_chan:
			delete(connected_elevators_status, msg.Controller_id)

		default:
			// status_to_order_handler := make([]utilities.StatusMessage, 0, 1)
			// status_to_order_handler = append(status_to_order_handler, status_message)
			// new_odm := order_handler.Order_handler(status_to_order_handler)
			// if new_odm != prev_odm {
			// 	fmt.Println("Sending new ODM to network")
			// 	ODM_to_network_chan <- new_odm
			// 	prev_odm = new_odm
			// }

		}
		if !net.Connection {
			*state = utilities.State_disconnected
			kill_base_ctrl_chan <- true
			fmt.Println("Switching to disconnected mode")
			return
		}
	}
}

func Disconnected(
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
