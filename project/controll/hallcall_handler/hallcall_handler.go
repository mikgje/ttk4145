package hallcall_handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

// AssignHallRequests prepares JSON input, runs the hall_request_assigner executable with `-i`, and returns the updated assignments.
func AssignHallRequests(input map[string]interface{}) (map[string]interface{}, error) {
	// Convert input to JSON
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("error marshalling input JSON: %v", err)
	}



	// Debugging: Print JSON input
	fmt.Println("Sending to hall_request_assigner:", string(inputJSON))


	
	// lager cmd
	cmd := exec.Command("./hall_request_assigner", "-i", string(inputJSON))

	// Capture output
	var outputBuffer bytes.Buffer
	cmd.Stdout = &outputBuffer
	cmd.Stderr = &outputBuffer

	fmt.Println("Running hall_request_assigner...")

	err = cmd.Run()
	if err != nil {
		fmt.Println("Error output:", outputBuffer.String()) // Print stderr
		return nil, fmt.Errorf("error running hall_request_assigner: %v", err)
	}

	// Debugging: Print raw output
	fmt.Println("Raw output from hall_request_assigner:", outputBuffer.String())

	// Parse JSON output
	var output map[string]interface{}
	err = json.Unmarshal(outputBuffer.Bytes(), &output)
	if err != nil {
		return nil, fmt.Errorf("error parsing output JSON: %v", err)
	}

	return output, nil
}