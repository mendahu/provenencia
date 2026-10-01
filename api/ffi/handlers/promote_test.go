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

func TestListSubjectMemberships(t *testing.T) {
	// fixture: one promoted and one unpromoted person Subject on a Source.
	type fixture struct {
		dir, sourceID, promotedID, plainID, handleRef, claimID string
	}
	setup := func(t *testing.T) fixture {
		t.Helper()
		dir, userID, sourceID, typeID := subjectFixture(t)
		create := func(label string) string {
			out, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
				ProjectDir: dir, UserId: userID, SourceId: sourceID, SubjectTypeId: typeID, Label: label,
			}))
			if err != nil {
				t.Fatal(err)
			}
			var created engine.CreateSubjectResponse
			if err := proto.Unmarshal(out, &created); err != nil {
				t.Fatal(err)
			}
			return created.Subject.GetId()
		}
		f := fixture{dir: dir, sourceID: sourceID, promotedID: create("James"), plainID: create("Jim")}
		out, err := PromoteSubject(marshalProto(t, &engine.PromoteSubjectRequest{
			ProjectDir: dir, UserId: userID, SubjectId: f.promotedID,
		}))
		if err != nil {
			t.Fatal(err)
		}
		var promoted engine.PromoteSubjectResponse
		if err := proto.Unmarshal(out, &promoted); err != nil {
			t.Fatal(err)
		}
		f.handleRef = promoted.Entity.GetRef()
		f.claimID = promoted.Claim.GetId()
		return f
	}

	t.Run("promoted listed with handle and kind, unpromoted absent", func(t *testing.T) {
		f := setup(t)
		out, err := ListSubjectMemberships(marshalProto(t, &engine.ListSubjectMembershipsRequest{
			ProjectDir: f.dir, SourceId: f.sourceID,
		}))
		if err != nil {
			t.Fatal(err)
		}
		var resp engine.ListSubjectMembershipsResponse
		if err := proto.Unmarshal(out, &resp); err != nil {
			t.Fatal(err)
		}
		if len(resp.Memberships) != 1 {
			t.Fatalf("%+v", resp.Memberships)
		}
		m := resp.Memberships[0]
		if m.GetSubjectId() != f.promotedID || m.GetKind() != "person" || m.GetClaimId() != f.claimID ||
			m.Entity.GetRef() != f.handleRef || !strings.HasPrefix(f.handleRef, "PER-") {
			t.Fatalf("%+v", m)
		}
		for _, other := range resp.Memberships {
			if other.GetSubjectId() == f.plainID {
				t.Fatal("unpromoted subject listed")
			}
		}
	})

	runRPC(t, ListSubjectMemberships, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "bad source id",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, _ := subjectFixture(t)
				return &engine.ListSubjectMembershipsRequest{ProjectDir: dir, SourceId: "nope"}
			},
			wantErr: true,
		},
	})
}
