package helpers

import (
	"cmp"
	"sort"
)

// Map is a map with helper methods for ordered key types.
type Map[K cmp.Ordered, V any] map[K]V

// Keys returns the keys of m in unspecified order.
func (m Map[K, V]) Keys() []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

// SortedKeys returns the keys of m in ascending order.
func (m Map[K, V]) SortedKeys() []K {
	keys := m.Keys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}
