package merge

import (
	"maps"
)

func Maps[M ~map[K]V, K comparable, V any](items ...M) M {
	acc := make(M)
	for _, item := range items {
		maps.Copy(acc, item)
	}
	return acc
}
