package util

import (
	"fmt"
	"strconv"
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

func FormatID[T ~int | ~int32 | ~int64](v T) string {
	return strconv.FormatInt(CastInt64(v), 10)
}

func ParseID(v string) (int64, error) {
	return strconv.ParseInt(v, 10, 32)
}
