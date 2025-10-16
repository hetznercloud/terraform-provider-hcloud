package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTitleCase(t *testing.T) {
	assert.Equal(t, "Hello world", TitleCase("hello world"))
	assert.Equal(t, "A", TitleCase("a"))
	assert.Equal(t, "", TitleCase(""))
}
