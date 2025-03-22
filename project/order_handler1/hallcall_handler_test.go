package order_assigner // vi brukre ikke main fordi vi ikke skal kjøre dette programmet, vi skal bare teste det.

import (
	"fmt"
	"main/order_handler1/generate_input"
	"testing"

	"main/elev_algo_go/elevator"
	"main/utilities"
)

func TestAssignHallRequests(t *testing.T) {

	sampleStatuses := []utilities.StatusMessage{
		{
			Label:         "does not matter",
			Controller_id: 0,
			Behaviour:     "idle",
			Floor:         2,
			Direction:     "up",
			Node_orders: [elevator.N_FLOORS][elevator.N_BUTTONS]bool{
				{false, false, false},
				{true, false, false},
				{false, false, true},
				{false, false, true},
			},
		},
		{
			Label:         "does not matter",
			Controller_id: 1,
			Behaviour:     "idle",
			Floor:         0,
			Direction:     "stop",
			Node_orders: [elevator.N_FLOORS][elevator.N_BUTTONS]bool{
				{false, false, false},
				{false, false, false},
				{false, false, false},
				{false, false, false},
			},
		},
		{
			Label:         "does not matter",
			Controller_id: 2,
			Behaviour:     "idle",
			Floor:         1,
			Direction:     "up",
			Node_orders: [elevator.N_FLOORS][elevator.N_BUTTONS]bool{
				{false, false, true},
				{false, false, false},
				{false, false, false},
				{false, false, false},
			},
		},
	}

	// := bruker man denne deklarer man og initaliserer man i samme step

	input := generate_input.GenerateAssignerInputFromStatusMessages(sampleStatuses)
	output, err := AssignHallRequests(input)
	if err != nil {
		t.Fatalf("AssignHallRequests failed: %v", err)
	}

	t.Logf("Output: %v", output)
	fmt.Printf("done")

	fmt.Printf("			HERE IS THE FINAL OUTPUT TO BE DISTRUBUTED: 			")

	OrderDistributionMessage(output)
}
