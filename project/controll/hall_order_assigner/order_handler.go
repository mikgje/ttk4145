package order_handler

import (
	"fmt"
	"main/utilities"

	"bytes"
	"encoding/json"
	"os/exec"
	"strconv"

)


func incorporate_unhealthy_orders(statuses []utilities.StatusMessage) []utilities.StatusMessage {
	var healthy_statuses []utilities.StatusMessage
	// Bruker N_BUTTONS-1 for å ekskludere cab calls
	var aggregated_orders [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool

	// vi går gjennom alle statusene og aggreger hall-orders fra de "unhealthy" statusene
	for _, status := range statuses {
		if status.Behaviour == "unhealthy" {
			for floor := 0; floor < utilities.N_FLOORS; floor++ {
				for btn := 0; btn < utilities.N_BUTTONS-1; btn++ {
					aggregated_orders[floor][btn] = aggregated_orders[floor][btn] || status.Node_orders[floor][btn]
				}
			}
		} else {
			healthy_statuses = append(healthy_statuses, status)
		}
	}

	// tl slutt legger vi de aggregerte hall-orders til den første gyldige (ikke-unhealthy) statusen
	if len(healthy_statuses) > 0 {
		for floor := 0; floor < utilities.N_FLOORS; floor++ {
			for btn := 0; btn < utilities.N_BUTTONS-1; btn++ {
				healthy_statuses[0].Node_orders[floor][btn] = healthy_statuses[0].Node_orders[floor][btn] || aggregated_orders[floor][btn]
			}
		}
	}

	return healthy_statuses
}



func build_assigner_input_from_status_messages(statuses []utilities.StatusMessage) map[string]interface{} {

	hall_requests := make([][]bool, utilities.N_FLOORS) 
	for f := 0; f < utilities.N_FLOORS; f++ {
		hall_requests[f] = make([]bool, 2) 
	} 

	states := make(map[string]interface{}) 

	// based on the status of each elevator we create a map of all hall_requests and states
	for _, status := range statuses {
		cab_requests := make([]bool, utilities.N_FLOORS)

		for f := 0; f < utilities.N_FLOORS; f++ {
			hall_requests[f][0] = hall_requests[f][0] || status.Node_orders[f][0]
			hall_requests[f][1] = hall_requests[f][1] || status.Node_orders[f][1]
			
			cab_requests[f] = status.Node_orders[f][2]
		}

		key := fmt.Sprintf("%d", status.Controller_id)
		states[key] = map[string]interface{}{
			"behaviour":   status.Behaviour,
			"floor":       status.Floor,
			"direction":   status.Direction,
			"cabRequests": cab_requests,
		}
	}

	return map[string]interface{}{
		"hallRequests": hall_requests, 
		"states":        states,      
	}
}






// this function runs the executable hall_request_assigner. It's intended input is created and explained in build_assinger_input
func assign_hall_requests(input map[string]interface{}) (map[string]interface{}, error) {

	// to JSON format
	input_json, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("error marshalling input JSON: %v", err)
	}

	//fmt.Println("Sending to hall_request_assigner:", string(input_json))

	// cmd is our executable, where the orders are truly assigned
	cmd := exec.Command("controll/hall_order_assigner/hall_request_assigner", "-i", string(input_json))

	var output_buffer bytes.Buffer
	cmd.Stdout = &output_buffer
	cmd.Stderr = &output_buffer

	//fmt.Println("Running hall_request_assigner...")

	err = cmd.Run()
	if err != nil {
		fmt.Println("Error output:", output_buffer.String()) 
		return nil, fmt.Errorf("error running hall_request_assigner: %v", err)
	}

	// fmt.Println("Raw output from hall_request_assigner:", output_buffer.String())

	var output map[string]interface{}
	// back to Go from JSON format
	err = json.Unmarshal(output_buffer.Bytes(), &output)
	if err != nil {
		return nil, fmt.Errorf("error parsing output JSON: %v", err)
	}

	return output, nil
}






func order_distribution_message(raw_output map[string]interface{}) (utilities.OrderDistributionMessage, error) {
	var ODM utilities.OrderDistributionMessage 
	
	for i := 0; i < len(ODM.Orderlines); i++ {
		key := strconv.Itoa(i)
		orders, ok := raw_output[key].([]interface{})
		if !ok {
			// key missing(i.e an ): we let the base value (false) remain.
			continue
		}

		for j := 0; j < utilities.N_FLOORS && j < len(orders); j++ {
			floor_orders, ok := orders[j].([]interface{})
			if !ok {
				return ODM, fmt.Errorf("unexpected type for key %s, floor %d", key, j)
			}

			for k := 0; k < utilities.N_BUTTONS-1 && k < len(floor_orders); k++ {
				state, ok := floor_orders[k].(bool)
				if !ok {
					return ODM, fmt.Errorf("unexpected type for key %s, floor %d, button %d", key, j, k)
				}
				ODM.Orderlines[i][j][k] = state
			}
		}
	}

//	fmt.Println(ODM)
	return ODM, nil
}






func Order_handler(statuses []utilities.StatusMessage) utilities.OrderDistributionMessage {
	builder_input := incorporate_unhealthy_orders(statuses)
	assigner_input := build_assigner_input_from_status_messages(builder_input)
	assigner_outut, err := assign_hall_requests(assigner_input)
	if err != nil {
		fmt.Println(err)
		return utilities.OrderDistributionMessage{}
	}
	ODM, err := order_distribution_message(assigner_outut)
	if err != nil {
		fmt.Println(err)
		return utilities.OrderDistributionMessage{}
	}
	return ODM
}
