package util

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func CastInt64(v any) int64 {
	switch o := v.(type) {
	case int64:
		return o
	case int:
		return int64(o)
	default:
		panic(fmt.Sprintf("unexpected type %T for %#v", v, v))
	}
}

func CastInt(v any) int {
	switch o := v.(type) {
	case int:
		return o
	case int64:
		if o > math.MaxInt32 && strconv.IntSize == 32 {
			panic("cannot cast int64 value to int, value is larger than max int on this system")
		}
		return int(o)
	default:
		panic(fmt.Sprintf("unexpected type %T for %#v", v, v))
	}
}

func FormatID[T ~int | ~int64](v T) string {
	return strconv.FormatInt(CastInt64(v), 10)
}

func ParseID(v string) (int64, error) {
	return strconv.ParseInt(v, 10, 64)
}

type InvalidIDError struct {
	value    string
	expected string
	hint     string
}

func NewInvalidIDError(value string, expected string) *InvalidIDError {
	return &InvalidIDError{value: value, expected: expected, hint: ""}
}

func (e *InvalidIDError) WithHint(hint string) *InvalidIDError {
	e.hint = hint
	return e
}

func (e *InvalidIDError) Error() string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "unexpected id '%s', expected '%s'", e.value, e.expected)
	if e.hint != "" {
		fmt.Fprintf(b, " (%s)", e.hint)
	}
	return b.String()
}
