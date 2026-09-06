package identity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadJSONFileMissing(t *testing.T) {
	_, err := readJSONFile(filepath.Join(t.TempDir(), "nope.json"))
	if err != errFileNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestWriteJSONFileRoundTripAndRemove(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "f.json")
	if err := writeJSONFile(path, map[string]string{"k": "v"}); err != nil {
		t.Fatal(err)
	}
	b, err := readJSONFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("empty")
	}
	if err := removeJSONFile(path); err != nil {
		t.Fatal(err)
	}
	if _, err := readJSONFile(path); err != errFileNotFound {
		t.Fatalf("got %v", err)
	}
	if err := removeJSONFile(path); err != nil {
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
