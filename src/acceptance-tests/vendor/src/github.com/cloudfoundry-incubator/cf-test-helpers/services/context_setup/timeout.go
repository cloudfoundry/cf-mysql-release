package context_setup

import (
	"time"
)

var TimeoutScale float64

func ScaledTimeout(timeout time.Duration) time.Duration {
	scaledTimeoutSeconds := timeout.Seconds() * TimeoutScale
	return time.Duration(scaledTimeoutSeconds) * time.Second
}
