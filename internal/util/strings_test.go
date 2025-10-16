package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDedent(t *testing.T) {
	testCases := []struct {
		desc  string
		given string
		want  string
	}{
		{
			desc:  "with indent",
			given: "\thello\n\t\tworld",
			want:  "hello\n\tworld",
		},
		{
			desc:  "with negative indent",
			given: "\thello\nworld",
			want:  "hello\nworld",
		},
		{
			desc:  "with extra new lines",
			given: "\n\thello\n\t\tworld\n",
			want:  "hello\n\tworld",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			require.Equal(t, testCase.want, Dedent(testCase.given))
		})
	}
}

func TestMarkdownDescription(t *testing.T) {
	testCases := []struct {
		desc  string
		given string
		want  string
	}{
		{
			desc:  "transform backticks",
			given: "hello ''world''",
			want:  "hello `world`",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			require.Equal(t, testCase.want, MarkdownDescription(testCase.given))
		})
	}
}

func TestTitleCase(t *testing.T) {
	assert.Equal(t, "Hello world", TitleCase("hello world"))
	assert.Equal(t, "A", TitleCase("a"))
	assert.Equal(t, "", TitleCase(""))
}
