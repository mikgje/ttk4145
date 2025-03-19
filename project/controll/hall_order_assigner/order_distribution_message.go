package order_handler

import (
	"fmt"
	"strconv"
	"main/utilities"
)


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

	fmt.Println(ODM)
	return ODM, nil
}
