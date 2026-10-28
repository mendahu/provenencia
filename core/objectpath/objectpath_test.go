package objectpath_test

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/objectpath"
)

func TestRel(t *testing.T) {
	const sum = "8fce3b0000000000000000000000000000000000000000000000000000000000"
	const dir = "objects/8f/ce/" + sum
	tests := []struct {
		name      string
		checksum  string
		mediaType string
		want      string
		wantErr   error
	}{
		{name: "jpeg", checksum: sum, mediaType: "image/jpeg", want: dir + ".jpg"},
		{name: "pdf", checksum: sum, mediaType: "application/pdf", want: dir + ".pdf"},
		{name: "plain text with charset", checksum: sum, mediaType: "text/plain; charset=utf-8", want: dir + ".txt"},
		{name: "upper case type", checksum: sum, mediaType: "IMAGE/PNG", want: dir + ".png"},
		{name: "shell script", checksum: sum, mediaType: "application/x-sh", wantErr: objectpath.ErrDisallowedType},
		{name: "octet stream", checksum: sum, mediaType: "application/octet-stream", wantErr: objectpath.ErrDisallowedType},
		{name: "empty type", checksum: sum, mediaType: "", wantErr: objectpath.ErrDisallowedType},
		{name: "jpg alias is not stored", checksum: sum, mediaType: "image/jpg", wantErr: objectpath.ErrDisallowedType},
		{name: "short checksum", checksum: "abcd", mediaType: "image/png", wantErr: objectpath.ErrInvalidChecksum},
		{name: "short checksum, empty type", checksum: "abcd", mediaType: "", wantErr: objectpath.ErrInvalidChecksum},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := objectpath.Rel(tt.checksum, tt.mediaType)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err %v want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}
