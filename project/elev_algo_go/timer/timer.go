package timer

import (
	"time"
)

var timer_end_time time.Time
var timer_active int

func Timer_stop() {
	timer_active = 0
}

func Timer_timed_out() int {
	var timer_fin int
	if timer_end_time.Compare(time.Now()) != 1 {
		timer_fin = 1
	}
	if timer_active == timer_fin {
		return 1
	} else {
		return 0
	}
}

func Timer_start(duration float64, trigger_channel chan<- bool, kill_timer_channel chan bool) {
	select {
	case kill_timer_channel <- true:
	default:
	}
	timer := time.NewTimer(time.Duration(duration) * time.Second)
	select {
	case <-timer.C:
		trigger_channel <- true
		return
	case <-kill_timer_channel:
		return
	}
}
func Obstruction_timer(duration int, trigger_channel chan<- bool, abort chan bool) {
	running := true
	obstruction_timer := time.NewTimer(time.Duration(duration) * time.Second)
	for running {
		select {
		case <-obstruction_timer.C:
			trigger_channel <- true
			running = false
			println("Obstruction timer fired")
		case <-abort:
			obstruction_timer.Stop()
			running = false
			println("Obstruction timer aborted")
		}
	}
	println("leaving obstruction timer")
}
