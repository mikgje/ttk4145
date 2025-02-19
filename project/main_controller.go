package main
// Interface between controller network and elevator
import (
	"fmt"
)

func main_controller() {

	for {
		select {
		case msg := <-elev_to_ctrl_chan:
			fmt.Println("Controller received message from elevator: ", msg)
		}
	
	}

}