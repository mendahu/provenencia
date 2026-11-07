package recordlinks

import "math"

// Lines is the one-pair count, for the tests that compare two graphs.
var Lines = pairLines

// PairWork is the work of a pair of graphs, for the cap's measurement.
func PairWork(page []byte, a, b Graph, cfg Config) int {
	return pairWork(sharedOf(page, a, b), a, b, holdings{}, holdings{}, cfg, math.MaxInt)
}

// ReadWork is the work of a read of graphs as Pairs counts it, uncapped.
func ReadWork(page []byte, graphs []Graph, cfg Config) int {
	return readWork(page, graphs, holdings{}, cfg, math.MaxInt)
}
