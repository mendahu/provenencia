//go:build unix

package objectstore_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/objectstore"
)

func TestOpenObjectFIFO(t *testing.T) {
	const rel = "objects/ab/cd/abcd.png"
	p := t.TempDir()
	path := filepath.Join(p, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(path, 0o644); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		f, err := objectstore.OpenObject(p, rel)
		if f != nil {
			_ = f.Close()
		}
		done <- err
	}()
	select {
	case err := <-done:
		if !errors.Is(err, objectstore.ErrNotRegular) {
			t.Fatalf("err %v, want ErrNotRegular", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("OpenObject blocked on a FIFO")
	}
}

// writeObject installs data at rel through the store and returns its absolute path.
func writeObject(t *testing.T, projectDir, rel string, data []byte) string {
	t.Helper()
	sum := sha256.Sum256(data)
	if _, err := objectstore.WriteIfAbsent(projectDir, rel, hex.EncodeToString(sum[:]), data); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(projectDir, filepath.FromSlash(rel))
}

func TestOpenObjectExecutable(t *testing.T) {
	const rel = "objects/ab/cd/abcd.png"
	p := t.TempDir()
	path := writeObject(t, p, rel, []byte("data"))
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := objectstore.OpenObject(p, rel)
	if f != nil {
		_ = f.Close()
	}
	if !errors.Is(err, objectstore.ErrExecutable) || !errors.Is(err, objectstore.ErrNotRegular) {
		t.Fatalf("err %v, want ErrExecutable and ErrNotRegular", err)
	}
}

func TestWriteIfAbsentReplacesExecutable(t *testing.T) {
	const rel = "objects/ab/cd/abcd.png"
	data := []byte("data")
	sum := sha256.Sum256(data)
	checksum := hex.EncodeToString(sum[:])
	p := t.TempDir()
	path := writeObject(t, p, rel, data)
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatal(err)
	}
	wrote, err := objectstore.WriteIfAbsent(p, rel, checksum, data)
	if err != nil {
		t.Fatal(err)
	}
	if !wrote {
		t.Fatal("wrote = false, want the executable object replaced")
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm()&0o111 != 0 {
		t.Fatalf("mode %v still has an execute bit", st.Mode().Perm())
	}
}
