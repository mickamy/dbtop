package durations

import "time"

func FromSeconds(s float64) time.Duration {
	return time.Duration(s * float64(time.Second))
}

func FromMillis(ms float64) time.Duration {
	return time.Duration(ms * float64(time.Millisecond))
}
