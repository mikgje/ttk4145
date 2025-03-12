package controller_modes

import (
	"fmt"
	"main/elev_algo_go/elevator"
	"main/utilities"
	"main/elevio"
	"main/controll/controller_tools"
)

func Normal_controller(controller_id int, elev_to_ctrl_chan <-chan elevator.Elevator, elev_to_ctrl_button_chan <-chan elevio.ButtonEvent, 
	ctrl_to_elev_chan chan<- elevio.ButtonEvent, network_receive_order_chan <-chan utilities.OrderDistributionMessage, network_send_chan chan<- utilities.StatusMessage) {
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

			if new_order_floor == utilities.UNHEALTHY_FLAG { //This funcitonality is here to stay, ish.
				fmt.Println("Elevator is unhealthy")
				is_elevator_healthy = false

			} else if new_order_floor == utilities.HEALTHY_FLAG {
				fmt.Println("Elevator is healthy")
				is_elevator_healthy = true

			} else if new_order_button == elevio.BT_Cab { 
				if !is_elevator_healthy { //Skal vi alltid sende cabcalls lokalt? Uavhengig av om heisen er healthy?
					fmt.Println("Elevator is unhealthy, not sending order")
					break
				}
				ctrl_to_elev_chan <- new_order

			} else { //Sending to network
				augmented_requests = controller_tools.Augment_request_array(current_elevator.Requests, new_order);
				status_message := utilities.StatusMessage{Controller_id: 0, Behaviour: "normal", Floor: 2, Direction: "up", Node_orders: augmented_requests}
				network_send_chan <- status_message
			}
		case msg := <-network_receive_order_chan:
			new_orders := controller_tools.Extract_orderline(controller_id, msg)
			fmt.Println("Received new orders from network")
			for floor := 0; floor < utilities.N_FLOORS; floor++ {
				for button := 0; button < utilities.N_BUTTONS-1; button++ {
					if new_orders[floor][button] {
						new_order := elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(button)}
						ctrl_to_elev_chan <- new_order
					}
				}
			}
		}
	}
}

func Backup_controller(controller_id int, elev_to_ctrl_chan <-chan elevator.Elevator, elev_to_ctrl_button_chan <-chan elevio.ButtonEvent, 
	ctrl_to_elev_chan chan<- elevio.ButtonEvent, network_receive_order_chan <-chan utilities.OrderDistributionMessage, network_send_chan chan<- utilities.StatusMessage) {

}

func Primary_controller(controller_id int, elev_to_ctrl_chan <-chan elevator.Elevator, elev_to_ctrl_button_chan <-chan elevio.ButtonEvent,
	ctrl_to_elev_chan chan<- elevio.ButtonEvent, network_receive_order_chan <-chan utilities.OrderDistributionMessage, network_send_chan chan<- utilities.StatusMessage) {

}

func Disconnected_controller(controller_id int, elev_to_ctrl_chan <-chan elevator.Elevator, elev_to_ctrl_button_chan <-chan elevio.ButtonEvent, 
	ctrl_to_elev_chan chan<- elevio.ButtonEvent, network_receive_order_chan <-chan utilities.OrderDistributionMessage, network_send_chan chan<- utilities.StatusMessage) {

}
