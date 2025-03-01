package main

// Interface between controller network and elevator
import (
	"fmt"
	"main/elev_algo_go/elevator"
	"main/elevio"

	//For testing
	"main/elev_algo_go/timer"
)

type State int
/*
Here a datatype called StatusMessage will be defined, but is not yet implemented.
This datatype will be used to send messages to the network module.
*/

type OrderDistributionMessage struct{
	orderline0 struct{
		new_hall_orders [elevator.N_FLOORS][elevator.N_BUTTONS-1]bool
	}
	orderline1 struct{
		new_hall_orders [elevator.N_FLOORS][elevator.N_BUTTONS-1]bool
	}
	orderline2 struct{
		new_hall_orders [elevator.N_FLOORS][elevator.N_BUTTONS-1]bool
	}
}

var (
	controller_id int;
	is_elevator_healthy bool = true //Used to signal to network, the current elevator is not to be considered for new orders
	state State = 0
	// augmented_requests [elevator.N_FLOORS][elevator.N_BUTTONS]bool
	// network_send_chan = make(chan StatusMessage)
	// network_receive_status_chan = make(chan StatusMessage) //Unsure about the datatype of this channel
	network_receive_order_chan = make(chan OrderDistributionMessage)
)

const (
	state_normal State = iota
	state_backup
	state_primary
	state_disconnected
)

// func augment_request_array(elevator_service_orders [elevator.N_FLOORS][elevator.N_BUTTONS]bool, new_order elevio.ButtonEvent) [elevator.N_FLOORS][elevator.N_BUTTONS]bool {
// 	augmented_requests = elevator_service_orders
// 	augmented_requests[new_order.Floor][new_order.Button] = true
// 	return augmented_requests

// }


// Extracts the correct order line from the masters order list based on controller id
func extract_orderline(orderlines OrderDistributionMessage) [elevator.N_FLOORS][elevator.N_BUTTONS-1]bool {
	switch controller_id{
	case 0:
		return orderlines.orderline0.new_hall_orders
	case 1:
		return orderlines.orderline1.new_hall_orders
	case 2:
		return orderlines.orderline2.new_hall_orders
	default:
		panic("Controller ID is not a valid ID")
	}
}

// This is a temporary test function to simulate recieving orders over the network.
func TEMP_receive_network(network_receive_chan chan<- OrderDistributionMessage){
	fmt.Println("Starting network receive function")
	var new_order_list = OrderDistributionMessage{
		orderline0: struct {
			new_hall_orders [elevator.N_FLOORS][elevator.N_BUTTONS-1]bool
		}{
			new_hall_orders: [elevator.N_FLOORS][elevator.N_BUTTONS-1]bool{
				{true, false},
				{false, true},
				{true, true},
				{false, false},
			},
		},
		orderline1: struct {
			new_hall_orders [elevator.N_FLOORS][elevator.N_BUTTONS-1]bool
		}{
			new_hall_orders: [elevator.N_FLOORS][elevator.N_BUTTONS-1]bool{
				{false, true},
				{true, false},
				{false, false},
				{true, true},
			},
		},
		orderline2: struct {
			new_hall_orders [elevator.N_FLOORS][elevator.N_BUTTONS-1]bool
		}{
			new_hall_orders: [elevator.N_FLOORS][elevator.N_BUTTONS-1]bool{
				{true, true},
				{false, false},
				{true, false},
				{false, true},
			},
		},
	}
	timer_chan := make(chan bool)
	go timer.Timer_start2(10, timer_chan)
	fmt.Println("Starting network receive timer")
	for {
		select {
		case <-timer_chan:
			network_receive_chan<- new_order_list	
			timer.Timer_start2(10, timer_chan)	
			fmt.Println("Starting network receive timer in loop")
		}
	}
}

/*
The functions below this comment are used to realise the controller logic and operation.
The controller has several different states to operate in, where each state has a slightly different behaviour.
*/


//Function to run the controller state machine and network modules
func main_controller() {
	/* Placeholder for network routines */
	// go TEMP_transmit_network(network_send_chan)
	go TEMP_receive_network(network_receive_order_chan)
	controller_id = 0;
	/* End of placeholder */


	controller_state_machine(state)
	
}

//Function to run the controller state machine
func controller_state_machine(state State) {
	switch state {
	case state_normal:
		normal_controller()
	case state_backup:
		backup_controller()
	case state_primary:
		primary_controller()
	case state_disconnected:
		disconnected_controller()
	}
}


func normal_controller() {

	/*
	In this state the controller works purely as a slave and is not concerned with the status of the other nodes.
	It will send status messages and recieve order messages.
	*/

	for {
		select {
		case msg := <-elev_to_ctrl_chan:
			new_order_floor := msg.Floor
			new_order_button := msg.Button
			new_order := elevio.ButtonEvent{Floor: new_order_floor, Button: new_order_button}
			fmt.Printf("New order from local elevator: Floor %d, Button %s\n", new_order_floor, elevator.Button_to_string[new_order_button])

			if new_order_floor == UNHEALTHY_FLAG { //This funcitonality is here to stay, ish.
				fmt.Println("Elevator is unhealthy")
				is_elevator_healthy = false

			} else if new_order_floor == HEALTHY_FLAG {
				fmt.Println("Elevator is healthy")
				is_elevator_healthy = true

			//TODO implement communication with network and arbitration before sending any orders to elevator

			} else if new_order_button == elevio.BT_Cab {
				if !is_elevator_healthy {
					fmt.Println("Elevator is unhealthy, not sending order")
					break
				}	
				ctrl_to_elev_chan <- new_order

			// } else { //Sending to network
			// 	/* BELOW THIS LINE IS NOT WORKING, TEMPORARY PLACEHOLDER */
			// 	augmented_requests = augment_request_array(elevator_service_orders, new_order);	
			// 	status_message := StatusMessage{controller_id: 0, behaviour: "normal", floor: 2, direction: "up", node_orders: augmented_requests}
			// 	network_send_chan <- status_message
			// 	ctrl_to_elev_chan <- new_order
			// 	/* END OF PLACEHOLDER */
			}
		case msg := <-network_receive_order_chan: //Recieve new orders from network and sending them to elevator in sequence.
			new_orders := extract_orderline(msg)
			fmt.Println("Received new orders from network")
			for floor := 0; floor < elevator.N_FLOORS; floor++ {
				for button := 0; button < elevator.N_BUTTONS-1; button++ {
					if new_orders[floor][button] {
						new_order := elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(button)}
						ctrl_to_elev_chan <- new_order
					}
				}
			}

		}
	}
}

func backup_controller() {
	/*
	This state allows the controller to both work as a normal slave on the network, while monitoring the status of the primary controller.
	If the controller detects that the primary controller is not functioning, it will take over as the primary controller.
	*/
}

func primary_controller() {
	/*
	This state allows the controller to work as the primary controller on the network.
	The primary controller is responsible for distributing orders to the other controllers on the network.
	When starting up as primary, the controller will choose its backup controller.
	*/

}

func disconnected_controller() {
	/*
	The controller will go to this state when it detects it has lost connection to the network.
	In this state the controller will try to reconnect to the network.
	While disconnected, the elevator will only serve cab calls and ignore hall calls from its panel.
	*/


	for {
		select {
		case msg := <-elev_to_ctrl_chan:
			new_order_floor := msg.Floor
			new_order_button := msg.Button
			new_order := elevio.ButtonEvent{Floor: new_order_floor, Button: new_order_button}
			fmt.Printf("New order from local elevator: Floor %d, Button %s\n", new_order_floor, elevator.Button_to_string[new_order_button])

			if new_order_floor == UNHEALTHY_FLAG { //This funcitonality is here to stay, ish.
				fmt.Println("Elevator is unhealthy")
				is_elevator_healthy = false

			} else if new_order_floor == HEALTHY_FLAG {
				fmt.Println("Elevator is healthy")
				is_elevator_healthy = true

			} else if new_order_button == elevio.BT_Cab {
				if !is_elevator_healthy {
					fmt.Println("Elevator is unhealthy, not sending order")
					break
				}	
				ctrl_to_elev_chan <- new_order
			}
		default:
			//Check internet connection
			//If connection is reestablished, change state to normal
			//If connection is not reestablished, keep checking
		}
	}

}

