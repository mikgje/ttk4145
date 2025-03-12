package main_controller

// Interface between controller network and elevator
import (
	"fmt"
	"main/elev_algo_go/elevator"
	"main/elevio"
	"main/utilities"
	"main/controll/controller_modes"

	//For testing
	"main/elev_algo_go/timer"
)

type State int

var (
	controller_id 		int
	is_elevator_healthy bool  = true //Used to signal to network, the current elevator is not to be considered for new orders
	state               State = 0
	augmented_requests [utilities.N_FLOORS][utilities.N_BUTTONS]bool
	network_send_chan = make(chan utilities.StatusMessage)
	network_receive_status_chan = make(chan utilities.StatusMessage) //Unsure about the datatype of this channel
	network_receive_order_chan = make(chan utilities.OrderDistributionMessage)
	current_elevator 		  elevator.Elevator
)

const (
	state_normal State = iota
	state_backup
	state_primary
	state_disconnected
)

func TEMP_receive_network(network_receive_chan chan<- utilities.OrderDistributionMessage) {
	fmt.Println("Starting network receive function")
	var new_order_list = utilities.OrderDistributionMessage{
		Orderlines: [3][utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool{
			{
				{true, false},
				{false, true},
				{true, true},
				{false, false},
			},
			{
				{false, true},
				{true, false},
				{false, false},
				{true, true},
			},
			{
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
			network_receive_chan <- new_order_list
			go timer.Timer_start2(10, timer_chan)
			fmt.Println("Starting network receive timer in loop")
		}
	}
}

func Main_controller(controller_id int, elev_to_ctrl_chan <-chan elevator.Elevator, elev_to_ctrl_button_chan <-chan elevio.ButtonEvent, ctrl_to_elev_chan chan<- elevio.ButtonEvent) {
	/* Placeholder for network routines */
	// go TEMP_transmit_network(network_send_chan)
	go TEMP_receive_network(network_receive_order_chan)
	controller_id = 0
	/* End of placeholder */

	controller_state_machine(state, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan)

}

func controller_state_machine(state State, elev_to_ctrl_chan <-chan elevator.Elevator, elev_to_ctrl_button_chan <-chan elevio.ButtonEvent, ctrl_to_elev_chan chan<- elevio.ButtonEvent
	network_receive_chan <-chan utilities.OrderDistributionMessage, network_send_chan chan<- utilities.StatusMessage) {
	switch state {
	case state_normal:
		fmt.Println("Starting normal controller")
		controller_modes.Normal_controller(controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, network_receive_chan, network_send_chan)
	case state_backup:
		fmt.Println("Starting backup controller")
		controller_modes.Backup_controller(controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, network_receive_chan, network_send_chan)
	case state_primary:
		fmt.Println("Starting primary controller")
		controller_modes.Primary_controller(controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, network_receive_chan, network_send_chan)
	case state_disconnected:
		fmt.Println("Starting disconnected controller")
		controller_modes.Disconnected_controller(controller_id, elev_to_ctrl_chan, elev_to_ctrl_button_chan, ctrl_to_elev_chan, network_receive_chan, network_send_chan)
	}
}

