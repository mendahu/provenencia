package handlers

import (
	"testing"

	"github.com/mendahu/provenance/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

func TestCompleteOnboarding(t *testing.T) {
	runRPC(t, CompleteOnboarding, []rpcTest{
		{
			name: "rejects blank names",
			reqFn: func(t *testing.T) proto.Message {
				return &engine.CompleteOnboardingRequest{IdentityDir: t.TempDir(), ParentDir: t.TempDir()}
			},
			wantErr: true,
		},
		{
			name: "writes project and identity",
			reqFn: func(t *testing.T) proto.Message {
				return &engine.CompleteOnboardingRequest{
					IdentityDir: t.TempDir(),
					ParentDir:   t.TempDir(),
					DisplayName: "Jake Robins",
					FamilyName:  "Robins Family",
				}
			},
			want: &engine.CompleteOnboardingResponse{DisplayName: "Jake Robins"},
			after: func(t *testing.T, out []byte, req proto.Message) {
				assertOnboardingWroteProject(t, out, "Robins Family")
				assertIdentityMatchesComplete(t, out, req)
			},
		},
		{
			name: "already exists",
			reqFn: func(t *testing.T) proto.Message {
				return &engine.CompleteOnboardingRequest{
					IdentityDir: t.TempDir(),
					ParentDir:   t.TempDir(),
					DisplayName: "Jake",
					FamilyName:  "Dup",
				}
			},
			calls:   2,
			wantErr: true,
		},
	})
}
