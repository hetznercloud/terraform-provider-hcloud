package hcloud

import (
	"errors"
	"log"
	"time"
)

const defaultMaxRetries = 5

type abortErr struct {
	Err error
}

func (e abortErr) Error() string {
	return e.Err.Error()
}

func (e abortErr) Unwrap() error {
	return e.Err
}

func abortRetry(err error) error {
	return abortErr{Err: err}
}

func retry(maxTries int, f func() error) error {
	var err error

	for try := 0; try < maxTries; try++ {
		var aerr abortErr

		err = f()
		if errors.As(err, &aerr) {
			return aerr.Err
		}
		if err != nil {
			d := time.Duration(try) * time.Second
			log.Printf("[WARN] try %d/%d failed: retrying after %v: error: %v", try, maxTries, d, err)
			time.Sleep(time.Duration(try) * time.Second)
			continue
		}

		return nil
	}

	return err
}
