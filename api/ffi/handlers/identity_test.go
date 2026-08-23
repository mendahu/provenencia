package handlers

import (
	"testing"

	"github.com/mendahu/provenance/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

func TestGetInstallIdentity(t *testing.T) {
	runRPC(t, GetInstallIdentity, []rpcTest{
		{
			name: "not found",
			reqFn: func(t *testing.T) proto.Message {
				return &engine.GetInstallIdentityRequest{IdentityDir: t.TempDir()}
			},
			after: assertIdentityNotFound,
		},
		{
			name: "found",
			reqFn: func(t *testing.T) proto.Message {
				dir, _ := saveIdentity(t, "Jake")
				return &engine.GetInstallIdentityRequest{IdentityDir: dir}
			},
			want: &engine.GetInstallIdentityResponse{Found: true, DisplayName: "Jake"},
		},
	})
}
