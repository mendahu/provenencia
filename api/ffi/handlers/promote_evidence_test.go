package handlers

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/observations/observationstest"
)

// evidenceFixture is a handle of three people who all say Mary Smith, male, on
// three Sources, and a fourth person on a Source of its own.
type evidenceFixture struct {
	dir, userID, typeID string
	handle              string
	members, claims     []string // Subject ids and their accepted claims' ids
	sourceID, incoming  string
}

func newEvidenceFixture(t *testing.T, polarity string) evidenceFixture {
	t.Helper()
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	dir, userID, _, typeID := subjectFixture(t)
	f := evidenceFixture{dir: dir, userID: userID, typeID: typeID}
	for i := range 3 {
		_, subject := personOnNewSource(t, dir, userID, typeID, "male", observations.PolarityPositive)
		req := &engine.PromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: subject}
		if i > 0 {
			req.EntityId = f.handle
		}
		res := promoteOK(t, req)
		f.handle = res.Entity.GetId()
		f.members = append(f.members, subject)
		f.claims = append(f.claims, res.Claim.GetId())
	}
	f.sourceID, f.incoming = personOnNewSource(t, dir, userID, typeID, "male", polarity)
	return f
}

func (f evidenceFixture) request() *engine.GetPromoteClaimEvidenceRequest {
	return &engine.GetPromoteClaimEvidenceRequest{ProjectDir: f.dir, SourceId: f.sourceID, SubjectId: f.incoming, HandleId: f.handle}
}

func (f evidenceFixture) evidence(t *testing.T) *engine.GetPromoteClaimEvidenceResponse {
	t.Helper()
	var resp engine.GetPromoteClaimEvidenceResponse
	mustRPC(t, GetPromoteClaimEvidence, f.request(), &resp)
	return &resp
}

func sectionOf(resp *engine.GetPromoteClaimEvidenceResponse, key string) *engine.PromoteClaimEvidenceSection {
	for _, s := range resp.GetSections() {
		if s.GetPropertyKey() == key {
			return s
		}
	}
	return nil
}

func offerOf(resp *engine.GetPromoteClaimEvidenceResponse, id string) *engine.PromoteClaimEvidenceOffer {
	for _, o := range resp.GetOffers() {
		if o.GetId() == id {
			return o
		}
	}
	return nil
}

// T-7: the RPC sends a row's lines against its target with the pins on offer.
func TestGetPromoteClaimEvidence(t *testing.T) {
	t.Run("a matching sex at birth is one checked line with a pin per member", func(t *testing.T) {
		f := newEvidenceFixture(t, observations.PolarityPositive)
		resp := f.evidence(t)
		s := sectionOf(resp, "sex_at_birth")
		if s == nil || len(s.GetLines()) != 1 {
			t.Fatalf("sections = %+v", resp.GetSections())
		}
		l := s.GetLines()[0]
		if l.GetOutcome() != "agree" || l.GetRelation() != "same" || !l.GetPinnable() || !l.GetChecked() || l.GetIncomingNegative() {
			t.Fatalf("line = %+v", l)
		}
		if l.GetIncoming().GetTerm().GetKey() != "male" || l.GetIncoming().GetTerm().GetLabel() != "Male" {
			t.Fatalf("incoming = %+v", l.GetIncoming())
		}
		if len(l.GetMembers()) != 3 {
			t.Fatalf("pins = %d, want 3", len(l.GetMembers()))
		}
		titles := map[string]bool{}
		for _, src := range resp.GetSources() {
			titles[src.GetId()] = src.GetTitle() != ""
		}
		for i, p := range l.GetMembers() {
			if want := "l:" + l.GetIncomingObservationId() + ":" + p.GetMemberObservationId(); p.GetId() != want {
				t.Errorf("pin id = %q, want %q", p.GetId(), want)
			}
			if p.GetMemberSubjectId() != f.members[i] || p.GetValue().GetTerm().GetKey() != "male" || !titles[p.GetSourceId()] {
				t.Errorf("pin %d = %+v", i, p)
			}
			if o := offerOf(resp, p.GetId()); o == nil || !o.GetChecked() {
				t.Errorf("pin %d offer = %+v", i, o)
			}
		}
		if len(resp.GetSources()) != 3 || len(resp.GetRelationships()) != 0 {
			t.Errorf("sources = %d, relationships = %d", len(resp.GetSources()), len(resp.GetRelationships()))
		}
	})

	t.Run("a negative sex at birth is a conflict that is not pinnable and offers nothing", func(t *testing.T) {
		f := newEvidenceFixture(t, observations.PolarityNegative)
		resp := f.evidence(t)
		s := sectionOf(resp, "sex_at_birth")
		if s == nil || len(s.GetLines()) != 1 {
			t.Fatalf("sections = %+v", resp.GetSections())
		}
		l := s.GetLines()[0]
		if l.GetOutcome() != "conflict" || l.GetPinnable() || l.GetChecked() || !l.GetIncomingNegative() {
			t.Fatalf("line = %+v", l)
		}
		for _, p := range l.GetMembers() {
			if offerOf(resp, p.GetId()) != nil {
				t.Errorf("a conflict pin %s is offered", p.GetId())
			}
		}
	})

	t.Run("a request that names nothing valid is promote.invalid", func(t *testing.T) {
		f := newEvidenceFixture(t, observations.PolarityPositive)
		otherSource, _ := personOnNewSource(t, f.dir, f.userID, f.typeID, "male", observations.PolarityPositive)
		for name, mutate := range map[string]func(*engine.GetPromoteClaimEvidenceRequest){
			"an unknown Source":           func(r *engine.GetPromoteClaimEvidenceRequest) { r.SourceId = uuid.Must(uuid.NewV7()).String() },
			"a Subject not on the Source": func(r *engine.GetPromoteClaimEvidenceRequest) { r.SourceId = otherSource },
			"an empty handle":             func(r *engine.GetPromoteClaimEvidenceRequest) { r.HandleId = "" },
			"an unknown handle":           func(r *engine.GetPromoteClaimEvidenceRequest) { r.HandleId = uuid.Must(uuid.NewV7()).String() },
		} {
			req := f.request()
			mutate(req)
			var resp engine.GetPromoteClaimEvidenceResponse
			err := rpc(t, GetPromoteClaimEvidence, req, &resp)
			var ae *apperr.Error
			if !errors.As(err, &ae) || ae.Code() != apperr.CodePromoteInvalid {
				t.Errorf("%s: err = %v, want %s", name, err, apperr.CodePromoteInvalid)
			}
		}
	})

	runRPC(t, GetPromoteClaimEvidence, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{name: "bad source id", req: &engine.GetPromoteClaimEvidenceRequest{SourceId: "nope"}, wantErr: true},
	})
}

// pinnedOn is the Observation ids pinned on a claim.
func pinnedOn(t *testing.T, dir, claimID string) map[string]bool {
	t.Helper()
	id := uuid.MustParse(claimID)
	out := map[string]bool{}
	if err := withProjectCatalog(dir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		ids, err := identityclaims.PinnedObservations(db, id[:])
		for _, o := range ids {
			out[database.UUIDString(o)] = true
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return out
}

// T-10: a likeness arrives as a depiction with the stored crop's path, on the
// line and on its pin.
func TestGetPromoteClaimEvidenceLikeness(t *testing.T) {
	f := newEvidenceFixture(t, observations.PolarityPositive)
	uid, member := uuid.MustParse(f.userID), uuid.MustParse(f.members[0])
	incoming, source := uuid.MustParse(f.incoming), uuid.MustParse(f.sourceID)
	err := withProjectCatalog(f.dir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		var id []byte
		if err := db.QueryRow(`SELECT source_id FROM subjects WHERE id = ?`, member[:]).Scan(&id); err != nil {
			return err
		}
		observationstest.SaveLikeness(t, c, uid[:], source[:], incoming[:], 40)
		observationstest.SaveLikeness(t, c, uid[:], id, member[:], 40)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	resp := f.evidence(t)
	s := sectionOf(resp, "likeness")
	if s == nil || len(s.GetLines()) != 1 {
		t.Fatalf("sections = %+v", resp.GetSections())
	}
	l := s.GetLines()[0]
	if l.GetOutcome() != "agree" || l.GetRelation() != "same" || !l.GetChecked() {
		t.Fatalf("line = %+v", l)
	}
	if l.GetIncoming().GetDepiction().GetImageRelPath() == "" || len(l.GetMembers()) != 1 ||
		l.GetMembers()[0].GetValue().GetDepiction().GetImageRelPath() != l.GetIncoming().GetDepiction().GetImageRelPath() {
		t.Fatalf("depiction paths: incoming %+v, pins %+v", l.GetIncoming(), l.GetMembers())
	}
}

// T-14: the sections follow the type's binding order, and a reorder shows on
// the next open with no restart.
func TestGetPromoteClaimEvidenceFollowsBindingOrder(t *testing.T) {
	f := newEvidenceFixture(t, observations.PolarityPositive)
	keys := func() []string {
		var out []string
		for _, s := range f.evidence(t).GetSections() {
			out = append(out, s.GetPropertyKey())
		}
		return out
	}
	if got := keys(); len(got) != 2 || got[0] != "name" || got[1] != "sex_at_birth" {
		t.Fatalf("sections = %v, want name then sex_at_birth", got)
	}
	var list engine.ListSubjectTypePropertiesResponse
	mustRPC(t, ListSubjectTypeProperties, &engine.ListSubjectTypePropertiesRequest{ProjectDir: f.dir, SubjectTypeId: f.typeID}, &list)
	var sex string
	var rest []string
	for _, p := range list.GetProperties() {
		if p.GetProperty().GetKey() == "sex_at_birth" {
			sex = p.GetProperty().GetId()
		} else {
			rest = append(rest, p.GetProperty().GetId())
		}
	}
	var reordered engine.ReorderSubjectTypePropertiesResponse
	mustRPC(t, ReorderSubjectTypeProperties, &engine.ReorderSubjectTypePropertiesRequest{
		ProjectDir: f.dir, UserId: f.userID, SubjectTypeId: f.typeID, PropertyIds: append([]string{sex}, rest...),
	}, &reordered)
	if got := keys(); len(got) != 2 || got[0] != "sex_at_birth" || got[1] != "name" {
		t.Fatalf("sections after reorder = %v, want sex_at_birth then name", got)
	}
}
