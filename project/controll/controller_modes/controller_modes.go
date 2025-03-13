package controller_modes

import (
	"fmt"
	"main/controll/controller_tools"
	"main/elev_algo_go/elevator"
	"main/elevio"
	"main/utilities"
)

func Normal_controller(controller_id int,
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	network_receive_order_chan <-chan utilities.OrderDistributionMessage,
	network_send_chan chan<- utilities.StatusMessage) {

	var current_elevator elevator.Elevator
	var augmented_requests [utilities.N_FLOORS][utilities.N_BUTTONS]bool
	var is_elevator_healthy bool = true
	for {
		select {
		case msg := <-elev_to_ctrl_chan:
			current_elevator = msg
		case msg := <-elev_to_ctrl_button_chan:
			new_order_floor := msg.Floor
			new_order_button := msg.Button
			new_order := elevio.ButtonEvent{Floor: new_order_floor, Button: new_order_button}
			fmt.Printf("New order from local elevator: Floor %d, Button %s\n", new_order_floor, elevator.Button_to_string[new_order_button])
			if new_order_button == elevio.BT_Cab {
				if !is_elevator_healthy { //Skal vi alltid sende cabcalls lokalt? Uavhengig av om heisen er healthy?
					fmt.Println("Elevator is unhealthy, not sending order")
					break
				}
				ctrl_to_elev_cab_chan <- new_order

			} else { //Sending to network
				augmented_requests = controller_tools.Augment_request_array(current_elevator.Requests, new_order)
				status_message := utilities.StatusMessage{Controller_id: 0, Behaviour: "normal", Floor: 2, Direction: "up", Node_orders: augmented_requests}
				network_send_chan <- status_message
			}
		case msg := <-network_receive_order_chan:
			new_orders := controller_tools.Extract_orderline(controller_id, msg)
			fmt.Println("Received new orders from network")
			ctrl_to_elev_chan <- new_orders
		}
	}
}

func Backup_controller(controller_id int, 
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	network_receive_order_chan <-chan utilities.OrderDistributionMessage,
	network_send_chan chan<- utilities.StatusMessage) {

}

func Primary_controller(controller_id int, 
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	network_receive_order_chan <-chan utilities.OrderDistributionMessage,
	network_send_chan chan<- utilities.StatusMessage) {

}

func Disconnected_controller(controller_id int, 
	elev_to_ctrl_chan <-chan elevator.Elevator,
	elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool,
	ctrl_to_elev_cab_chan chan<- elevio.ButtonEvent,
	network_receive_order_chan <-chan utilities.OrderDistributionMessage,
	network_send_chan chan<- utilities.StatusMessage) {

}
