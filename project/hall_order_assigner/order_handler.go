package order_handler

import (
	"fmt"
	"main/elev_algo_go/elevator"
	"main/utilities"
	"bytes"
	"encoding/json"
	"os/exec"
	"strconv"
)


func redistribute_obstructed_service_orders(statuses []utilities.Status_message) []utilities.Status_message {
	var unobstructed_statuses []utilities.Status_message
	var all_service_orders [utilities.N_FLOORS][utilities.N_BUTTONS - 1]bool
	for _, status := range statuses {
		if status.Behaviour == elevator.EB_to_string[elevator.EB_Obstructed] {
			for floor := 0; floor < utilities.N_FLOORS; floor++ {
				for btn := 0; btn < utilities.N_BUTTONS-1; btn++ {
					all_service_orders[floor][btn] = all_service_orders[floor][btn] || status.Node_orders[floor][btn]
				}
			}
		} else {
			unobstructed_statuses = append(unobstructed_statuses, status)
		}
	}
	if len(unobstructed_statuses) > 0 {
		for floor := 0; floor < utilities.N_FLOORS; floor++ {
			for btn := 0; btn < utilities.N_BUTTONS-1; btn++ {
				unobstructed_statuses[0].Node_orders[floor][btn] = unobstructed_statuses[0].Node_orders[floor][btn] || all_service_orders[floor][btn]
			}
		}
	}
	return unobstructed_statuses
}

func build_assigner_input(statuses []utilities.Status_message) map[string]interface{} {
	hall_requests := make([][]bool, utilities.N_FLOORS) 
	for f := 0; f < utilities.N_FLOORS; f++ {
		hall_requests[f] = make([]bool, 2) 
	} 
	states := make(map[string]interface{}) 
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

func run_hall_assigner(input map[string]interface{}) (map[string]interface{}, error) {
	input_json, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("error marshalling input JSON: %v", err)
	}
	cmd := exec.Command("hall_order_assigner/hall_request_assigner", "-i", string(input_json))
	var output_buffer bytes.Buffer
	cmd.Stdout = &output_buffer
	cmd.Stderr = &output_buffer
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error output:", output_buffer.String()) 
		return nil, fmt.Errorf("error running hall_request_assigner: %v", err)
	}
	var output map[string]interface{}
	err = json.Unmarshal(output_buffer.Bytes(), &output)
	if err != nil {
		return nil, fmt.Errorf("error parsing output JSON: %v", err)
	}
	return output, nil
}

func create_order_distribution_message(raw_output map[string]interface{}) (utilities.Order_distribution_message, error) {
	var ODM utilities.Order_distribution_message 
	for i := 0; i < len(ODM.Orderlines); i++ {
		key := strconv.Itoa(i)
		orders, ok := raw_output[key].([]interface{})
		if !ok {
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
	return ODM, nil
}

func Order_handler(statuses []utilities.Status_message) utilities.Order_distribution_message {
	builder_input := redistribute_obstructed_service_orders(statuses)
	assigner_input := build_assigner_input(builder_input)
	assigner_output, err := run_hall_assigner(assigner_input)
	if err != nil {
		fmt.Println(err)
		return utilities.Order_distribution_message{}
	}
	ODM, err := create_order_distribution_message(assigner_output)
	if err != nil {
		fmt.Println(err)
		return utilities.Order_distribution_message{}
	}
	return ODM
}
