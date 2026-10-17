package datevalue

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	n := func(v int) *int { return &v }
	tests := []struct {
		name    string
		v       Value
		wantErr bool
	}{
		{name: "point month 13", v: Value{Kind: KindPoint, StartYear: n(1882), StartMonth: n(13)}, wantErr: true},
		{name: "range start after end", v: Value{Kind: KindRange, StartYear: n(1900), EndYear: n(1890)}, wantErr: true},
		{name: "point about year", v: Value{Kind: KindPoint, Qualifier: QualifierABT, StartYear: n(1817)}},
		{name: "empty value", v: Value{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.v)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("Validate = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Validate = %v, want nil", err)
			}
		})
	}
}
