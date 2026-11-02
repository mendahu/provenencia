package handlers

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"google.golang.org/protobuf/proto"
)

// promotableSubject creates a person Subject and returns its project dir, user and id.
func promotableSubject(t *testing.T) (dir, userID, subjectID string) {
	t.Helper()
	dir, userID, sourceID, typeID := subjectFixture(t)
	out, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
		ProjectDir: dir, UserId: userID, GraphId: soleGraphID(t, dir, sourceID), SubjectTypeId: typeID, Label: "James",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var created engine.CreateSubjectResponse
	if err := proto.Unmarshal(out, &created); err != nil {
		t.Fatal(err)
	}
	return dir, userID, created.Subject.GetId()
}

// promoteInDir files the Subject through the engine's filing rule, as a test seed: onto a new
// handle when entityID is empty, else onto that handle.
func promoteInDir(t *testing.T, dir, userID, subjectID, entityID string) promote.Result {
	t.Helper()
	subject, user := uuid.MustParse(subjectID), uuid.MustParse(userID)
	in := promote.Input{SubjectID: subject[:]}
	if entityID != "" {
		entity := uuid.MustParse(entityID)
		in.EntityID = entity[:]
	}
	var res promote.Result
	if err := withProjectCatalog(dir, func(c *database.Catalog) error {
		res = promotetest.Save(t, c, user[:], in)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return res
}

func TestUnpromoteSubject(t *testing.T) {
	promoted := func(t *testing.T) *engine.UnpromoteSubjectRequest {
		t.Helper()
		dir, userID, subjectID := promotableSubject(t)
		promoteInDir(t, dir, userID, subjectID, "")
		return &engine.UnpromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: subjectID}
	}
	memberships := func(t *testing.T, dir, subjectID string) int {
		t.Helper()
		sources, err := ListSources(marshalProto(t, &engine.ListSourcesRequest{ProjectDir: dir}))
		if err != nil {
			t.Fatal(err)
		}
		var list engine.ListSourcesResponse
		if err := proto.Unmarshal(sources, &list); err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, src := range list.Sources {
			out, err := ListSubjectMemberships(marshalProto(t, &engine.ListSubjectMembershipsRequest{ProjectDir: dir, GraphId: soleGraphID(t, dir, src.GetId())}))
			if err != nil {
				t.Fatal(err)
			}
			var resp engine.ListSubjectMembershipsResponse
			if err := proto.Unmarshal(out, &resp); err != nil {
				t.Fatal(err)
			}
			for _, m := range resp.Memberships {
				if m.GetSubjectId() == subjectID {
					n++
				}
			}
		}
		return n
	}
	runRPC(t, UnpromoteSubject, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "bad subject id",
			reqFn: func(t *testing.T) proto.Message {
				req := promoted(t)
				req.SubjectId = "nope"
				return req
			},
			wantErr: true,
		},
		{
			name: "unknown subject",
			reqFn: func(t *testing.T) proto.Message {
				req := promoted(t)
				req.SubjectId = uuid.Must(uuid.NewV7()).String()
				return req
			},
			wantErr:   true,
			wantErrIs: promote.ErrInvalid,
		},
		{
			name:  "a filed subject is unfiled and the write reports effects",
			reqFn: func(t *testing.T) proto.Message { return promoted(t) },
			after: func(t *testing.T, out []byte, req proto.Message) {
				var resp engine.UnpromoteSubjectResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				e := resp.GetEffects()
				if len(e.GetEvidenceGraphs()) == 0 || len(e.GetHandles()) == 0 {
					t.Fatalf("effects %+v", e)
				}
				in := req.(*engine.UnpromoteSubjectRequest)
				if n := memberships(t, in.GetProjectDir(), in.GetSubjectId()); n != 0 {
					t.Fatalf("still a member %d times", n)
				}
			},
		},
		{
			name:  "a second unpromote writes nothing and succeeds",
			reqFn: func(t *testing.T) proto.Message { return promoted(t) },
			calls: 2,
		},
		{
			name: "a bridge is unsupported",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, _ := subjectFixture(t)
				typesOut, err := ListSubjectTypes(marshalProto(t, &engine.ListSubjectTypesRequest{ProjectDir: dir}))
				if err != nil {
					t.Fatal(err)
				}
				var types engine.ListSubjectTypesResponse
				if err := proto.Unmarshal(typesOut, &types); err != nil {
					t.Fatal(err)
				}
				relID := ""
				for _, typ := range types.Types {
					if typ.GetKey() == "relationship" {
						relID = typ.GetId()
					}
				}
				out, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
					ProjectDir: dir, UserId: userID, GraphId: soleGraphID(t, dir, sourceID), SubjectTypeId: relID, Label: "bridge",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSubjectResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.UnpromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: created.Subject.GetId()}
			},
			wantErr:   true,
			wantErrIs: promote.ErrUnsupportedType,
		},
	})
}

func TestGetDeleteImpactPromotedSubject(t *testing.T) {
	dir, userID, subjectID := promotableSubject(t)
	promoted := promoteInDir(t, dir, userID, subjectID, "")
	out, err := GetDeleteImpact(marshalProto(t, &engine.GetDeleteImpactRequest{
		ProjectDir: dir, Kind: "subject", Id: subjectID,
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
		len(g.GetListed()) != 1 || g.GetListed()[0].GetRef() != promoted.Entity.Ref {
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
				ProjectDir: dir, UserId: userID, GraphId: soleGraphID(t, dir, sourceID), SubjectTypeId: typeID, Label: label,
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
		promoted := promoteInDir(t, dir, userID, f.promotedID, "")
		f.handleRef = promoted.Entity.Ref
		f.claimID = database.UUIDString(promoted.Claim.ID)
		return f
	}

	t.Run("promoted listed with handle and kind, unpromoted absent", func(t *testing.T) {
		f := setup(t)
		out, err := ListSubjectMemberships(marshalProto(t, &engine.ListSubjectMembershipsRequest{
			ProjectDir: f.dir, GraphId: soleGraphID(t, f.dir, f.sourceID),
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
				return &engine.ListSubjectMembershipsRequest{ProjectDir: dir, GraphId: "nope"}
			},
			wantErr: true,
		},
	})
}
