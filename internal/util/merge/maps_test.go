package merge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaps(t *testing.T) {
	assert.Equal(t,
		map[string]string{
			"a": "4",
			"b": "2",
			"c": "3",
		},
		Maps(
			map[string]string{"a": "1"},
			map[string]string{"b": "2"},
			map[string]string{"c": "3", "a": "4"},
		),
	)
}
