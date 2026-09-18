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
	tests := []struct {
		in   string
		want string
	}{
		{in: "image/jpeg", want: ".jpg"},
		{in: "image/jpg", want: ".jpg"},
		{in: "IMAGE/PNG", want: ".png"},
		{in: "video/quicktime", want: ".mov"},
		{in: "audio/x-wav", want: ".wav"},
		{in: "text/plain; charset=utf-8", want: ".txt"},
		{in: "text/csv", want: ".csv"},
		{in: "text/markdown", want: ".md"},
		{in: "application/msword", want: ".doc"},
		{in: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", want: ".docx"},
		{in: "audio/flac", want: ".flac"},
		{in: "image/heic", want: ".heic"},
		{in: "image/webp", want: ".webp"},
		{in: "image/gif", want: ".gif"},
		{in: "application/pdf", want: ".pdf"},
		{in: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := objectpath.Extension(tt.in); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}
