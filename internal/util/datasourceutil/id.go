package datasourceutil

import (
	"crypto/sha1" // nolint: gosec
	"fmt"
)

func ListID[T any](ids []T) string {
	var b []byte
	for _, id := range ids {
		b = fmt.Append(b, id)
	}
	return fmt.Sprintf("%x", sha1.Sum(b)) // nolint: gosec
}
