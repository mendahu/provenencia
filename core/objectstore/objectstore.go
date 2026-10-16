// Package objectstore verifies and removes content-addressed object files on disk.
package objectstore

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// MatchesChecksum reports whether objPath is a regular file whose SHA-256
// equals wantChecksum (lowercase hex). Missing paths return (false, nil).
func MatchesChecksum(objPath, wantChecksum string) (bool, error) {
	f, err := os.Open(objPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return false, err
	}
	if !st.Mode().IsRegular() {
		return false, nil
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}
	return hex.EncodeToString(h.Sum(nil)) == wantChecksum, nil
}

// RemoveObject removes one object file under projectDir/objects without
// following a symlink out of that folder. rel is the slash-separated path from
// the project folder and must start with objects/. A missing object returns nil.
func RemoveObject(projectDir, rel string) error {
	inner, ok := strings.CutPrefix(rel, "objects/")
	if !ok || inner == "" {
		return fmt.Errorf("objectstore: %q is not under objects/", rel)
	}
	dir := filepath.Join(projectDir, "objects")
	st, err := os.Lstat(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	if st.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("objectstore: %s is a symlink", dir)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()
	if err := root.Remove(filepath.FromSlash(inner)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}
