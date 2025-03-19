package order_handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)


// this function runs the executable hall_request_assigner. It's intended input is created and explained in build_assinger_input
func assign_hall_requests(input map[string]interface{}) (map[string]interface{}, error) {

	// to JSON format
	input_json, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("error marshalling input JSON: %v", err)
	}

	fmt.Println("Sending to hall_request_assigner:", string(input_json))

	// cmd is our executable, where the orders are truly assigned
	cmd := exec.Command("./hall_request_assigner", "-i", string(input_json))

	var output_buffer bytes.Buffer
	cmd.Stdout = &output_buffer
	cmd.Stderr = &output_buffer

	fmt.Println("Running hall_request_assigner...")

	err = cmd.Run()
	if err != nil {
		fmt.Println("Error output:", output_buffer.String()) 
		return nil, fmt.Errorf("error running hall_request_assigner: %v", err)
	}

	fmt.Println("Raw output from hall_request_assigner:", output_buffer.String())

	var output map[string]interface{}
	// back to Go from JSON format
	err = json.Unmarshal(output_buffer.Bytes(), &output)
	if err != nil {
		return nil, fmt.Errorf("error parsing output JSON: %v", err)
	}

	return output, nil
}
