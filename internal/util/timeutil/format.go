package timeutil

import "time"

// TimeStringLayout is the layout used by time/Time.String() to create a string
// representation of a Time value.
const TimeStringLayout = "2006-01-02 15:04:05.999999999 -0700 MST"

// ConvertFormat converts the time format of cur from curLayout to newLayout.
//
// This is achieved by first attempting to parse cur using old layout. If
// parsing fails the error is returned verbatim. Usually this indicates that
// cur did not match curLayout and the caller may act accordingly.
//
// If parsing cur was successful it is converted to newLayout using
// time/Time.Format().
func ConvertFormat(cur, curLayout, newLayout string) (string, error) {
	t, err := time.Parse(curLayout, cur)
	if err != nil {
		return cur, err
	}
	return t.Format(newLayout), nil
}
