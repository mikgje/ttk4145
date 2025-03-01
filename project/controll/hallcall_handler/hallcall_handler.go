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

	// Execute hall_request_assigner with `-i` argument
	cmd := exec.Command("./hall_request_assigner", "-i", string(inputJSON)) // ✅ Fix: Pass JSON as argument

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

// The AssignHallRequests function prepares the JSON input, runs the hall_request_assigner executable, and returns the updated hall request assignments.

//this code is used to run the hall_request_assigner executable, what it does is that it takes the input map
// and converts it to a JSON format, then it runs the hall_request_assigner executable 
// and passes the JSON input to it. The output of the executable is captured and parsed to a map and returned.