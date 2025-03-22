package order_handler

import (
	"fmt"
	"main/utilities"
)

// This function builds an input for our order_assigner, so that the executable hall_request_assigner accepts it.
// The input is a slice of StatusMessages from utilities (this input is yet to be made)
func build_assigner_input_from_status_messages(statuses []utilities.StatusMessage) map[string]interface{} {

	hall_requests := make([][]bool, utilities.N_FLOORS) 
	for f := 0; f < utilities.N_FLOORS; f++ {
		hall_requests[f] = make([]bool, 2) 
	} 

	states := make(map[string]interface{}) 

	// based on the status of each elevator we create a map of all hall_requests and states
	for _, s := range statuses {
		cab_requests := make([]bool, utilities.N_FLOORS)

		for f := 0; f < utilities.N_FLOORS; f++ {
			hall_requests[f][0] = hall_requests[f][0] || s.Node_orders[f][0]
			hall_requests[f][1] = hall_requests[f][1] || s.Node_orders[f][1]
			
			cab_requests[f] = s.Node_orders[f][2]
		}

		key := fmt.Sprintf("%d", s.Controller_id)
		states[key] = map[string]interface{}{
			"behaviour":   s.Behaviour,
			"floor":       s.Floor,
			"direction":   s.Direction,
			"cabRequests": cab_requests,
		}
	}

	return map[string]interface{}{
		"hallRequests": hall_requests, 
		"states":        states,      
	}
}
