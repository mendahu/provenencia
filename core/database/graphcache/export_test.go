package graphcache

import "slices"

// SnapshotRegistryForTest keeps the registry as it is and returns a func that
// puts it back, so a test registers entries without leaving them for others.
func SnapshotRegistryForTest() (restore func()) {
	saved := slices.Clone(registry)
	return func() { registry = saved }
}
