package control

import (
	"errors"
	"log"
	"time"
)

// DefaultRetries is a constant for the maximum number of retries we usually do.
// However callers of Retry are free to choose a different number.
const DefaultRetries = 5

type abortErr struct {
	Err error
}

func (e abortErr) Error() string {
	return e.Err.Error()
}

func (e abortErr) Unwrap() error {
	return e.Err
}

// AbortRetry aborts any further attempts of retrying an operation.
//
// If err is passed Retry returns the passed error. If nil is passed, Retry
// returns nil.
func AbortRetry(err error) error {
	if err == nil {
		return nil
	}
	return abortErr{Err: err}
}

// Retry executes f at most maxTries times.
func Retry(maxTries int, f func() error) error {
	var err error

	for try := 0; try < maxTries; try++ {
		var aerr abortErr

		err = f()
		if errors.As(err, &aerr) {
			return aerr.Err
		}
		if err != nil {
			d := time.Duration(try+1) * time.Second
			log.Printf("[WARN] try %d/%d failed: retrying after %v: error: %v", try+1, maxTries, d, err)
			time.Sleep(time.Duration(try) * time.Second)
			continue
		}

		return nil
	}

	return err
}
