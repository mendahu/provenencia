package handlers

import (
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
)

func TestPing(t *testing.T) {
	runRPC(t, Ping, []rpcTest{
		{name: "echoes message", req: &engine.PingRequest{Message: "hello"}, want: &engine.PingResponse{Message: "hello"}, exact: true},
		{name: "echoes empty string", req: &engine.PingRequest{Message: ""}, want: &engine.PingResponse{Message: ""}, exact: true},
		{name: "rejects invalid protobuf", raw: []byte{0xff, 0xff, 0xff, 0xff}, wantErr: true},
	})
}
