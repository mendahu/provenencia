package handlers

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/subjects"
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

// personBeside creates another person Subject on the same Source as req's,
// named by forms (one Observation each), and returns its id.
func personBeside(t *testing.T, req *engine.PromoteSubjectRequest, forms ...string) string {
	t.Helper()
	first := uuid.MustParse(req.GetSubjectId())
	userID := uuid.MustParse(req.GetUserId())
	var id []byte
	if err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		var sourceID, typeID []byte
		if err := db.QueryRow(`SELECT source_id, subject_type_id FROM subjects WHERE id = ?`, first[:]).Scan(&sourceID, &typeID); err != nil {
			return err
		}
		s, err := subjects.Create(c, userID[:], subjects.CreateInput{SourceID: sourceID, SubjectTypeID: typeID}, nil)
		if err != nil {
			return err
		}
		id = s.ID
		if len(forms) == 0 {
			return nil
		}
		name, err := properties.Lookup(c, "name", properties.OriginProvenencia)
		if err != nil {
			return err
		}
		art, err := artifacts.Create(c, userID[:], artifacts.CreateInput{SourceID: sourceID, Label: "Scan"})
		if err != nil {
			return err
		}
		var in []observations.Input
		for _, form := range forms {
			in = append(in, observations.Input{SubjectID: s.ID, PropertyID: name.ID, Name: namevaluestest.Western(form)})
		}
		_, err = citations.CreateWithObservations(c, userID[:], citations.CreateInput{
			ArtifactID: art.ID, LocatorJSON: `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`,
		}, in)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return uuidString(id)
}

func promoteOK(t *testing.T, req *engine.PromoteSubjectRequest) *engine.PromoteSubjectResponse {
	t.Helper()
	out, err := PromoteSubject(marshalProto(t, req))
	if err != nil {
		t.Fatal(err)
	}
	var resp engine.PromoteSubjectResponse
	if err := proto.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	return &resp
}

func TestPromoteSubjectJoin(t *testing.T) {
	first := promotableSubject(t)
	minted := promoteOK(t, first)

	out, err := ListClaimConfidenceGrades(marshalProto(t, &engine.ListClaimConfidenceGradesRequest{ProjectDir: first.GetProjectDir()}))
	if err != nil {
		t.Fatal(err)
	}
	var grades engine.ListClaimConfidenceGradesResponse
	if err := proto.Unmarshal(out, &grades); err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, g := range grades.GetGrades() {
		keys = append(keys, g.GetKey())
	}
	if strings.Join(keys, ",") != "low_confidence,moderate,high_confidence" {
		t.Fatalf("grades %v", keys)
	}
	high := grades.GetGrades()[2]

	joined := promoteOK(t, &engine.PromoteSubjectRequest{
		ProjectDir: first.GetProjectDir(), UserId: first.GetUserId(),
		SubjectId: personBeside(t, first), EntityId: minted.Entity.GetId(),
		ConfidenceGradeId: high.GetId(), Argument: "Same name and age.",
	})
	if joined.Entity.GetRef() != minted.Entity.GetRef() || joined.Claim.GetEntityId() != minted.Entity.GetId() {
		t.Fatalf("joined %+v onto %+v", joined.Entity, minted.Entity)
	}
	if joined.Claim.GetConfidenceGradeId() != high.GetId() || joined.Claim.GetArgument() != "Same name and age." {
		t.Fatalf("claim %+v", joined.Claim)
	}

	for name, req := range map[string]*engine.PromoteSubjectRequest{
		"bad entity id": {ProjectDir: first.GetProjectDir(), UserId: first.GetUserId(), SubjectId: personBeside(t, first), EntityId: "nope"},
		"bad grade id":  {ProjectDir: first.GetProjectDir(), UserId: first.GetUserId(), SubjectId: personBeside(t, first), ConfidenceGradeId: "nope"},
	} {
		if _, err := PromoteSubject(marshalProto(t, req)); err == nil {
			t.Fatalf("%s accepted", name)
		}
	}
}

func TestListPromoteTargetSuggestions(t *testing.T) {
	first := promotableSubject(t)
	exact := promoteOK(t, &engine.PromoteSubjectRequest{
		ProjectDir: first.GetProjectDir(), UserId: first.GetUserId(), SubjectId: personBeside(t, first, "James Robins"),
	})
	shared := promoteOK(t, &engine.PromoteSubjectRequest{
		ProjectDir: first.GetProjectDir(), UserId: first.GetUserId(), SubjectId: personBeside(t, first, "Mary Robins"),
	})
	promoteOK(t, &engine.PromoteSubjectRequest{
		ProjectDir: first.GetProjectDir(), UserId: first.GetUserId(), SubjectId: personBeside(t, first, "Ada Lovelace"),
	})
	james := personBeside(t, first, "James Robins")

	runRPC(t, ListPromoteTargetSuggestions, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{name: "bad subject id", req: &engine.ListPromoteTargetSuggestionsRequest{ProjectDir: first.GetProjectDir(), SubjectId: "nope"}, wantErr: true},
		{
			name: "scored best first, with reasons and the person header",
			req:  &engine.ListPromoteTargetSuggestionsRequest{ProjectDir: first.GetProjectDir(), SubjectId: james},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.ListPromoteTargetSuggestionsResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				s := resp.GetSuggestions()
				if len(s) != 2 || s[0].Entity.GetRef() != exact.Entity.GetRef() || s[1].Entity.GetRef() != shared.Entity.GetRef() {
					t.Fatalf("%+v", s)
				}
				if s[0].GetScore() != 10 || s[0].GetPerson().GetName().GetForm() != "James Robins" || s[0].GetPerson().GetNameValueCount() != 1 {
					t.Fatalf("top %+v", s[0])
				}
				if r := s[0].GetReasons(); len(r) != 1 || r[0].GetPropertyKey() != "name" || r[0].GetSimilarity() != 1 || r[0].GetContribution() != 10 {
					t.Fatalf("reasons %+v", r)
				}
			},
		},
		{
			name: "limit",
			req:  &engine.ListPromoteTargetSuggestionsRequest{ProjectDir: first.GetProjectDir(), SubjectId: james, Limit: 1},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.ListPromoteTargetSuggestionsResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if len(resp.GetSuggestions()) != 1 {
					t.Fatalf("%+v", resp.GetSuggestions())
				}
			},
		},
	})
}

func TestListPromoteComparisonAndPins(t *testing.T) {
	first := promotableSubject(t)
	minted := promoteOK(t, &engine.PromoteSubjectRequest{
		ProjectDir: first.GetProjectDir(), UserId: first.GetUserId(),
		SubjectId: personBeside(t, first, "James Robins"), Argument: "The birth record.",
	})
	incoming := personBeside(t, first, "J. Robins", "Jack Robbins")

	out, err := ListPromoteComparison(marshalProto(t, &engine.ListPromoteComparisonRequest{
		ProjectDir: first.GetProjectDir(), SubjectId: incoming, EntityId: minted.Entity.GetId(),
	}))
	if err != nil {
		t.Fatal(err)
	}
	var cmp engine.ListPromoteComparisonResponse
	if err := proto.Unmarshal(out, &cmp); err != nil {
		t.Fatal(err)
	}
	if cmp.GetMemberCount() != 1 || len(cmp.GetProperties()) != 1 || cmp.GetProperties()[0].GetPropertyKey() != "name" {
		t.Fatalf("comparison %+v", &cmp)
	}
	var (
		marks []string
		pairs []*engine.ObservationPair
	)
	for _, in := range cmp.GetProperties()[0].GetIncoming() {
		if in.GetRecord().GetClaimId() != "" || in.GetRecord().GetSubjectId() != incoming {
			t.Fatalf("incoming record %+v", in.GetRecord())
		}
		for _, pr := range in.GetPairs() {
			if pr.GetMember().GetClaimId() != minted.Claim.GetId() || pr.GetMember().GetValue().GetName() == nil {
				t.Fatalf("member record %+v", pr.GetMember())
			}
			marks = append(marks, in.GetRecord().GetValue().GetName().GetForm()+":"+strconv.FormatBool(pr.GetCompatible()))
			if pr.GetCompatible() {
				pairs = append(pairs, &engine.ObservationPair{
					IncomingObservationId: in.GetRecord().GetObservationId(),
					MemberObservationId:   pr.GetMember().GetObservationId(),
				})
			}
		}
	}
	sort.Strings(marks)
	if strings.Join(marks, ",") != "J. Robins:true,Jack Robbins:false" {
		t.Fatalf("marks %v", marks)
	}

	joined := promoteOK(t, &engine.PromoteSubjectRequest{
		ProjectDir: first.GetProjectDir(), UserId: first.GetUserId(),
		SubjectId: incoming, EntityId: minted.Entity.GetId(), Pairs: pairs,
	})
	if joined.GetPinCount() != 2 {
		t.Fatalf("pin_count %d", joined.GetPinCount())
	}
	if err := withProjectCatalog(first.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		var n int
		claim := uuid.MustParse(minted.Claim.GetId())
		if err := db.QueryRow(`SELECT COUNT(*) FROM identity_claim_evidence WHERE identity_claim_id = ?`, claim[:]).Scan(&n); err != nil {
			return err
		}
		if n != 2 {
			t.Fatalf("member claim carries %d pins, want 2 (backfill)", n)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	for name, req := range map[string]*engine.ListPromoteComparisonRequest{
		"bad subject id": {ProjectDir: first.GetProjectDir(), SubjectId: "nope", EntityId: minted.Entity.GetId()},
		"bad entity id":  {ProjectDir: first.GetProjectDir(), SubjectId: incoming, EntityId: "nope"},
		"unknown entity": {ProjectDir: first.GetProjectDir(), SubjectId: incoming, EntityId: uuid.Must(uuid.NewV7()).String()},
	} {
		if _, err := ListPromoteComparison(marshalProto(t, req)); err == nil {
			t.Fatalf("%s accepted", name)
		}
	}
	if _, err := ListPromoteComparison([]byte{0xff}); err == nil {
		t.Fatal("bad proto accepted")
	}
	bad := &engine.PromoteSubjectRequest{
		ProjectDir: first.GetProjectDir(), UserId: first.GetUserId(),
		SubjectId: personBeside(t, first, "James Robins"), EntityId: minted.Entity.GetId(),
		Pairs: []*engine.ObservationPair{{IncomingObservationId: "nope", MemberObservationId: "nope"}},
	}
	if _, err := PromoteSubject(marshalProto(t, bad)); err == nil {
		t.Fatal("bad pair id accepted")
	}
}
