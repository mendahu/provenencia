// Package objectstore verifies, writes, and removes content-addressed object files on disk.
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

// WriteIfAbsent writes data at objPath unless a file there already matches
// checksum (lowercase hex SHA-256 of data). It writes a temp file, syncs it, and
// renames it into place, so a reader never sees a partial object. wrote is true
// when this call left new bytes at objPath. A checksum that does not match data
// is an error and nothing is written. A directory-sync error is ignored.
func WriteIfAbsent(objPath, checksum string, data []byte) (wrote bool, err error) {
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != checksum {
		return false, fmt.Errorf("objectstore: data does not match checksum %s", checksum)
	}
	if match, err := MatchesChecksum(objPath, checksum); err != nil {
		return false, err
	} else if match {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(objPath), 0o755); err != nil {
		return false, err
	}
	tmp, err := os.CreateTemp(filepath.Dir(objPath), ".write-*")
	if err != nil {
		return false, err
	}
	tmpName := tmp.Name()
	done := false
	defer func() {
		if !done {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return false, err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return false, err
	}
	if err := tmp.Close(); err != nil {
		return false, err
	}
	if err := os.Rename(tmpName, objPath); err != nil {
		if match, statErr := MatchesChecksum(objPath, checksum); statErr == nil && match {
			return false, nil
		}
		return false, err
	}
	done = true
	_ = SyncDir(filepath.Dir(objPath))
	return true, nil
}
