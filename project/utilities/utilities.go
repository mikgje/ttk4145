package utilities

type StatusMessage struct {
	Controller_id int
	Behaviour     string
	Floor         int
	Direction     string
	Node_orders   [N_FLOORS][N_BUTTONS]bool
}

type OrderDistributionMessage struct {
	Orderlines [3][N_FLOORS][N_BUTTONS - 1]bool
}

type State int

const (
	State_slave State = iota
	State_master
	State_disconnected
)

/*========================================PROJECT CONSTANTS========================================*/

const (
	N_FLOORS  	int = 4
	N_BUTTONS 	int = 3
	N_ELEVS 	int = 3
)
