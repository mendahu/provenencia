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

	"github.com/mendahu/provenencia/core/objectpath"
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

// ErrNotRegular means an object path is a symlink, FIFO, folder, or device
// instead of a regular file.
var ErrNotRegular = errors.New("objectstore: object is not a regular file")

// ErrUnsafeObjects means the project's objects folder is a symlink or not a folder.
var ErrUnsafeObjects = errors.New("objectstore: objects folder is a symlink or not a folder")

// innerPath cuts the objects/ prefix off rel and returns the rest in the
// platform's separator.
func innerPath(rel string) (string, error) {
	inner, ok := strings.CutPrefix(rel, objectpath.Dir+"/")
	if !ok || inner == "" {
		return "", fmt.Errorf("objectstore: %q is not under %s/", rel, objectpath.Dir)
	}
	return filepath.FromSlash(inner), nil
}

// openObjectsRoot opens projectDir/objects as an os.Root. With create it makes
// the folder first. A missing folder is fs.ErrNotExist; a symlink or a
// non-folder is ErrUnsafeObjects. The check and the open both happen inside
// one opened project root, so a swap of objects/ between them cannot escape.
func openObjectsRoot(projectDir string, create bool) (*os.Root, error) {
	proj, err := os.OpenRoot(projectDir)
	if err != nil {
		return nil, err
	}
	defer proj.Close()
	if create {
		if err := proj.Mkdir(objectpath.Dir, 0o755); err != nil && !errors.Is(err, fs.ErrExist) {
			return nil, err
		}
	}
	st, err := proj.Lstat(objectpath.Dir)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() {
		return nil, ErrUnsafeObjects
	}
	return proj.OpenRoot(objectpath.Dir)
}

// openRegular opens inner under root for reading. Anything but a regular file
// is ErrNotRegular, checked before the open so a FIFO cannot block it.
func openRegular(root *os.Root, inner string) (*os.File, error) {
	st, err := root.Lstat(inner)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, ErrNotRegular
	}
	f, err := root.OpenFile(inner, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	st, err = f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if !st.Mode().IsRegular() {
		_ = f.Close()
		return nil, ErrNotRegular
	}
	return f, nil
}

// OpenObject opens one object file for reading without following a symlink
// out of projectDir/objects. rel is the slash-separated path from the project
// folder and must start with objects/. A missing object is fs.ErrNotExist, a
// symlinked objects folder is ErrUnsafeObjects, and an object that is not a
// regular file is ErrNotRegular. The caller closes the file.
func OpenObject(projectDir, rel string) (*os.File, error) {
	inner, err := innerPath(rel)
	if err != nil {
		return nil, err
	}
	root, err := openObjectsRoot(projectDir, false)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return openRegular(root, inner)
}

// matches reports whether inner under root is a regular file whose SHA-256
// equals checksum (lowercase hex). A missing or non-regular path is false.
func matches(root *os.Root, inner, checksum string) (bool, error) {
	f, err := openRegular(root, inner)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, ErrNotRegular) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}
	return hex.EncodeToString(h.Sum(nil)) == checksum, nil
}

// RemoveObject removes one object file under projectDir/objects without
// following a symlink out of that folder. rel is the slash-separated path from
// the project folder and must start with objects/. A missing object returns nil.
func RemoveObject(projectDir, rel string) error {
	inner, err := innerPath(rel)
	if err != nil {
		return err
	}
	root, err := openObjectsRoot(projectDir, false)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	defer root.Close()
	if err := root.Remove(inner); err != nil && !errors.Is(err, fs.ErrNotExist) {
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
