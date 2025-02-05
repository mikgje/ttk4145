package timer

import (
	"time"
)

var (
	timer_end_time time.Time
	timer_active   int
	timer_channel  chan bool
)

// Assuming duration is in seconds
func Timer_start(duration float64) {
	timer_end_time = time.Now().Add(time.Duration(duration * float64(time.Second)))
	timer_active = 1
	timer_channel = make(chan bool, 1)
	go func() {
		time.Sleep(time.Duration(duration * float64(time.Second)))
		timer_channel <- true
	}()
}

func Timer_stop() {
	timer_active = 0
	close(timer_channel)
}

func Timer_timed_out() int {
	select {
	case <-timer_channel:
		timer_active = 0
		return 1
	default:
		return 0
	}
}
