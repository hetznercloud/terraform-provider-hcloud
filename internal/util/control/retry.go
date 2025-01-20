package control

import (
	"errors"
	"log"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// DefaultRetries is a constant for the maximum number of retries we usually do.
// However, callers of Retry are free to choose a different number.
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

	backoff := hcloud.ExponentialBackoff(2, 1*time.Second)

	for try := 0; try < maxTries; try++ {
		var aerr abortErr

		err = f()
		if errors.As(err, &aerr) {
			return aerr.Err
		}
		if err != nil {
			sleep := backoff(try)
			log.Printf("[WARN] try %d/%d failed: retrying after %v: error: %v", try+1, maxTries, sleep, err)
			time.Sleep(sleep)
			continue
		}

		return nil
	}

	return err
}
