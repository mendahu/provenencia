package database

import (
	"bytes"
	"testing"
)

func TestNullBlob(t *testing.T) {
	id := bytes.Repeat([]byte{0xab}, 16)
	tests := []struct {
		name string
		in   []byte
		want any
	}{
		{name: "nil", in: nil, want: nil},
		{name: "empty slice", in: []byte{}, want: nil},
		{name: "16-byte id", in: id, want: id},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NullBlob(tt.in)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("NullBlob(%v) = %v, want nil", tt.in, got)
				}
				return
			}
			b, ok := got.([]byte)
			if !ok || !bytes.Equal(b, tt.want.([]byte)) {
				t.Fatalf("NullBlob(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
