package datasourceutil

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListID(t *testing.T) {
	expected := "40bd001563085fc35165329ea1ff5c5ecbdbbeef"

	assert.Equal(t, expected, ListID([]string{"1", "2", "3"}))
	assert.Equal(t, expected, ListID([]int{1, 2, 3}))
	// Regression test
	assert.Equal(t, expected, fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join([]string{"1", "2", "3"}, ""))))) // nolint: gosec
}
