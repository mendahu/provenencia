package ffi_test

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/api/ffi"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"google.golang.org/protobuf/proto"
)

func TestEncodeError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantCode   string
		wantKind   engine.ErrorKind
		wantParams []string
	}{
		{
			name:     "sentinel",
			err:      database.ErrAlreadyExists,
			wantCode: apperr.CodeCatalogAlreadyExists,
			wantKind: engine.ErrorKind_ERROR_KIND_CONFLICT,
		},
		{
			name:       "with params",
			err:        database.ErrUnsupportedVersion.WithParams("9"),
			wantCode:   apperr.CodeCatalogUnsupportedVersion,
			wantKind:   engine.ErrorKind_ERROR_KIND_USER,
			wantParams: []string{"9"},
		},
		{
			name:     "unknown",
			err:      errors.New("sqlite boom"),
			wantCode: apperr.CodeInternalUnknown,
			wantKind: engine.ErrorKind_ERROR_KIND_INTERNAL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := ffi.EncodeError(tt.err)
			if err != nil {
				t.Fatal(err)
			}
			var msg engine.Error
			if err := proto.Unmarshal(b, &msg); err != nil {
				t.Fatal(err)
			}
			if msg.GetCode() != tt.wantCode {
				t.Fatalf("code = %q want %q", msg.GetCode(), tt.wantCode)
			}
			if msg.GetKind() != tt.wantKind {
				t.Fatalf("kind = %v want %v", msg.GetKind(), tt.wantKind)
			}
			if len(msg.GetParams()) != len(tt.wantParams) {
				t.Fatalf("params = %v want %v", msg.GetParams(), tt.wantParams)
			}
			for i, p := range tt.wantParams {
				if msg.GetParams()[i] != p {
					t.Fatalf("params[%d] = %q want %q", i, msg.GetParams()[i], p)
				}
			}
		})
	}
}

func TestCallUnknownMethod(t *testing.T) {
	_, err := ffi.Call(999, nil)
	if !errors.Is(err, apperr.New(apperr.CodeInternalUnknownMethod, apperr.KindInternal)) {
		t.Fatalf("got %v", err)
	}
}
