package core

import "testing"

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Fatal("Version must be set")
	}
}
