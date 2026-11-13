package graphcache

import "slices"

// SnapshotRegistryForTest keeps the registry as it is and returns a func that
// puts it back, so a test registers entries without leaving them for others.
func SnapshotRegistryForTest() (restore func()) {
	saved := slices.Clone(registry)
	return func() { registry = saved }
}

// InputIndexSizeForTest counts the (index key, held key) pairs of a keyed
// entry's input index.
func InputIndexSizeForTest[T any](g *Graph, k Keyed[T]) int {
	kv := k.data(g)
	if kv == nil {
		return 0
	}
	n := 0
	for _, keys := range kv.byInput {
		n += len(keys)
	}
	return n
}
