package util

import (
	"strings"
	"unicode"
)

func TitleCase(s string) string {
	if s == "" {
		return ""
	}
	return string(unicode.ToTitle(rune(s[0]))) + s[1:]
}

func Dedent(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimLeft(line, "\t")
	}
	return strings.Join(lines, "\n")
}

func MarkdownDescription(s string) string {
	s = Dedent(s)
	s = strings.ReplaceAll(s, "''", "`")
	return s
}
