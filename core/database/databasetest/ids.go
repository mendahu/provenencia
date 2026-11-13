package databasetest

import "encoding/binary"

// BenchID is a 16-byte id for a benchmark's n-th row of a kind: kind in the
// first byte, n big-endian in the last eight.
func BenchID(kind byte, n int) []byte {
	id := make([]byte, 16)
	id[0] = kind
	binary.BigEndian.PutUint64(id[8:], uint64(n))
	return id
}
