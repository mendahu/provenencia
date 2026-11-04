package namevalue_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
)

func TestPartValueOK(t *testing.T) {
	if !namevalue.PartValueOK(strings.Repeat("a", namevalue.MaxPartBytes)) {
		t.Fatal("at limit should be ok")
	}
	if namevalue.PartValueOK(strings.Repeat("a", namevalue.MaxPartBytes+1)) {
		t.Fatal("over limit should not be ok")
	}
}

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
