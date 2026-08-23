package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/mendahu/provenance/api/ffi"
	"github.com/mendahu/provenance/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

func TestMain(m *testing.M) {
	if os.Getenv("CGO_ENABLED") == "0" {
		fmt.Fprintln(os.Stderr, "CGO_ENABLED=0: api/libprovenance tests require cgo")
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestProvenanceCall(t *testing.T) {
	mustMarshal := func(m proto.Message) []byte {
		t.Helper()
		b, err := proto.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}

	tests := []struct {
		name       string
		method     int32
		in         []byte
		wantStatus int
		wantOutNil bool
		check      func(t *testing.T, out []byte)
	}{
		{
			name:   "ping through malloc",
			method: ffi.MethodPing,
			in:     mustMarshal(&engine.PingRequest{Message: "hello"}),
			check: func(t *testing.T, out []byte) {
				want, err := ffi.Call(ffi.MethodPing, mustMarshal(&engine.PingRequest{Message: "hello"}))
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(out, want) {
					t.Fatalf("round-trip mismatch")
				}
			},
		},
		{
			name:   "get version through malloc",
			method: ffi.MethodGetVersion,
			in:     mustMarshal(&engine.GetVersionRequest{}),
			check: func(t *testing.T, out []byte) {
				want, err := ffi.Call(ffi.MethodGetVersion, mustMarshal(&engine.GetVersionRequest{}))
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(out, want) {
					t.Fatalf("round-trip mismatch")
				}
			},
		},
		{
			name:       "nil input ping",
			method:     ffi.MethodPing,
			in:         nil,
			wantOutNil: true,
			check: func(t *testing.T, out []byte) {
				want, err := ffi.Call(ffi.MethodPing, nil)
				if err != nil {
					t.Fatal(err)
				}
				if len(want) != 0 {
					t.Fatalf("ffi.Call produced %d bytes", len(want))
				}
			},
		},
		{
			name:       "empty input ping",
			method:     ffi.MethodPing,
			in:         []byte{},
			wantOutNil: true,
			check: func(t *testing.T, out []byte) {
				want, err := ffi.Call(ffi.MethodPing, []byte{})
				if err != nil {
					t.Fatal(err)
				}
				if len(want) != 0 {
					t.Fatalf("ffi.Call produced %d bytes", len(want))
				}
			},
		},
		{
			name:       "unknown method no leaked out",
			method:     99,
			in:         mustMarshal(&engine.PingRequest{}),
			wantStatus: 1,
			wantOutNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, out, outNil := callForTest(tt.method, tt.in)
			if status != tt.wantStatus {
				t.Fatalf("status %d want %d", status, tt.wantStatus)
			}
			if tt.wantOutNil && !outNil {
				t.Fatal("expected nil out pointer")
			}
			if tt.check != nil {
				if !tt.wantOutNil && outNil {
					t.Fatal("expected out pointer")
				}
				tt.check(t, out)
			}
		})
	}
}
