//go:build !unix

package objectstore

import "os"

// syncDirIn flushes directory entries to disk where the platform supports it.
func syncDirIn(*os.Root, string) error { return nil }
