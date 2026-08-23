package handlers

import (
	"testing"

	"github.com/mendahu/provenance/api/proto/engine"
	"github.com/mendahu/provenance/core"
)

func TestGetVersion(t *testing.T) {
	runRPC(t, GetVersion, []rpcTest{
		{name: "matches core.Version", req: &engine.GetVersionRequest{}, want: &engine.GetVersionResponse{Version: core.Version}},
		{name: "empty payload", want: &engine.GetVersionResponse{Version: core.Version}},
	})
}
