package depictionvalue

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/depictionvalue/depictionvaluetest"
)

func TestValidate(t *testing.T) {
	id := []byte("0123456789abcdef")
	tests := []struct {
		name    string
		v       Value
		invalid bool
	}{
		{name: "empty region", v: Value{RegionJSON: "", ImageFileID: id}, invalid: true},
		{name: "short image id", v: Value{RegionJSON: depictionvaluetest.Region, ImageFileID: id[:15]}, invalid: true},
		{name: "no image id", v: Value{RegionJSON: depictionvaluetest.Region}, invalid: true},
		{name: "storable", v: Value{RegionJSON: depictionvaluetest.Region, ImageFileID: id}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.v)
			if !tt.invalid {
				if err != nil {
					t.Fatalf("got %v", err)
				}
				return
			}
			if !errors.Is(err, ErrInvalid) {
				t.Fatalf("got %v", err)
			}
			if got := apperr.From(err).Code(); got != apperr.CodeDepictionValuesInvalid {
				t.Fatalf("code %v", got)
			}
		})
	}
}
