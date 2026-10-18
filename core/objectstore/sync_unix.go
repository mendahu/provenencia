//go:build unix

package objectstore

import "os"

// syncDirIn flushes the entries of dir under root to disk.
func syncDirIn(root *os.Root, dir string) error {
	f, err := root.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
