package handlers

import (
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

func TestGetProjectInfo(t *testing.T) {
	runRPC(t, GetProjectInfo, []rpcTest{
		{
			name: "missing project",
			req: &engine.GetProjectInfoRequest{
				ProjectDir: filepath.Join(t.TempDir(), "nope.provenencia"),
			},
			wantErr: true,
		},
		{
			name: "after complete",
			reqFn: func(t *testing.T) proto.Message {
				parent := t.TempDir()
				ident := t.TempDir()
				out, err := CompleteOnboarding(marshalProto(t, &engine.CompleteOnboardingRequest{
					IdentityDir: ident,
					ParentDir:   parent,
					DisplayName: "Jake",
					FamilyName:  "Robins Family",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var done engine.CompleteOnboardingResponse
				if err := proto.Unmarshal(out, &done); err != nil {
					t.Fatal(err)
				}
				return &engine.GetProjectInfoRequest{ProjectDir: done.GetProjectDir()}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.GetProjectInfoResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				p := resp.GetProject()
				if p.GetLabel() != "Robins Family" {
					t.Fatalf("label %q", p.GetLabel())
				}
				if p.GetFolderName() != "robins-family.provenencia" {
					t.Fatalf("folder %q", p.GetFolderName())
				}
				if p.GetUpdatedByDisplayName() != "Jake" {
					t.Fatalf("updated_by %q", p.GetUpdatedByDisplayName())
				}
				if p.GetUpdatedByRef() == "" || p.GetCreatedAt() == "" {
					t.Fatalf("missing meta %+v", p)
				}
			},
		},
	})
}
