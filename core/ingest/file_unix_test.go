//go:build unix

package ingest

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/database"
)

func TestFileRejectsFIFO(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	fifo := filepath.Join(t.TempDir(), "pipe")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := File(c, fifo, userID)
		done <- err
	}()
	select {
	case err := <-done:
		if !errors.Is(err, ErrNotAFile) {
			t.Fatalf("got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("FIFO ingest hung")
	}
}

func TestOpenSourceFIFODoesNotHang(t *testing.T) {
	fifo := filepath.Join(t.TempDir(), "pipe")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		f, err := openSource(fifo)
		if err != nil {
			done <- err
			return
		}
		_ = f.Close()
		done <- nil
	}()
	select {
	case err := <-done:
		if err != nil {
			// Open may succeed with O_NONBLOCK; either way must not hang.
			t.Logf("openSource: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("openSource on FIFO hung")
	}
	_ = os.Remove(fifo)
}

func TestMapOpenErrUnix(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "eloop", err: syscall.ELOOP, want: ErrSymlink},
		{name: "eacces", err: syscall.EACCES, want: ErrPermissionDenied},
		{name: "eperm", err: syscall.EPERM, want: ErrPermissionDenied},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapOpenErr(tt.err)
			if !errors.Is(got, tt.want) {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
