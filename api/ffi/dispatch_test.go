package ffi

import (
	"testing"

	"github.com/mendahu/provenance/api/proto/engine"
	"github.com/mendahu/provenance/core"
	"google.golang.org/protobuf/proto"
)

func TestCallPing(t *testing.T) {
	in, err := proto.Marshal(&engine.PingRequest{Message: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	out, err := Call(MethodPing, in)
	if err != nil {
		t.Fatal(err)
	}
	var resp engine.PingResponse
	if err := proto.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.GetMessage() != "hello" {
		t.Fatalf("got %q", resp.GetMessage())
	}
}

func TestCallGetVersion(t *testing.T) {
	in, err := proto.Marshal(&engine.GetVersionRequest{})
	if err != nil {
		t.Fatal(err)
	}
	out, err := Call(MethodGetVersion, in)
	if err != nil {
		t.Fatal(err)
	}
	var resp engine.GetVersionResponse
	if err := proto.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.GetVersion() != core.Version {
		t.Fatalf("got %q want %q", resp.GetVersion(), core.Version)
	}
}

func TestCallUnknownMethod(t *testing.T) {
	if _, err := Call(99, nil); err == nil {
		t.Fatal("expected error")
	}
}
