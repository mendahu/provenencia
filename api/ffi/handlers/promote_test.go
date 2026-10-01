package handlers

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"google.golang.org/protobuf/proto"
)

// promotableSubject creates a person Subject and returns a promote request for it.
func promotableSubject(t *testing.T) *engine.PromoteSubjectRequest {
	t.Helper()
	dir, userID, sourceID, typeID := subjectFixture(t)
	out, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
		ProjectDir: dir, UserId: userID, SourceId: sourceID, SubjectTypeId: typeID, Label: "James",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var created engine.CreateSubjectResponse
	if err := proto.Unmarshal(out, &created); err != nil {
		t.Fatal(err)
	}
	return &engine.PromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: created.Subject.GetId()}
}

func TestPromoteSubject(t *testing.T) {
	runRPC(t, PromoteSubject, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "bad subject id",
			reqFn: func(t *testing.T) proto.Message {
				req := promotableSubject(t)
				req.SubjectId = "nope"
				return req
			},
			wantErr: true,
		},
		{
			name: "unknown subject",
			reqFn: func(t *testing.T) proto.Message {
				req := promotableSubject(t)
				req.SubjectId = uuid.Must(uuid.NewV7()).String()
				return req
			},
			wantErr: true,
		},
		{
			name:  "mints a handle and an accepted claim",
			reqFn: func(t *testing.T) proto.Message { return promotableSubject(t) },
			after: func(t *testing.T, out []byte, req proto.Message) {
				var resp engine.PromoteSubjectResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.Claim.GetStatus() != identityclaims.StatusAccepted {
					t.Fatalf("status %q", resp.Claim.GetStatus())
				}
				if !strings.HasPrefix(resp.Entity.GetRef(), "PER-") {
					t.Fatalf("ref %q", resp.Entity.GetRef())
				}
				in := req.(*engine.PromoteSubjectRequest)
				if resp.Claim.GetSubjectId() != in.GetSubjectId() || resp.Claim.GetEntityId() != resp.Entity.GetId() {
					t.Fatalf("%+v", resp.Claim)
				}
			},
		},
		{
			name:      "second promote refused",
			reqFn:     func(t *testing.T) proto.Message { return promotableSubject(t) },
			calls:     2,
			wantErr:   true,
			wantErrIs: identityclaims.ErrAlreadyMember,
		},
	})
}

func TestGetDeleteImpactPromotedSubject(t *testing.T) {
	req := promotableSubject(t)
	out, err := PromoteSubject(marshalProto(t, req))
	if err != nil {
		t.Fatal(err)
	}
	var promoted engine.PromoteSubjectResponse
	if err := proto.Unmarshal(out, &promoted); err != nil {
		t.Fatal(err)
	}
	out, err = GetDeleteImpact(marshalProto(t, &engine.GetDeleteImpactRequest{
		ProjectDir: req.GetProjectDir(), Kind: "subject", Id: req.GetSubjectId(),
	}))
	if err != nil {
		t.Fatal(err)
	}
	var report engine.GetDeleteImpactResponse
	if err := proto.Unmarshal(out, &report); err != nil {
		t.Fatal(err)
	}
	if !report.GetAllowed() || len(report.GetCascades()) != 1 {
		t.Fatalf("%+v", &report)
	}
	g := report.GetCascades()[0]
	if g.GetVia() != "identity_claims.subject_id" || g.GetKind() != "canonical_entity" ||
		len(g.GetListed()) != 1 || g.GetListed()[0].GetRef() != promoted.Entity.GetRef() {
		t.Fatalf("%+v", g)
	}
}
