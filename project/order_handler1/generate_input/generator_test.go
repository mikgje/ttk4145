package generate_input

import (
	//"encoding/json"
	"fmt"
	"testing"

	"main/elev_algo_go/elevator"
	"main/utilities"
)

var floor int

func TestGenerateAssignerInputFromStatusMessages(t *testing.T) {
	sampleStatuses := []utilities.StatusMessage{
		{
			Label:         "does not matter",
			Controller_id: 0,
			Behaviour:     "",
			Floor:         floor,
			Direction:     "up",
			Node_orders: [elevator.N_FLOORS][elevator.N_BUTTONS]bool{
				{true, false, false},
				{true, false, false},
				{false, true, true},
				{false, false, true},
			},
		},
		{
			Label:         "does not matter",
			Controller_id: 1,
			Behaviour:     "",
			Floor:         0,
			Direction:     "stop",
			Node_orders: [elevator.N_FLOORS][elevator.N_BUTTONS]bool{
				{false, false, false},
				{false, false, false},
				{true, false, false},
				{false, false, false},
			},
		},
		{
			Label:         "does not matter",
			Controller_id: 2,
			Behaviour:     "",
			Floor:         1,
			Direction:     "up",
			Node_orders: [elevator.N_FLOORS][elevator.N_BUTTONS]bool{
				{false, true, true},
				{false, false, false},
				{false, false, false},
				{false, false, false},
			},
		},
	}

	input := GenerateAssignerInputFromStatusMessages(sampleStatuses)

	fmt.Println(input)

}
