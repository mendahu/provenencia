package handlers

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
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
				hasCounts := false
				for _, list := range e.GetLists() {
					if list == "counts" {
						hasCounts = true
					}
				}
				if !hasCounts {
					t.Fatalf("lists %+v want counts", e.GetLists())
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
				relID := subjectTypeIDByKey(t, dir, "relationship")
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
	if g.GetVia() != deleteimpact.ViaLastMember || g.GetKind() != "canonical_entity" ||
		len(g.GetListed()) != 1 || g.GetListed()[0].GetRef() != promoted.Entity.Ref {
		t.Fatalf("%+v", g)
	}
}

func TestGetUnpromoteImpact(t *testing.T) {
	dir, userID, subjectID := promotableSubject(t)
	promoted := promoteInDir(t, dir, userID, subjectID, "")
	runRPC(t, GetUnpromoteImpact, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "malformed subject id",
			reqFn: func(t *testing.T) proto.Message {
				return &engine.GetUnpromoteImpactRequest{ProjectDir: dir, SubjectId: "nope"}
			},
			wantErr: true,
		},
		{
			name: "unknown subject",
			reqFn: func(t *testing.T) proto.Message {
				id, err := uuid.NewV7()
				if err != nil {
					t.Fatal(err)
				}
				return &engine.GetUnpromoteImpactRequest{
					ProjectDir: dir,
					SubjectId:  id.String(),
				}
			},
			wantErr: true,
		},
	})
	out, err := GetUnpromoteImpact(marshalProto(t, &engine.GetUnpromoteImpactRequest{
		ProjectDir: dir, SubjectId: subjectID,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var sole engine.GetUnpromoteImpactResponse
	if err := proto.Unmarshal(out, &sole); err != nil {
		t.Fatal(err)
	}
	if sole.Deletes == nil || sole.Deletes.GetVia() != deleteimpact.ViaLastMember ||
		sole.Deletes.GetKind() != "canonical_entity" || sole.Deletes.GetTotal() != 1 ||
		len(sole.Deletes.GetListed()) != 1 || sole.Deletes.GetListed()[0].GetRef() != promoted.Entity.Ref ||
		sole.Deletes.GetListed()[0].GetLocation().GetSection() != "persons" ||
		sole.Deletes.GetListed()[0].GetLocation().GetEntityId() == "" {
		t.Fatalf("%+v", sole.Deletes)
	}

	dir2, userID2, sourceID2, typeID2 := subjectFixture(t)
	createSubject := func(label string) string {
		out, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
			ProjectDir: dir2, UserId: userID2, GraphId: soleGraphID(t, dir2, sourceID2), SubjectTypeId: typeID2, Label: label,
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
	subjectA, subjectB := createSubject("James"), createSubject("Jim")
	res := promoteInDir(t, dir2, userID2, subjectA, "")
	promoteInDir(t, dir2, userID2, subjectB, database.UUIDString(res.Entity.ID))
	out, err = GetUnpromoteImpact(marshalProto(t, &engine.GetUnpromoteImpactRequest{
		ProjectDir: dir2, SubjectId: subjectA,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var twoMember engine.GetUnpromoteImpactResponse
	if err := proto.Unmarshal(out, &twoMember); err != nil {
		t.Fatal(err)
	}
	if twoMember.Deletes != nil {
		t.Fatalf("expected unset deletes, got %+v", twoMember.Deletes)
	}

	dir3, _, plainID := promotableSubject(t)
	out, err = GetUnpromoteImpact(marshalProto(t, &engine.GetUnpromoteImpactRequest{
		ProjectDir: dir3, SubjectId: plainID,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var unpromoted engine.GetUnpromoteImpactResponse
	if err := proto.Unmarshal(out, &unpromoted); err != nil {
		t.Fatal(err)
	}
	if unpromoted.Deletes != nil {
		t.Fatalf("unpromoted: %+v", unpromoted.Deletes)
	}

	dir4, userID4, sourceID4, typeID4 := subjectFixture(t)
	createPerson := func(label string) string {
		createdOut, createErr := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
			ProjectDir: dir4, UserId: userID4, GraphId: soleGraphID(t, dir4, sourceID4), SubjectTypeId: typeID4, Label: label,
		}))
		if createErr != nil {
			t.Fatal(createErr)
		}
		var created engine.CreateSubjectResponse
		if err := proto.Unmarshal(createdOut, &created); err != nil {
			t.Fatal(err)
		}
		return created.Subject.GetId()
	}
	jamesID := createPerson("James")
	promotedA := promoteInDir(t, dir4, userID4, jamesID, "")
	promotedC := promoteInDir(t, dir4, userID4, createPerson("Jim"), "")
	if err := withProjectCatalog(dir4, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		_, err = db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`,
			promotedA.Entity.ID, promotedC.Entity.ID)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	out, err = GetUnpromoteImpact(marshalProto(t, &engine.GetUnpromoteImpactRequest{
		ProjectDir: dir4, SubjectId: jamesID,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var mergedInto engine.GetUnpromoteImpactResponse
	if err := proto.Unmarshal(out, &mergedInto); err != nil {
		t.Fatal(err)
	}
	if mergedInto.Deletes != nil {
		t.Fatalf("merged into: %+v", mergedInto.Deletes)
	}

	runRPC(t, GetUnpromoteImpact, []rpcTest{
		{
			name: "bridge Subject",
			reqFn: func(t *testing.T) proto.Message {
				dir, user, sourceID, _ := subjectFixture(t)
				relID := subjectTypeIDByKey(t, dir, "relationship")
				out, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
					ProjectDir: dir, UserId: user, GraphId: soleGraphID(t, dir, sourceID), SubjectTypeId: relID, Label: "bridge",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSubjectResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.GetUnpromoteImpactRequest{ProjectDir: dir, SubjectId: created.Subject.GetId()}
			},
			wantErr:   true,
			wantErrIs: promote.ErrUnsupportedType,
		},
	})
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
