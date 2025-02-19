package main

// Interface between controller network and elevator
import (
	"fmt"
	"main/elev_algo_go/elevator"
	"main/elevio"
)

var is_elevator_healthy bool = true //Used to signal to network, the current elevator is not to be considered for new orders

func main_controller() {

	for {
		select {
		case msg := <-elev_to_ctrl_chan:
			new_order_floor := msg.Floor
			new_order_button := msg.Button
			fmt.Printf("New order from local elevator: Floor %d, Button %s\n", new_order_floor, elevator.Button_to_string[new_order_button])

			//TODO implement communication with network and arbitration before sending any orders to elevator

			if new_order_floor == UNHEALTHY_FLAG { //This funcitonality is here to stay, ish.
				fmt.Println("Elevator is unhealthy")
				is_elevator_healthy = false

			} else if new_order_floor == HEALTHY_FLAG {
				fmt.Println("Elevator is healthy")
				is_elevator_healthy = true

			} else { //This is temporary, and will be replaced by network communication
				if !is_elevator_healthy {
					fmt.Println("Elevator is not healthy, not sending order")
					break
				}				
				new_order := elevio.ButtonEvent{Floor: new_order_floor, Button: new_order_button}
				fmt.Println("Sending order to elevator")
				ctrl_to_elev_chan <- new_order
			}
		}
	}
}
