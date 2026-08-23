package handlers

import (
	"testing"

	"github.com/mendahu/provenance/api/proto/engine"
	"github.com/mendahu/provenance/core/installstate"
	"github.com/mendahu/provenance/core/onboarding"
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
				assertActiveMatchesComplete(t, out, req)
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

func TestOpenProject(t *testing.T) {
	runRPC(t, OpenProject, []rpcTest{
		{
			name:    "bad proto",
			raw:     []byte{0xff, 0xff, 0xff, 0xff},
			wantErr: true,
		},
		{
			name: "opens existing",
			reqFn: func(t *testing.T) proto.Message {
				ident := t.TempDir()
				parent := t.TempDir()
				res, err := onboarding.Complete(ident, parent, "Jake", "One")
				if err != nil {
					t.Fatal(err)
				}
				if err := installstate.Remove(ident); err != nil {
					t.Fatal(err)
				}
				return &engine.OpenProjectRequest{
					IdentityDir: ident,
					ProjectDir:  res.ProjectDir,
				}
			},
			want: &engine.OpenProjectResponse{DisplayName: "Jake"},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var done engine.OpenProjectResponse
				if err := proto.Unmarshal(out, &done); err != nil {
					t.Fatal(err)
				}
				or := req.(*engine.OpenProjectRequest)
				got, err := GetActiveProject(marshalProto(t, &engine.GetActiveProjectRequest{IdentityDir: or.GetIdentityDir()}))
				if err != nil {
					t.Fatal(err)
				}
				assertWantFields(t, got, &engine.GetActiveProjectResponse{
					Found:      true,
					ProjectDir: done.GetProjectDir(),
				})
			},
		},
		{
			name: "not a project",
			reqFn: func(t *testing.T) proto.Message {
				dir, _ := saveIdentity(t, "Jake")
				return &engine.OpenProjectRequest{IdentityDir: dir, ProjectDir: t.TempDir()}
			},
			wantErr: true,
		},
		{
			name: "adopts user",
			reqFn: func(t *testing.T) proto.Message {
				other := t.TempDir()
				parent := t.TempDir()
				res, err := onboarding.Complete(other, parent, "Jake", "One")
				if err != nil {
					t.Fatal(err)
				}
				return &engine.OpenProjectRequest{
					IdentityDir: t.TempDir(),
					ProjectDir:  res.ProjectDir,
					AdoptUserId: res.Identity.UserID.String(),
				}
			},
			want: &engine.OpenProjectResponse{DisplayName: "Jake"},
		},
	})
}

func TestListProjectUsers(t *testing.T) {
	runRPC(t, ListProjectUsers, []rpcTest{
		{
			name:    "bad proto",
			raw:     []byte{0xff, 0xff, 0xff, 0xff},
			wantErr: true,
		},
		{
			name: "lists users",
			reqFn: func(t *testing.T) proto.Message {
				res, err := onboarding.Complete(t.TempDir(), t.TempDir(), "Jake", "One")
				if err != nil {
					t.Fatal(err)
				}
				return &engine.ListProjectUsersRequest{ProjectDir: res.ProjectDir}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.ListProjectUsersResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if len(resp.GetUsers()) != 1 || resp.GetUsers()[0].GetDisplayName() != "Jake" {
					t.Fatalf("%v", resp.GetUsers())
				}
			},
		},
	})
}
