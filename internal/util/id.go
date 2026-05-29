package util

import (
	"fmt"
	"math"
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

func CastInt32(v any) int32 {
	switch o := v.(type) {
	case int64:
		if o > math.MaxInt32 || o < math.MinInt32 {
			panic(fmt.Sprintf("integer value %d out of bounds for int32", o))
		}
		return int32(o)
	case int:
		if o > math.MaxInt32 || o < math.MinInt32 {
			panic(fmt.Sprintf("integer value %d out of bounds for int32", o))
		}
		return int32(o)
	case int32:
		return o
	case float32:
		return CastInt32(int64(o))
	case float64:
		return CastInt32(int64(o))
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
