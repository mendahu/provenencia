package jsonfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadMissing(t *testing.T) {
	_, err := Read(filepath.Join(t.TempDir(), "nope.json"))
	if err != ErrNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestWriteRoundTripAndRemove(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "f.json")
	if err := Write(path, map[string]string{"k": "v"}); err != nil {
		t.Fatal(err)
	}
	b, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("empty")
	}
	if err := Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(path); err != ErrNotFound {
		t.Fatalf("got %v", err)
	}
	if err := Remove(path); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatal("dir")
	}
}
