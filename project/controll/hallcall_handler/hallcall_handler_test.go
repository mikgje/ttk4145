package hallcall_handler  // vi brukre ikke main fordi vi ikke skal kjøre dette programmet, vi skal bare teste det.
// main brukes i det tilfellet at vi skal

import (
	"testing"
)




func TestAssignHallRequests(t *testing.T) {  //denn inputen brukes ikke i funksjonen, men er hva som faktisk skal brukes, 
	input := map[string]interface{}{ // algoritmen ser ut til å kunne ta inn ubregrenset antall heiser 
		"hallRequests": [][]bool{
			{false, false},
			{true, false},
			{false, false},
			{false, true},
		},
		"states": map[string]interface{}{
			"one": map[string]interface{}{
				"behaviour":   "moving",
				"floor":       2,
				"direction":   "up",
				"cabRequests": []bool{false, false, true, true},
			},
			"two": map[string]interface{}{
				"behaviour":   "idle",
				"floor":       0,
				"direction":   "stop",
				"cabRequests": []bool{false, false, false, false},
			},
			"three": map[string]interface{}{
				"behaviour":   "moving",
				"floor":       1,
				"direction":   "up",
				"cabRequests": []bool{true, false, false, false},
			},
		},
	}

	output, err := AssignHallRequests(input)  // her sakl det egentlig være t, eller en variabl som arstein larger, og mikeal sender. 
	if err != nil {
		t.Fatalf("AssignHallRequests failed: %v", err)
	}

	t.Logf("Output: %v", output)  // ✅ Logs output
}
