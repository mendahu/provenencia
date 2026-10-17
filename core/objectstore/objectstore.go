// Package objectstore verifies, writes, and removes content-addressed object files on disk.
package objectstore

import (
	"crypto/rand"
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

// openRegular opens inner under root for reading. Anything but a regular file,
// or a shard folder that is not a real folder, is ErrNotRegular. The checks
// come before the open so a FIFO cannot block it.
func openRegular(root *os.Root, inner string) (*os.File, error) {
	// Each shard folder must be a real folder: a symlink there is refused like
	// a symlinked object. The folders are checked top-down, so a missing one
	// is fs.ErrNotExist.
	var dirs []string
	for dir := filepath.Dir(inner); dir != "."; dir = filepath.Dir(dir) {
		dirs = append(dirs, dir)
	}
	for i := len(dirs) - 1; i >= 0; i-- {
		dst, err := root.Lstat(dirs[i])
		if err != nil {
			return nil, err
		}
		if !dst.IsDir() {
			return nil, ErrNotRegular
		}
	}
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
// It then removes the shard folders the removal left empty, up to but not
// including objects/. A folder whose only entry is a Finder `.DS_Store` counts
// as empty.
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
	// root.Remove refuses a folder that still holds anything, which ends the walk.
	for dir := filepath.Dir(inner); dir != "."; dir = filepath.Dir(dir) {
		if root.Remove(dir) == nil {
			continue
		}
		if !onlyFinderMetadata(root, dir) || root.Remove(filepath.Join(dir, finderMetadata)) != nil || root.Remove(dir) != nil {
			break
		}
	}
	return nil
}

// finderMetadata is the file Finder leaves in a folder it has shown.
const finderMetadata = ".DS_Store"

// onlyFinderMetadata reports whether dir under root holds nothing but a
// regular file named .DS_Store.
func onlyFinderMetadata(root *os.Root, dir string) bool {
	d, err := root.Open(dir)
	if err != nil {
		return false
	}
	defer d.Close()
	entries, err := d.ReadDir(2)
	return err == nil && len(entries) == 1 && entries[0].Name() == finderMetadata && entries[0].Type().IsRegular()
}

// createTempIn creates an exclusive temp file named prefix plus 16 random hex
// characters at the top of root. It returns the file and its root-relative name.
func createTempIn(root *os.Root, prefix string) (*os.File, string, error) {
	var lastErr error
	for range 10 {
		var b [8]byte
		if _, err := rand.Read(b[:]); err != nil {
			return nil, "", err
		}
		name := prefix + hex.EncodeToString(b[:])
		f, err := root.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
		if err == nil {
			return f, name, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return nil, "", err
		}
		lastErr = err
	}
	return nil, "", lastErr
}

// installTemp moves the finished temp file tmp to inner under root unless inner
// already matches checksum. It always removes tmp on return. A symlink or other
// non-matching entry at inner is replaced, never followed. wrote is true when
// this call left new bytes at inner; a directory-sync error comes back with it.
func installTemp(root *os.Root, tmp, inner, checksum string) (wrote bool, err error) {
	defer func() { _ = root.Remove(tmp) }()

	if match, err := matches(root, inner, checksum); err != nil {
		return false, err
	} else if match {
		return false, nil
	}
	if err := root.MkdirAll(filepath.Dir(inner), 0o755); err != nil {
		return false, err
	}
	if err := root.Remove(inner); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	if err := root.Rename(tmp, inner); err != nil {
		if match, matchErr := matches(root, inner, checksum); matchErr == nil && match {
			return false, nil
		}
		return false, err
	}
	if err := syncDirIn(root, filepath.Dir(inner)); err != nil {
		return true, err
	}
	return true, nil
}

// Temp is a streamed object being written inside projectDir/objects. Install
// moves it into place; Discard removes it. Callers never see its path.
type Temp struct {
	root         *os.Root
	f            *os.File
	name         string
	closed, done bool
}

// CreateTemp creates a temp file at the top of projectDir/objects (making the
// folder when absent). A symlinked objects folder is ErrUnsafeObjects. Finish
// with Install or Discard; either releases the handle on the folder.
func CreateTemp(projectDir, prefix string) (*Temp, error) {
	root, err := openObjectsRoot(projectDir, true)
	if err != nil {
		return nil, err
	}
	f, name, err := createTempIn(root, prefix)
	if err != nil {
		_ = root.Close()
		return nil, err
	}
	return &Temp{root: root, f: f, name: name}, nil
}

// Write appends p to the temp file.
func (t *Temp) Write(p []byte) (int, error) { return t.f.Write(p) }

// Close syncs and closes the temp file. A second call returns nil. The handle
// on the objects folder stays open until Install or Discard.
func (t *Temp) Close() error {
	if t.closed {
		return nil
	}
	t.closed = true
	if err := t.f.Sync(); err != nil {
		_ = t.f.Close()
		return err
	}
	return t.f.Close()
}

// Install moves the temp file to rel (slash-separated, starting with objects/)
// unless a file there already matches checksum. wrote is true when this call
// left new bytes at rel. The temp is gone afterward.
func (t *Temp) Install(rel, checksum string) (wrote bool, err error) {
	inner, err := innerPath(rel)
	if err != nil {
		return false, err
	}
	if !t.closed {
		t.closed = true
		_ = t.f.Close()
	}
	wrote, err = installTemp(t.root, t.name, inner, checksum)
	_ = t.root.Close()
	t.done = true
	return wrote, err
}

// Discard removes the temp file without installing it. It does nothing after
// Install or an earlier Discard.
func (t *Temp) Discard() {
	if t.done {
		return
	}
	if !t.closed {
		t.closed = true
		_ = t.f.Close()
	}
	_ = t.root.Remove(t.name)
	_ = t.root.Close()
	t.done = true
}

// WriteIfAbsent writes data at rel under projectDir unless a file there already
// matches checksum (lowercase hex SHA-256 of data). rel is the slash-separated
// path from the project folder and must start with objects/. It writes a temp
// file, syncs it, and renames it into place, so a reader never sees a partial
// object; every step stays inside projectDir/objects. wrote is true when this
// call left new bytes at rel. A checksum that does not match data is an error
// and nothing is written. A directory-sync error is ignored.
func WriteIfAbsent(projectDir, rel, checksum string, data []byte) (wrote bool, err error) {
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != checksum {
		return false, fmt.Errorf("objectstore: data does not match checksum %s", checksum)
	}
	inner, err := innerPath(rel)
	if err != nil {
		return false, err
	}
	root, err := openObjectsRoot(projectDir, true)
	if err != nil {
		return false, err
	}
	defer root.Close()
	if match, err := matches(root, inner, checksum); err != nil {
		return false, err
	} else if match {
		return false, nil
	}
	tmp, name, err := createTempIn(root, ".write-")
	if err != nil {
		return false, err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = root.Remove(name)
		return false, err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = root.Remove(name)
		return false, err
	}
	if err := tmp.Close(); err != nil {
		_ = root.Remove(name)
		return false, err
	}
	wrote, err = installTemp(root, name, inner, checksum)
	if wrote {
		return true, nil
	}
	return false, err
}
