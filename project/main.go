package main

import (
	// "main/elev_algo_go/elevator"
	"main/elev_algo_go/fsm"
	"main/elev_algo_go/timer"
	"main/elevio"
	"time"
	// "fmt"
)

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)
	fsm.Fsm_init()

	// var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	if elevio.GetFloor() == -1 {
		fsm.Fsm_on_init_between_floors()
	}

	floor_channel := make(chan int)
	button_channel := make(chan elevio.ButtonEvent)
	obstruction_channel := make(chan bool)

	for {

		go elevio.PollButtons(button_channel)
		go elevio.PollFloorSensor(floor_channel)
		go elevio.PollObstructionSwitch(obstruction_channel)

		
		select {
		case button := <- button_channel:
			fsm.Fsm_on_request_button_press(button.Floor, button.Button)
		case floor := <- floor_channel:
			fsm.Fsm_on_floor_arrival(floor)
		// case obstruction := <- obstruction_channel:
		}

		if timer.Timer_timed_out() == 1{
			timer.Timer_stop()
			fsm.Fsm_on_door_timeout()
		}

		time.Sleep(time.Second)
	}

	//     drv_buttons := make(chan elevio.ButtonEvent)
	//     drv_floors  := make(chan int)
	//     drv_obstr   := make(chan bool)
	//     drv_stop    := make(chan bool)

	//     go elevio.PollButtons(drv_buttons)
	//     go elevio.PollFloorSensor(drv_floors)
	//     go elevio.PollObstructionSwitch(drv_obstr)
	//     go elevio.PollStopButton(drv_stop)

	//     for {
	//         select {
	//         case a := <- drv_buttons:
	//             fmt.Printf("%+v\n", a)
	//             elevio.SetButtonLamp(a.Button, a.Floor, true)

	//         case a := <- drv_floors:
	//             fmt.Printf("%+v\n", a)
	//             if a == numFloors-1 {
	//                 d = elevio.MD_Down
	//             } else if a == 0 {
	//                 d = elevio.MD_Up
	//             }
	//             elevio.SetMotorDirection(d)

	//         case a := <- drv_obstr:
	//             fmt.Printf("%+v\n", a)
	//             if a {
	//                 elevio.SetMotorDirection(elevio.MD_Stop)
	//             } else {
	//                 elevio.SetMotorDirection(d)
	//             }

	//	    case a := <- drv_stop:
	//	        fmt.Printf("%+v\n", a)
	//	        for f := 0; f < numFloors; f++ {
	//	            for b := elevio.ButtonType(0); b < 3; b++ {
	//	                elevio.SetButtonLamp(b, f, false)
	//	            }
	//	        }
	//	    }
	//	}
}
