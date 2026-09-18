package objectpath_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/objectpath"
)

func TestRel(t *testing.T) {
	const sum = "8fce3b0000000000000000000000000000000000000000000000000000000000"
	got, err := objectpath.Rel(sum, "image/jpeg")
	if err != nil {
		t.Fatal(err)
	}
	want := "objects/8f/ce/" + sum + ".jpg"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	got, err = objectpath.Rel(sum, "image/jpg")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("jpg alias: got %q want %q", got, want)
	}
	if _, err := objectpath.Rel("abcd", ""); err == nil {
		t.Fatal("short checksum: want error")
	}
}

func TestExtension(t *testing.T) {
	if got := objectpath.Extension("IMAGE/PNG"); got != ".png" {
		t.Fatalf("got %q", got)
	}
	if got := objectpath.Extension("text/plain; charset=utf-8"); got != ".txt" {
		t.Fatalf("got %q", got)
	}
	if got := objectpath.Extension(""); got != "" {
		t.Fatalf("got %q", got)
	}
	// Types searchindex previously inlined — must stay non-empty.
	for _, mt := range []string{"image/jpeg", "image/png", "image/webp", "image/gif", "application/pdf"} {
		if objectpath.Extension(mt) == "" {
			t.Fatalf("%s: empty extension", mt)
		}
	}
}
