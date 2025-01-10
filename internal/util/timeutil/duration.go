package timeutil

import (
	"time"
)

func DurationFromSeconds[T int | int32 | int64](value T) time.Duration {
	return time.Duration(float64(value) * float64(time.Second))
}
