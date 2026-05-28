package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidIDError(t *testing.T) {
	testCases := []struct {
		name  string
		given error
		want  string
	}{
		{
			name:  "without hint",
			given: NewInvalidIDError("p-1234-127.0.0.1", "$RESOURCE_PREFIX-$RESOURCE_ID-$IP_ADDRESS"),
			want:  "unexpected id 'p-1234-127.0.0.1', expected '$RESOURCE_PREFIX-$RESOURCE_ID-$IP_ADDRESS'",
		},
		{
			name:  "with hint",
			given: NewInvalidIDError("p-1234-127.0.0.1", "$RESOURCE_PREFIX-$RESOURCE_ID-$IP_ADDRESS").WithHint("is $RESOURCE_ID valid?"),
			want:  "unexpected id 'p-1234-127.0.0.1', expected '$RESOURCE_PREFIX-$RESOURCE_ID-$IP_ADDRESS' (is $RESOURCE_ID valid?)",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.given.Error())
		})
	}
}
