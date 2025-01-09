package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDurationFromSeconds(t *testing.T) {
	assert.Equal(t, time.Second, DurationFromSeconds(1))
	assert.Equal(t, 2*time.Second, DurationFromSeconds(int32(2)))
	assert.Equal(t, 2*time.Second, DurationFromSeconds(int64(2)))
}
