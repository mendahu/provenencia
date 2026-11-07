package recordlinks

import "math"

// PairWork is the work of a pair of graphs, for the cap's measurement.
func PairWork(page []byte, a, b Graph, cfg Config) int {
	return pairWork(string(page), a, b, holdings{}, holdings{}, cfg, math.MaxInt)
}
