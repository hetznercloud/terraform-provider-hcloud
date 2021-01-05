package hcloud

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	type testCase struct {
		name          string
		f             func(tt *testCase) error
		maxTries      int
		expectedTries int
		expectedErr   error

		actualTries int // modified during test
	}

	tests := []testCase{
		{
			name:          "No retries if successful",
			f:             func(_ *testCase) error { return nil },
			maxTries:      5,
			expectedTries: 1,
		},
		{
			name: "Retries if unsuccessful",
			f: func(tt *testCase) error {
				if tt.actualTries < tt.expectedTries {
					return errors.New("retry me")
				}
				return nil
			},
			maxTries:      5,
			expectedTries: 3,
		},
		{
			name: "Retries no more than max tries times",
			f: func(tt *testCase) error {
				return tt.expectedErr
			},
			maxTries:      2,
			expectedTries: 2,
			expectedErr:   errors.New("expected"),
		},
		{
			name: "Abort retries",
			f: func(tt *testCase) error {
				return abortRetry(tt.expectedErr)
			},
			maxTries:      5,
			expectedTries: 1,
			expectedErr:   errors.New("pointless"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := retry(tt.maxTries, func() error {
				tt.actualTries++
				t.Logf("Try %d/%d", tt.actualTries, tt.expectedTries)
				return tt.f(&tt)
			})
			assert.Equal(t, tt.expectedTries, tt.actualTries)
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v; got %v", tt.expectedErr, err)
			}
		})
	}
}
