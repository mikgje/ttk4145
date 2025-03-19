package order_handler

import (
	"fmt"
	"main/utilities"
)

func Order_handler(statuses []utilities.StatusMessage) utilities.OrderDistributionMessage {
	assigner_input := build_assigner_input_from_status_messages(statuses)
	assigner_outut, err := assign_hall_requests(assigner_input)
	if err != nil {
		fmt.Println(err)
		return utilities.OrderDistributionMessage{}
	}
	ODM, err := order_distribution_message(assigner_outut)
	if err != nil {
		fmt.Println(err)
		return utilities.OrderDistributionMessage{}
	}
	return ODM
}
