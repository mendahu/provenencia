package databasetest

import "runtime"

// HeapDelta returns the bytes fill adds to the live heap, read after a GC
// before and after, so the number is that fill's own cost.
func HeapDelta(fill func()) float64 {
	before := LiveHeap()
	fill()
	return LiveHeap() - before
}

// LiveHeap is HeapAlloc in bytes after a GC.
func LiveHeap() float64 {
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return float64(ms.HeapAlloc)
}
