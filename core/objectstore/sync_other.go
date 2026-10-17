//go:build !unix

package objectstore

import "os"

// SyncDir flushes directory entries to disk where the platform supports it.
func SyncDir(string) error { return nil }

// syncDirIn flushes directory entries to disk where the platform supports it.
func syncDirIn(*os.Root, string) error { return nil }
