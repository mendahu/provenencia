package handlers

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
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
		id = uuidString(raw)
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
				// James Robins + Jim Robins: one name, from one Source, so single.
				if name.GetState() != "single" || len(name.Values) != 1 || name.Values[0].GetReason() != "kept" ||
					name.Values[0].GetValue().GetName().GetForm() != "James Jim Robins" {
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
		{"term", conclusiondetails.Value{TermID: term, TermKey: "male", TermLabel: "Male"}, func(v *engine.ConclusionValue) bool {
			tm := v.GetTerm()
			return tm.GetId() == uuidString(term) && tm.GetKey() == "male" && tm.GetLabel() == "Male"
		}},
		{"date", conclusiondetails.Value{Date: &datevalues.Value{Kind: "point", Calendar: "gregorian", StartYear: &year}}, func(v *engine.ConclusionValue) bool {
			return v.GetDate().GetStartYear() == 1890
		}},
		{"name", conclusiondetails.Value{Name: namevaluestest.Western("Jim Robins")}, func(v *engine.ConclusionValue) bool {
			n := v.GetName()
			return len(n.GetParts()) == 2 && n.GetParts()[1].GetType() == namevalues.PartTypeSurname
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := conclusionValueProto(tc.in)
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
	if got.GetMemberCount() != 4 || o.GetArtifactId() != uuidString(artifact) || o.GetVoteSupport() != 2 || o.GetVoteTotal() != 3 {
		t.Fatalf("%+v", got)
	}
}
