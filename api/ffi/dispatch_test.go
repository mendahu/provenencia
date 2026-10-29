package ffi

import (
	"bytes"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/apperr"
)

func TestCallRouter(t *testing.T) {
	tests := []struct {
		name   string
		method int32
	}{
		{name: "unspecified method", method: MethodUnspecified},
		{name: "unknown method", method: 100},
		{name: "reserved update citation with observations", method: 71},
		{name: "claim evidence with no request", method: MethodGetPromoteClaimEvidence},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Call(tt.method, nil); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestRecovered(t *testing.T) {
	boom := errors.New("e")
	tests := []struct {
		name     string
		fn       func() ([]byte, error)
		wantOut  []byte
		wantCode string
		wantErr  error
	}{
		{name: "panic", fn: func() ([]byte, error) { panic("x") }, wantCode: apperr.CodeInternalUnknown},
		{name: "value", fn: func() ([]byte, error) { return []byte{1}, nil }, wantOut: []byte{1}},
		{name: "error passes through", fn: func() ([]byte, error) { return nil, boom }, wantErr: boom},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := recovered(1, tt.fn)
			if !bytes.Equal(out, tt.wantOut) {
				t.Fatalf("out %v, want %v", out, tt.wantOut)
			}
			switch {
			case tt.wantCode != "":
				ae := apperr.From(err)
				if ae == nil || ae.Code() != tt.wantCode || ae.Kind() != apperr.KindInternal {
					t.Fatalf("err %v", err)
				}
			case tt.wantErr != nil:
				if err != tt.wantErr {
					t.Fatalf("err %v, want %v", err, tt.wantErr)
				}
			default:
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
