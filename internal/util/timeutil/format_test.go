package timeutil_test

import (
	"testing"
	"time"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/timeutil"
	"github.com/stretchr/testify/assert"
)

func TestConvertFormat(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		curVal    string
		curLayout string
		newVal    string
		newLayout string
		expectErr bool
	}{
		{
			name:      "layout is the same",
			curVal:    now.Format(time.RFC3339),
			curLayout: time.RFC3339,
			newVal:    now.Format(time.RFC3339),
			newLayout: time.RFC3339,
		},
		{
			name:      "curLayout does not match curVal",
			curVal:    now.Format(time.RFC3339),
			curLayout: time.RFC1123Z,
			newLayout: time.RFC822,
			expectErr: true,
		},
		{
			name:      "successful conversion from cur to new",
			curVal:    now.Format(time.RFC3339),
			curLayout: time.RFC3339,
			newVal:    now.Format(time.RFC822),
			newLayout: time.RFC822,
		},
		{
			name:      "convert from TimeStringLayout",
			curVal:    now.UTC().String(),
			curLayout: timeutil.TimeStringLayout,
			newVal:    now.UTC().Format(time.RFC3339),
			newLayout: time.RFC3339,
		},
		{
			name:      "convert from TimeStringLayout",
			curVal:    "2021-05-27 14:39:46.877103 +0000 +0000",
			curLayout: timeutil.TimeStringLayout,
			newVal:    "2021-05-27T14:39:46Z",
			newLayout: time.RFC3339,
		},
		{
			name:      "convert from TimeStringLayout",
			curVal:    "2021-05-27 14:39:46.877103 +0000 UTC",
			curLayout: timeutil.TimeStringLayout,
			newVal:    "2021-05-27T14:39:46Z",
			newLayout: time.RFC3339,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			v, err := timeutil.ConvertFormat(tt.curVal, tt.curLayout, tt.newLayout)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.newVal, v)
		})
	}
}
