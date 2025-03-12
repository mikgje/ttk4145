package utilities

import (

)

type StatusMessage struct {
	Label string
	Controller_id int
	Behaviour     string
	Floor         int
	Direction     string
	Node_orders   [N_FLOORS][N_BUTTONS]bool
}

type OrderDistributionMessage struct {
	Label string
	Orderlines [3][N_FLOORS][N_BUTTONS - 1]bool
}


/*========================================PROJECT CONSTANTS========================================*/

const (
	N_FLOORS  int = 4
	N_BUTTONS int = 3
	UNHEALTHY_FLAG int = -1
	HEALTHY_FLAG int = 100
)