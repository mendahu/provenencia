//go:build !unix

package objectstore

// SyncDir flushes directory entries to disk where the platform supports it.
func SyncDir(string) error { return nil }
