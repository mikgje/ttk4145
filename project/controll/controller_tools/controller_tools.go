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

// TODO: make scalable
func Extract_orderline(controller_id int, orderlines utilities.OrderDistributionMessage) [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool {
	switch controller_id {
	case 0:
		return orderlines.Orderlines[0]
	case 1:
		return orderlines.Orderlines[1]
	case 2:
		return orderlines.Orderlines[2]
	default:
		panic("Controller ID is not a valid ID")
	}
}
