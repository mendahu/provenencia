package handlers

import (
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core"
)

func TestGetVersion(t *testing.T) {
	runRPC(t, GetVersion, []rpcTest{
		{name: "matches core.Version", req: &engine.GetVersionRequest{}, want: &engine.GetVersionResponse{Version: core.Version}},
		{name: "empty payload", want: &engine.GetVersionResponse{Version: core.Version}},
	})
}
