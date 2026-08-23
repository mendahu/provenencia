package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenance/api/proto/engine"
	"github.com/mendahu/provenance/core/identity"
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
		{
			name: "corrupt json",
			reqFn: func(t *testing.T) proto.Message {
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, identity.FileName), []byte("{"), 0o600); err != nil {
					t.Fatal(err)
				}
				return &engine.GetInstallIdentityRequest{IdentityDir: dir}
			},
			wantErr: true,
		},
	})
}
