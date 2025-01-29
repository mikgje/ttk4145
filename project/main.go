package main

import (
	"main/elev_algo_go/elevator"
	"main/elev_algo_go/fsm"
	"main/elev_algo_go/timer"
	"main/elevio"
	"time"
	// "fmt"
)

func main() {

	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	// var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	if elevio.GetFloor() == -1 {
		fsm.Fsm_on_init_between_floors()
	}

	var prev [elevator.N_FLOORS][elevator.N_BUTTONS]bool
	prev_const := -1

	for {
		{
			for f := 0; f < elevator.N_FLOORS; f++ {
				for i := 0; i < elevator.N_BUTTONS; i++ {
					b := elevio.ButtonType(i)
					v := elevio.GetButton(b, f)
					if v && (v != prev[f][i]) {
						fsm.Fsm_on_request_button_press(f, b)
					}
					prev[f][b] = v
				}
			}
		}

		{
			f := elevio.GetFloor()
			if (f != -1) && (f != prev_const) {
				fsm.Fsm_on_floor_arrival(f)
			}
			prev_const = f
		}

		{
			if timer.Timer_timed_out() == 1 {
				timer.Timer_stop()
				fsm.Fsm_on_door_timeout()
			}
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
