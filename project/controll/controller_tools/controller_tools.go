package controller_tools

import (
	"main/elevio"
	"main/utilities"
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

func Flush_status_messages(other_elevatos_status <-chan utilities.StatusMessage) {
	for i := 0; i < 1000; i++ {
		<-other_elevatos_status
	}
	// fmt.Println("Flushed status messages")
}
