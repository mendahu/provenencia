package handlers

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"google.golang.org/protobuf/proto"
)

// personEntityID returns the id of the handle namedPerson promoted.
func personEntityID(t *testing.T, dir, ref string) string {
	t.Helper()
	var id string
	if err := withProjectCatalog(dir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		var raw []byte
		if err := db.QueryRow(`SELECT id FROM canonical_entities WHERE ref = ?`, ref).Scan(&raw); err != nil {
			return err
		}
		id = database.UUIDString(raw)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestGetConclusionDetail(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	notFound := func(t *testing.T, err error) {
		t.Helper()
		var ae *apperr.Error
		if !errors.As(err, &ae) || ae.Code() != apperr.CodeConclusionDetailsNotFound {
			t.Fatalf("err = %v, want %s", err, apperr.CodeConclusionDetailsNotFound)
		}
	}
	runRPC(t, GetConclusionDetail, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "Person carries its fields, values and outcomes",
			reqFn: func(t *testing.T) proto.Message {
				dir, ref := namedPerson(t)
				return &engine.GetConclusionDetailRequest{ProjectDir: dir, EntityId: personEntityID(t, dir, ref)}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var d engine.ConclusionDetail
				if err := proto.Unmarshal(out, &d); err != nil {
					t.Fatal(err)
				}
				var name *engine.ConclusionField
				for _, f := range d.Fields {
					if f.GetPropertyKey() == "name" {
						name = f
					}
				}
				if d.Entity.GetRef() == "" || name == nil || d.GetMemberCount() != 1 {
					t.Fatalf("%+v", &d)
				}
				// James Robins + Kenneth Robins: one name, from one Source, so single.
				if name.GetState() != "single" || len(name.Values) != 1 || name.Values[0].GetReason() != "kept" ||
					name.Values[0].GetValue().GetName().GetForm() != "James Kenneth Robins" {
					t.Fatalf("name %+v", name)
				}
				if len(name.Outcomes) != 2 {
					t.Fatalf("outcomes %+v", name.Outcomes)
				}
				for _, o := range name.Outcomes {
					if o.GetValueRank() != 1 || o.GetObservationRef() == "" || o.GetSubjectRef() == "" ||
						o.GetSourceId() == "" || o.GetCitationId() == "" || o.GetArtifactId() == "" || o.GetRecorded().GetName() == nil ||
						o.GetVoteSupport() != 0 || o.GetVoteTotal() != 0 ||
						o.GetDeniedByObservationId() != "" {
						t.Fatalf("outcome %+v", o)
					}
				}
			},
		},
		{
			name: "unknown handle is not found",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, _ := subjectFixture(t)
				return &engine.GetConclusionDetailRequest{ProjectDir: dir, EntityId: "00000000-0000-0000-0000-000000000001"}
			},
			wantErr:   true,
			wantErrIs: conclusiondetails.ErrNotFound,
		},
	})
	t.Run("unknown handle is coded", func(t *testing.T) {
		dir, _, _, _ := subjectFixture(t)
		_, err := GetConclusionDetail(marshalProto(t, &engine.GetConclusionDetailRequest{ProjectDir: dir, EntityId: "00000000-0000-0000-0000-000000000001"}))
		notFound(t, err)
	})
	t.Run("malformed id is coded", func(t *testing.T) {
		dir, _, _, _ := subjectFixture(t)
		_, err := GetConclusionDetail(marshalProto(t, &engine.GetConclusionDetailRequest{ProjectDir: dir, EntityId: "nope"}))
		notFound(t, err)
	})
}

func TestConclusionValueProtoCoversEveryKind(t *testing.T) {
	year := 1890
	term := []byte{0: 1, 15: 2}
	cases := []struct {
		name  string
		in    conclusiondetails.Value
		check func(*engine.ConclusionValue) bool
	}{
		{"empty", conclusiondetails.Value{}, func(v *engine.ConclusionValue) bool { return v.GetKind() == nil }},
		{"text", conclusiondetails.Value{Text: "farmer", HasText: true}, func(v *engine.ConclusionValue) bool { return v.GetText() == "farmer" }},
		{"empty text", conclusiondetails.Value{HasText: true}, func(v *engine.ConclusionValue) bool {
			_, ok := v.GetKind().(*engine.ConclusionValue_Text)
			return ok
		}},
		{"integer", conclusiondetails.Value{Integer: 7, HasInteger: true}, func(v *engine.ConclusionValue) bool { return v.GetInteger() == 7 }},
		{"zero integer", conclusiondetails.Value{HasInteger: true}, func(v *engine.ConclusionValue) bool {
			_, ok := v.GetKind().(*engine.ConclusionValue_Integer)
			return ok
		}},
		{"term", conclusiondetails.Value{TermID: term, TermKey: "male", TermLabel: "Male", TermOrigin: "user"}, func(v *engine.ConclusionValue) bool {
			tm := v.GetTerm()
			return tm.GetId() == database.UUIDString(term) && tm.GetKey() == "male" && tm.GetLabel() == "Male" && tm.GetOrigin() == "user"
		}},
		{"term without an id", conclusiondetails.Value{TermKey: "male", TermLabel: "Male"}, func(v *engine.ConclusionValue) bool {
			tm := v.GetTerm()
			return tm != nil && tm.GetId() == "" && tm.GetKey() == "male" && tm.GetLabel() == "Male"
		}},
		{"date", conclusiondetails.Value{Date: &datevalue.Value{Kind: "point", Calendar: "gregorian", StartYear: &year}}, func(v *engine.ConclusionValue) bool {
			return v.GetDate().GetStartYear() == 1890
		}},
		{"name", conclusiondetails.Value{Name: namevaluetest.Western("Kenneth Robins")}, func(v *engine.ConclusionValue) bool {
			n := v.GetName()
			return len(n.GetParts()) == 2 && n.GetParts()[1].GetType() == namevalue.PartTypeSurname
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := conclusionValueProto(tc.in, nil)
			// Round-trip through the wire, as Swift receives it.
			b, err := proto.Marshal(v)
			if err != nil {
				t.Fatal(err)
			}
			var got engine.ConclusionValue
			if err := proto.Unmarshal(b, &got); err != nil {
				t.Fatal(err)
			}
			if !tc.check(&got) {
				t.Fatalf("%+v", &got)
			}
		})
	}
}

// A likeness has no field on conclusiondetails.Value: its depiction input is the
// whole result, before any other case.
func TestConclusionValueProtoDepiction(t *testing.T) {
	v := conclusionValueProto(conclusiondetails.Value{Text: "ignored", HasText: true}, &engine.DepictionValue{ImageRelPath: "objects/x.png"})
	b, err := proto.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var got engine.ConclusionValue
	if err := proto.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.GetDepiction().GetImageRelPath() != "objects/x.png" || got.GetDepiction().GetRegionJson() != "" || got.GetText() != "" {
		t.Fatalf("%+v", &got)
	}
}

func TestConclusionDetailProtoCarriesVoteArtifactAndMembers(t *testing.T) {
	artifact := []byte{0: 9, 15: 9}
	d := conclusiondetails.Detail{
		MemberCount: 4,
		Fields: []conclusiondetails.Field{{Outcomes: []conclusiondetails.Outcome{{
			Reason: "outvoted", ArtifactID: artifact, Vote: autoreconcile.Vote{Support: 2, Of: 3},
		}}}},
	}
	got := conclusionDetailProto(d)
	o := got.Fields[0].Outcomes[0]
	if got.GetMemberCount() != 4 || o.GetArtifactId() != database.UUIDString(artifact) || o.GetVoteSupport() != 2 || o.GetVoteTotal() != 3 {
		t.Fatalf("%+v", got)
	}
}

// A page reads its handle's header with its detail: one load, one error.
func TestGetConclusionDetailCarriesItsHeader(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	tests := []struct {
		name    string
		fixture func(t *testing.T) (dir, id string)
		check   func(t *testing.T, d *engine.ConclusionDetail, id string)
	}{
		{
			name:    "a Person carries the Person header",
			fixture: namedPersonID,
			check: func(t *testing.T, d *engine.ConclusionDetail, id string) {
				if h := d.GetPerson(); h.GetEntity().GetId() != id || h.GetName().GetForm() == "" {
					t.Fatalf("person header %+v", h)
				}
			},
		},
		{
			name:    "an Event carries the Event header and its title",
			fixture: citedEvent,
			check: func(t *testing.T, d *engine.ConclusionDetail, id string) {
				h := d.GetEvent()
				if h.GetEntity().GetId() != id || h.GetTitle().GetRule() != engine.EventTitleRule_EVENT_TITLE_RULE_RECORDED_NAME {
					t.Fatalf("event header %+v", h)
				}
			},
		},
		{
			name:    "a Place carries the Place header",
			fixture: citedPlace,
			check: func(t *testing.T, d *engine.ConclusionDetail, id string) {
				if h := d.GetPlace(); h.GetEntity().GetId() != id || len(h.GetNames()) == 0 {
					t.Fatalf("place header %+v", h)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, id := tt.fixture(t)
			out, err := GetConclusionDetail(marshalProto(t, &engine.GetConclusionDetailRequest{ProjectDir: dir, EntityId: id}))
			if err != nil {
				t.Fatal(err)
			}
			var d engine.ConclusionDetail
			if err := proto.Unmarshal(out, &d); err != nil {
				t.Fatal(err)
			}
			tt.check(t, &d, id)
		})
	}
}

func TestGetConclusionDetailListsLikenesses(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	dir, userID, _, artifactID, personID, likenessID := likenessFixture(t)
	saveLikeness(t, dir, userID, artifactID, personID, likenessID, writeCrop(t, 300, 200, 100))
	saveLikeness(t, dir, userID, artifactID, personID, likenessID, writeCrop(t, 300, 200, 200))
	promoted := promoteOK(t, &engine.PromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: personID})
	out, err := GetConclusionDetail(marshalProto(t, &engine.GetConclusionDetailRequest{ProjectDir: dir, EntityId: promoted.Entity.GetId()}))
	if err != nil {
		t.Fatal(err)
	}
	var d engine.ConclusionDetail
	if err := proto.Unmarshal(out, &d); err != nil {
		t.Fatal(err)
	}
	if len(d.GetPictures()) != 2 {
		t.Fatalf("pictures %+v", d.GetPictures())
	}
	for _, l := range d.GetPictures() {
		if !strings.HasPrefix(l.GetImageRelPath(), "objects/") || l.GetRecord().GetCitationId() == "" || l.GetRecord().GetSourceTitle() == "" {
			t.Fatalf("picture %+v", l)
		}
	}
	if d.GetPictures()[0].GetRecord().GetObservationId() >= d.GetPictures()[1].GetRecord().GetObservationId() {
		t.Fatalf("not earliest first: %+v", d.GetPictures())
	}
	if got := d.GetPerson().GetPictureImageRelPath(); got == "" || got != d.GetPictures()[0].GetImageRelPath() {
		t.Fatalf("person lead %q, first tile %q", got, d.GetPictures()[0].GetImageRelPath())
	}
}

func TestGetConclusionDetailEventPictureLead(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	dir, userID, _, artifactID, eventID, pictureID := depictionFixture(t, "event", "event_picture")
	saveLikeness(t, dir, userID, artifactID, eventID, pictureID, writeCrop(t, 300, 200, 100))
	saveLikeness(t, dir, userID, artifactID, eventID, pictureID, writeCrop(t, 300, 200, 200))
	promoted := promoteOK(t, &engine.PromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: eventID})
	out, err := GetConclusionDetail(marshalProto(t, &engine.GetConclusionDetailRequest{ProjectDir: dir, EntityId: promoted.Entity.GetId()}))
	if err != nil {
		t.Fatal(err)
	}
	var d engine.ConclusionDetail
	if err := proto.Unmarshal(out, &d); err != nil {
		t.Fatal(err)
	}
	var leadPath string
	for _, p := range d.GetPictures() {
		if p.GetLeads() {
			leadPath = p.GetImageRelPath()
			break
		}
	}
	if leadPath == "" {
		t.Fatalf("no leading picture among %+v", d.GetPictures())
	}
	if got := d.GetEvent().GetPictureImageRelPath(); got != leadPath {
		t.Fatalf("event lead %q, leading tile %q", got, leadPath)
	}
}

func TestSetPreferredLikeness(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	dir, userID, _, artifactID, personID, likenessID := likenessFixture(t)
	saveLikeness(t, dir, userID, artifactID, personID, likenessID, writeCrop(t, 300, 200, 100))
	b := saveLikeness(t, dir, userID, artifactID, personID, likenessID, writeCrop(t, 300, 200, 200)).GetObservations()[0].GetId()
	entity := promoteOK(t, &engine.PromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: personID}).Entity.GetId()
	set := func(observationID string) (*engine.SetPreferredLikenessResponse, error) {
		out, err := SetPreferredLikeness(marshalProto(t, &engine.SetPreferredLikenessRequest{
			ProjectDir: dir, UserId: userID, EntityId: entity, ObservationId: observationID,
		}))
		if err != nil {
			return nil, err
		}
		var resp engine.SetPreferredLikenessResponse
		if err := proto.Unmarshal(out, &resp); err != nil {
			t.Fatal(err)
		}
		return &resp, nil
	}
	detail := func() *engine.ConclusionDetail {
		t.Helper()
		out, err := GetConclusionDetail(marshalProto(t, &engine.GetConclusionDetailRequest{ProjectDir: dir, EntityId: entity}))
		if err != nil {
			t.Fatal(err)
		}
		var d engine.ConclusionDetail
		if err := proto.Unmarshal(out, &d); err != nil {
			t.Fatal(err)
		}
		return &d
	}
	contains := func(ids []string, id string) bool {
		for _, x := range ids {
			if x == id {
				return true
			}
		}
		return false
	}

	resp, err := set(b)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(resp.GetEffects().GetHandles(), entity) || !contains(resp.GetEffects().GetHandleScopes(), entity) {
		t.Fatalf("effects %+v", resp.GetEffects())
	}
	d := detail()
	if len(d.GetPictures()) != 2 || d.GetPictures()[0].GetPreferred() || !d.GetPictures()[1].GetPreferred() {
		t.Fatalf("pictures %+v", d.GetPictures())
	}
	if d.GetPictures()[0].GetLeads() || !d.GetPictures()[1].GetLeads() {
		t.Fatalf("leads after choosing B %+v", d.GetPictures())
	}
	if d.GetPerson().GetPictureImageRelPath() != d.GetPictures()[1].GetImageRelPath() {
		t.Fatalf("lead %q, second tile %q", d.GetPerson().GetPictureImageRelPath(), d.GetPictures()[1].GetImageRelPath())
	}

	resp, err = set(b)
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetEffects().GetRevision() != 0 {
		t.Fatalf("repeat wrote revision %d", resp.GetEffects().GetRevision())
	}

	if _, err := set(""); err != nil {
		t.Fatal(err)
	}
	d = detail()
	if d.GetPictures()[0].GetPreferred() || d.GetPictures()[1].GetPreferred() ||
		d.GetPerson().GetPictureImageRelPath() != d.GetPictures()[0].GetImageRelPath() {
		t.Fatalf("after clear %+v", d)
	}
	if !d.GetPictures()[0].GetLeads() || d.GetPictures()[1].GetLeads() {
		t.Fatalf("leads after clear %+v", d.GetPictures())
	}

	fresh, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	_, err = set(fresh.String())
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code() != apperr.CodeCanonicalEntitiesInvalid {
		t.Fatalf("err = %v, want %s", err, apperr.CodeCanonicalEntitiesInvalid)
	}

	if _, err := set(b); err != nil {
		t.Fatal(err)
	}
	if _, err := DeleteObservation(marshalProto(t, &engine.DeleteObservationRequest{ProjectDir: dir, UserId: userID, ObservationId: b})); err != nil {
		t.Fatal(err)
	}
	d = detail()
	if len(d.GetPictures()) != 1 || d.GetPerson().GetPictureImageRelPath() != d.GetPictures()[0].GetImageRelPath() || d.GetPictures()[0].GetPreferred() {
		t.Fatalf("after delete %+v", d)
	}
}
