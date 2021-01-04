package hcloud

import (
	"fmt"
	"log"
	"time"
)

const defaultMaxRetries = 5

func retry(maxRetries int, f func() error) error {
	try := 1
	for {
		if err := f(); err != nil {
			if try <= maxRetries {
				log.Printf("[WARN] function returned error in try %d, retry after %d: %v", try, time.Duration(try)*time.Second, err)
				time.Sleep(time.Duration(try) * time.Second)
				try++
			} else {
				return fmt.Errorf("func returned an error after %d try: %w", try, err)
			}
		} else {
			break
		}
	}
	return nil
}
