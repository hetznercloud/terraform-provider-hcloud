package merge

import "sort"

// StringSlice merges the elements in src into dst while retaining the order
// of already existing elements in dst.
func StringSlice(dst, src []string) []string {
	if len(dst) == 0 && len(src) == 0 {
		return nil
	}

	// Find all elements in src which are not in dst.
	candidates := make(map[string]int, len(src))
	for i, s := range src {
		candidates[s] = i
	}
	for _, s := range dst {
		delete(candidates, s)
	}

	// Order the remaining candidates by their initial index.
	newElems := make([]string, 0, len(candidates))
	for s := range candidates {
		newElems = append(newElems, s)
	}
	sort.Slice(newElems, func(i, j int) bool {
		si, sj := newElems[i], newElems[j]
		return candidates[si] < candidates[sj]
	})

	return append(dst, newElems...)
}
