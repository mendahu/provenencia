//go:build unix

package objectstore

import "os"

// SyncDir flushes directory entries to disk where the platform supports it.
func SyncDir(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
