package database

import "strings"

// UniqueBlobIDs returns distinct 16-byte identifiers in first-seen order.
func UniqueBlobIDs(ids [][]byte) [][]byte {
	seen := make(map[string]struct{}, len(ids))
	out := make([][]byte, 0, len(ids))
	for _, id := range ids {
		if len(id) != 16 {
			continue
		}
		k := string(id)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, append([]byte(nil), id...))
	}
	return out
}

// SQLInPlaceholders is n SQLite bind markers joined by commas.
func SQLInPlaceholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("?,", n-1) + "?"
}

// BlobArgs copies 16-byte ids into a Query/Exec args slice.
func BlobArgs(ids [][]byte) []any {
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return args
}

// InBatch bounds one IN list. SQLite binds at most 32,766 values per
// statement; a batch stays far under it, with room for a query's other
// arguments, so a read of any size is a fixed number of queries per batch.
const InBatch = 500

// ForEachBatch calls fn with consecutive slices of ids, each at most
// InBatch long. It stops at fn's first error.
func ForEachBatch(ids [][]byte, fn func(batch [][]byte) error) error {
	for start := 0; start < len(ids); start += InBatch {
		if err := fn(ids[start:min(start+InBatch, len(ids))]); err != nil {
			return err
		}
	}
	return nil
}
