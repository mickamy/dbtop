package durations

import "time"

func FromSeconds(s float64) time.Duration {
	return time.Duration(s * float64(time.Second))
}
