package controller_tools

import (
	"main/utilities"
	"main/elevio"

	//FOR TESTING
	// "fmt"
)

func Augment_request_array(elevator_service_orders [utilities.N_FLOORS][utilities.N_BUTTONS]bool, new_order elevio.ButtonEvent) [utilities.N_FLOORS][utilities.N_BUTTONS]bool {
	augmented_requests := elevator_service_orders
	augmented_requests[new_order.Floor][new_order.Button] = true

	// fmt.Println("Augmented requests:")
	// fmt.Println(augmented_requests)
	// fmt.Println("-------------------")

	return augmented_requests
}

func Extract_orderline(controller_id int, odm utilities.OrderDistributionMessage) [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool {
	if controller_id < utilities.N_ELEVS {
		return odm.Orderlines[controller_id]
	} else {
		panic("Controller id out of bounds")
	}
}

func Extract_other_orderlines(controller_id int, odm utilities.OrderDistributionMessage) [][utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool {
	other_orderlines := make([][utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool, 0, utilities.N_ELEVS-1)
	for i := 0; i < utilities.N_ELEVS; i++ {
		if i != controller_id {
			other_orderlines = append(other_orderlines, odm.Orderlines[i])
		}
	}
	return other_orderlines
}

func Flush_status_messages(status_chan <-chan utilities.StatusMessage) {
	for i := 0; i < 1000; i++ {
		<-status_chan
	}

}

func Update_confirmation(
	old_confirmation	[][utilities.N_FLOORS][utilities.N_BUTTONS]bool,
	odm 				utilities.OrderDistributionMessage, 
	statuses 			[]utilities.StatusMessage,
) (
	[][utilities.N_FLOORS][utilities.N_BUTTONS]bool,
	[]bool,
) {
	new_confirmation := make([][utilities.N_FLOORS][utilities.N_BUTTONS]bool, len(statuses))
	node_confirmation := make([]bool, len(statuses))
	for i := 0; i < len(statuses); i++ {
		node_confirmation[i] = true
		for j := 0; j < utilities.N_FLOORS; j++ {
			for k := 0; k < utilities.N_BUTTONS; k++ {
				new_confirmation[i][j][k] = true
				if k < utilities.N_BUTTONS-1 && odm.Orderlines[i][j][k] && !statuses[i].Node_orders[j][k] {
					new_confirmation[i][j][k] = false
				}
				if i < len(old_confirmation) {
					new_confirmation[i][j][k] = old_confirmation[i][j][k] || new_confirmation[i][j][k]
				}
				// Quick fix
				new_confirmation[i][0][0] = true
				new_confirmation[i][utilities.N_FLOORS-1][1] = true
				node_confirmation[i] = node_confirmation[i] && new_confirmation[i][j][k]
			}
		}
	}
	return new_confirmation, node_confirmation
}
