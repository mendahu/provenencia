package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Fatal("Version must be set")
	}
}

func TestVersionMatchesRootFile(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "VERSION"))
	if err != nil {
		t.Fatalf("read VERSION: %v", err)
	}
	got := strings.TrimSpace(string(b))
	if got != Version {
		t.Fatalf("VERSION file %q != core.Version %q", got, Version)
	}
}
