//go:build unix

package objectstore_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
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

// A file system that cannot hold a permission bit (exFAT, FAT) reports every
// file as executable, so the execute bit alone never refuses an object.
func TestOpenObjectIgnoresExecuteBit(t *testing.T) {
	const rel = "objects/ab/cd/abcd.png"
	p := t.TempDir()
	path := writeObject(t, p, rel, []byte("data"))
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := objectstore.OpenObject(p, rel)
	if err != nil {
		t.Fatalf("OpenObject with an execute bit: %v", err)
	}
	defer f.Close()
	got, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "data" {
		t.Fatalf("read %q, want %q", got, "data")
	}
}
