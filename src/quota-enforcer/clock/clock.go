package clock

import "time"

type Clock interface {
	After(time.Duration) <-chan time.Time
}

type clock struct{}

func DefaultClock() Clock {
	return &clock{}
}

func (this clock) After(interval time.Duration) <-chan time.Time {
	return time.After(interval)
}
