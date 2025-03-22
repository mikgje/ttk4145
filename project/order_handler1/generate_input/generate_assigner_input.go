package generate_input

import (
	//"main/elev_algo_go/fsm"
	"fmt"
	"main/elev_algo_go/elevator"
	"main/utilities"
)

type StatusMessage struct {
	Label         string
	Controller_id int
	Behaviour     string
	Floor         int
	Direction     string
	Node_orders   [elevator.N_FLOORS][elevator.N_BUTTONS]bool
}

type OrderDistributionMessage struct {
	Label      string
	Orderlines [3][elevator.N_FLOORS][elevator.N_BUTTONS - 1]bool
}

// så level for å skille i mellom de to, men vi vet jo at det første elementet i statusmessage er en int, mens det første elementet i Distrubion s

// Dette er en funksjon, som tar inn en slice (dynamisk liste på en måte), og returnerer et map, som er akkurat hva vi sender inn til hall_call_assigner
func GenerateAssignerInputFromStatusMessages(statuses []utilities.StatusMessage) map[string]interface{} {

	// først lager vi en placeholder for hallRequests som vi skal fylle inn ifra StatusMessage

	// vi lager N_FLOORS Slices, som inneholder indre bool slices
	hallRequests := make([][]bool, elevator.N_FLOORS) // --> [nil, nil, nil, nil], hvor hver nil forventes til å være en []bool
	for f := 0; f < elevator.N_FLOORS; f++ {
		hallRequests[f] = make([]bool, 2) // setter hver nil til en []bool med 2 indre elementer
	} // --> [[false, false],[false, false],[false, false],[false, false],]

	states := make(map[string]interface{}) // må lage denne
	// altså tar vi for oss hver av heisene s --> først zero: for hver etasje i denne heisen ser vi på om hall down og up er trykket
	for _, s := range statuses {
		cabRequests := make([]bool, elevator.N_FLOORS)

		for f := 0; f < elevator.N_FLOORS; f++ {
			// vi vet vi har en 4*3 matrise, her tar vi for oss de to første kolonnene og sjekker om button og floor er true
			hallRequests[f][0] = hallRequests[f][0] || s.Node_orders[f][0]
			hallRequests[f][1] = hallRequests[f][1] || s.Node_orders[f][1]
			// hallRequests definert over er allerede satt til false, her
			// ---------------------------------- Her er vi ferdig med å lage hallrequest delen av input  -------------------------------------
			cabRequests[f] = s.Node_orders[f][2]
		}

		key := fmt.Sprintf("%d", s.Controller_id)
		states[key] = map[string]interface{}{
			"behaviour":   s.Behaviour,
			"floor":       s.Floor,
			"direction":   s.Direction,
			"cabRequests": cabRequests,
		}
	}

	// 3) Return the final structure
	// This is exactly what your AssignHallRequests expects as input.
	return map[string]interface{}{
		"hallRequests": hallRequests, // shape: [N_FLOORS][2]
		"states":       states,       // dynamic states, keyed by label
	}
}
