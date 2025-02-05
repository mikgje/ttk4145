package timer

import (
	"time"
)

var timer_end_time time.Time;
var timer_active int;

// Assuming duration is in seconds
func Timer_start(duration float64) {
	timer_end_time = time.Now().Add(time.Duration(duration*1000000000))
	timer_active = 1
}

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


func Timer_start2(duration float64, channel chan<- bool){
	timer := time.NewTimer(time.Duration(duration) * time.Second)
	<- timer.C
	channel <- true

	println("Timer fired")

}