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

// syncDirIn flushes the entries of dir under root to disk.
func syncDirIn(root *os.Root, dir string) error {
	f, err := root.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
