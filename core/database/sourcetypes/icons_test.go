package sourcetypes

import (
	"errors"
	"testing"
)

func TestNormalizeIconKey(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "empty defaults", raw: "", want: DefaultIconKey},
		{name: "whitespace defaults", raw: "  ", want: DefaultIconKey},
		{name: "known key", raw: "type_certificate", want: "type_certificate"},
		{name: "unknown", raw: "type_nope", wantErr: true},
		{name: "file key rejected", raw: "file_pdf", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeIconKey(tt.raw)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("got %q %v", got, err)
			}
		})
	}
}
