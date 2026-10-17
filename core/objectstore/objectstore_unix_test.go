//go:build unix

package objectstore_test

import (
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
