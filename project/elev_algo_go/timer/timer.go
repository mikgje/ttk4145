package timer

// Timer functions for use in the single elevator system

import (
	"time"
)

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
	obstruction_timer := time.NewTimer(time.Duration(duration) * time.Second)
	select {
	case <-obstruction_timer.C:
		trigger_channel <- true
		println("Obstruction timer triggered")
		return
	case <-abort:
		obstruction_timer.Stop()
		println("Obstruction timer aborted")
		return
	}
}
