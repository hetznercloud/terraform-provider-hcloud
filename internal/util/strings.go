package util

import (
	"strings"
	"unicode"
)

// Dedent is a helper function that dedent a tab indented text. The first line defines
// the number of tabs to remove for the rest of the text.
func Dedent(s string) string {
	s = strings.TrimLeft(s, "\n")
	s = strings.TrimRight(s, "\n\t")

	lines := strings.Split(s, "\n")

	var prefix string
	for i, line := range lines {
		if i == 0 {
			lines[i] = strings.TrimLeft(line, "\t")
			prefix = line[:len(line)-len(lines[i])]
		} else {
			lines[i] = strings.TrimPrefix(line, prefix)
		}
	}

	return strings.Join(lines, "\n")
}

// MarkdownDescription is a helper function that transforms a Go friendly markdown text
// into real markdown.
//
// - Dedent the description
// - Replace 2 single quote with a backtick
func MarkdownDescription(s string) string {
	s = Dedent(s)
	s = strings.ReplaceAll(s, "''", "`")
	return s
}

func TitleCase(s string) string {
	if s == "" {
		return ""
	}
	return string(unicode.ToTitle(rune(s[0]))) + s[1:]
}
