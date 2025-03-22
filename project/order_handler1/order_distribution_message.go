package order_assigner

import (
	"fmt"
	"strconv"

	"main/elev_algo_go/elevator"
	"main/utilities"
)

// OrderDistributionMessage konverterer raw output (map[string]interface{})
// til en ferdig OrderDistributionMessage.
// Her forventes at nøklene "0", "1" og "2" korresponderer med rekkefølgen (topp til bunn)
// i Orderlines.


func OrderDistributionMessage(rawOutput map[string]interface{}) (utilities.OrderDistributionMessage, error) {
	var ODM utilities.OrderDistributionMessage // oppretter denne som allerede inneholder false
	ODM.Label = "D" // Sett ønsket label

	// Iterer over forventede nøkler "0", "1", "2"
	for i := 0; i < len(ODM.Orderlines); i++ {
		key := strconv.Itoa(i)
		orders, ok := rawOutput[key].([]interface{})
		if !ok {
			// Manglende nøkkel: vi lar standardverdien (false) ligge igjen.
			continue
		}

		// Forvent at hver nøkkel inneholder en slice med N_FLOORS elementer
		for j := 0; j < elevator.N_FLOORS && j < len(orders); j++ {
			// Hver etasje bør være en slice med N_BUTTONS-1 booleans
			floorOrders, ok := orders[j].([]interface{})
			if !ok {
				return ODM, fmt.Errorf("uventet type for nøkkel %s, etasje %d", key, j)
			}

			for k := 0; k < elevator.N_BUTTONS-1 && k < len(floorOrders); k++ {
				state, ok := floorOrders[k].(bool)
				if !ok {
					return ODM, fmt.Errorf("uventet type for nøkkel %s, etasje %d, knapp %d", key, j, k)
				}
				ODM.Orderlines[i][j][k] = state
			}
		}
	}

	fmt.Println(ODM)

	return ODM, nil
}
