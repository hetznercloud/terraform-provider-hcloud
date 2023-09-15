package merge_test

import (
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/merge"
	"github.com/stretchr/testify/assert"
)

func TestStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		src      []string
		dst      []string
		expected []string
	}{
		{
			name: "nil slices",
		},
		{
			name: "empty slices",
			src:  []string{},
			dst:  []string{},
		},
		{
			name:     "dst has all elements of src",
			src:      []string{"a", "b"},
			dst:      []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "src has additional elements",
			src:      []string{"a", "a1", "b", "b1"},
			dst:      []string{"a", "b"},
			expected: []string{"a", "b", "a1", "b1"},
		},
		{
			name:     "dst has additional elements",
			src:      []string{"a", "b"},
			dst:      []string{"a", "a1", "b", "b1"},
			expected: []string{"a", "a1", "b", "b1"},
		},
		{
			name:     "src and dst are disjunct",
			src:      []string{"c", "d"},
			dst:      []string{"a", "b"},
			expected: []string{"a", "b", "c", "d"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual := merge.StringSlice(tt.dst, tt.src)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
