package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "is set",
			run: func(t *testing.T) {
				if Version == "" {
					t.Fatal("Version must be set")
				}
			},
		},
		{
			name: "matches root VERSION file",
			run: func(t *testing.T) {
				b, err := os.ReadFile(filepath.Join("..", "VERSION"))
				if err != nil {
					t.Fatalf("read VERSION: %v", err)
				}
				got := strings.TrimSpace(string(b))
				if got != Version {
					t.Fatalf("VERSION file %q != core.Version %q", got, Version)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
