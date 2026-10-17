package namevalue_test

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		v       namevalue.Value
		wantErr bool
	}{
		{
			name: "unknown part type",
			v: namevalue.Value{Form: "James Robins", Parts: []namevalue.Part{
				{Idx: 0, Value: "James", Type: "not_a_type"},
			}},
			wantErr: true,
		},
		{name: "form and parts empty", v: namevalue.Value{}, wantErr: true},
		{name: "western name", v: *namevaluetest.Western("James Robins")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := namevalue.Validate(tt.v)
			if tt.wantErr {
				if !errors.Is(err, namevalue.ErrInvalid) {
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
