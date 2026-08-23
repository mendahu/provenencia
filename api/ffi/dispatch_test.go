package ffi

import (
	"testing"

	"github.com/mendahu/provenance/api/proto/engine"
	"github.com/mendahu/provenance/core"
	"google.golang.org/protobuf/proto"
)

func TestCall(t *testing.T) {
	mustMarshal := func(m proto.Message) []byte {
		t.Helper()
		b, err := proto.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}

	tests := []struct {
		name    string
		method  int32
		in      []byte
		wantErr bool
		check   func(t *testing.T, out []byte)
	}{
		{
			name:   "ping echoes message",
			method: MethodPing,
			in:     mustMarshal(&engine.PingRequest{Message: "hello"}),
			check: func(t *testing.T, out []byte) {
				var resp engine.PingResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.GetMessage() != "hello" {
					t.Fatalf("got %q", resp.GetMessage())
				}
			},
		},
		{
			name:   "ping echoes empty string",
			method: MethodPing,
			in:     mustMarshal(&engine.PingRequest{Message: ""}),
			check: func(t *testing.T, out []byte) {
				var resp engine.PingResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.GetMessage() != "" {
					t.Fatalf("got %q", resp.GetMessage())
				}
			},
		},
		{
			name:   "get version matches core.Version",
			method: MethodGetVersion,
			in:     mustMarshal(&engine.GetVersionRequest{}),
			check: func(t *testing.T, out []byte) {
				var resp engine.GetVersionResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.GetVersion() != core.Version {
					t.Fatalf("got %q want %q", resp.GetVersion(), core.Version)
				}
			},
		},
		{
			name:    "unknown method",
			method:  99,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := Call(tt.method, tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.check != nil {
				tt.check(t, out)
			}
		})
	}
}
