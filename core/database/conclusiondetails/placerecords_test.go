package conclusiondetails_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

func (f *fixture) prop(key string) properties.Property {
	f.t.Helper()
	p, err := properties.Lookup(f.c, key, vocabkey.OriginProvenencia)
	must(f.t, err)
	return p
}

func (f *fixture) bridge(src sources.Source, kind string, from, to subjects.Subject, obs ...observations.Input) {
	f.t.Helper()
	_, err := writes.Call(f.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, userID, connect.CreateInput{
			FromSubjectID: from.ID, ToSubjectID: to.ID, BridgeTypeKey: kind,
			Citation:     citations.CreateInput{ArtifactID: f.arts[string(src.ID)].ID, LocatorJSON: locator},
			Observations: obs,
		})
	})
	must(f.t, err)
}

func (f *fixture) at(s subjects.Subject, x, y int64) {
	f.t.Helper()
	_, err := writes.Call(f.c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
		return subjectpositions.Set(tx, s.ID, x, y)
	})
	must(f.t, err)
}

// birthAt files a Person born at a Place whose toponym records are given,
// promotes all three, and returns the Person's handle.
func (f *fixture) birthAt(src sources.Source, toponyms ...observations.Input) ([]byte, subjects.Subject) {
	f.t.Helper()
	person, birth, place := f.personOn(src), f.subjectOn(src, "event"), f.subjectOn(src, "place")
	f.at(person, 0, 0)
	f.at(birth, 4, 0)
	f.at(place, 8, 0)
	f.cite(person, f.nameIn(person, "form:Ann Lee"))
	eventType := f.prop("event_type")
	term, err := propertyterms.Lookup(f.c, eventType.ID, "birth", vocabkey.OriginProvenencia)
	must(f.t, err)
	f.cite(birth, observations.Input{SubjectID: birth.ID, PropertyID: eventType.ID, ValueTermID: term.ID})
	for _, in := range toponyms {
		in.SubjectID, in.PropertyID = place.ID, f.prop("toponym").ID
		f.cite(place, in)
	}
	// A blank toponym can't be filed through the write path; blank the marker.
	db, err := f.c.DB()
	must(f.t, err)
	_, err = db.Exec(`UPDATE observations SET value_text = '  ' || char(9) || ' ' WHERE value_text = ?`, blankMarker)
	must(f.t, err)
	role, err := propertyterms.Lookup(f.c, f.prop("role").ID, "subject", vocabkey.OriginProvenencia)
	must(f.t, err)
	f.bridge(src, "participation", person, birth,
		observations.Input{PropertyID: f.prop("person").ID, ValueSubjectID: person.ID},
		observations.Input{PropertyID: f.prop("event").ID, ValueSubjectID: birth.ID},
		observations.Input{PropertyID: f.prop("role").ID, ValueTermID: role.ID})
	f.bridge(src, "location", birth, place,
		observations.Input{PropertyID: f.prop("event").ID, ValueSubjectID: birth.ID},
		observations.Input{PropertyID: f.prop("place").ID, ValueSubjectID: place.ID})
	h := f.promote(person, nil)
	f.promote(birth, nil)
	f.promote(place, nil)
	return h, place
}

const blankMarker = "BLANK"

func text(s string) observations.Input { return observations.Input{ValueText: s, HasText: true} }

func (f *fixture) placeRows(person []byte) []conclusiondetails.OwnerField {
	f.t.Helper()
	rows, err := conclusiondetails.PersonDerivedRows(f.c, person)
	must(f.t, err)
	for _, r := range rows {
		if r.Kind == conclusiondetails.DerivedRowBirthPlace {
			return r.Owners
		}
	}
	return nil
}

func TestLinkRecordReadsAsItsPlacesToponym(t *testing.T) {
	neg := func(s string) observations.Input {
		in := text(s)
		in.Polarity = observations.PolarityNegative
		return in
	}
	tests := []struct {
		name     string
		toponyms []observations.Input
		want     string // "" = the lowest-ref positive, non-blank toponym is read off the catalog
		empty    bool
	}{
		{name: "one toponym", toponyms: []observations.Input{text("York")}, want: "York"},
		{name: "none reads empty", empty: true},
		{name: "a negative and a blank one are skipped", toponyms: []observations.Input{neg("Leeds"), text(blankMarker), text("York")}, want: "York"},
		{name: "several: the lowest ref", toponyms: []observations.Input{text("York"), text("Eboracum"), text("Jorvik")}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			h, place := f.birthAt(f.source("Register"), tc.toponyms...)
			owners := f.placeRows(h)
			if len(owners) != 1 || len(owners[0].Field.Outcomes) != 1 {
				t.Fatalf("owners %+v, want one link with one record", owners)
			}
			got := owners[0].Field.Outcomes[0].Recorded
			want := tc.want
			if want == "" && !tc.empty {
				db, err := f.c.DB()
				must(t, err)
				must(t, db.QueryRow(`SELECT value_text FROM observations WHERE subject_id = ? AND polarity = 'positive' AND property_id = ?
					ORDER BY ref COLLATE NOCASE LIMIT 1`, place.ID, f.prop("toponym").ID).Scan(&want))
			}
			if tc.empty {
				if got.HasText || got.Text != "" {
					t.Fatalf("recorded %+v, want empty", got)
				}
				return
			}
			if !got.HasText || got.Text != want {
				t.Fatalf("recorded %+v, want %q", got, want)
			}
		})
	}
}

func TestLinkRecordReadsKeepQueriesConstant(t *testing.T) {
	count := func(unrelated int) int {
		f := newFixture(t)
		src := f.source("Register")
		h, _ := f.birthAt(src, text("York"))
		owners := f.placeRows(h)
		for i := range unrelated {
			other := f.subjectOn(src, "place")
			f.cite(other, observations.Input{SubjectID: other.ID, PropertyID: f.prop("toponym").ID, ValueText: fmt.Sprintf("Town %d", i), HasText: true})
		}
		var ids [][]byte
		for _, o := range owners {
			ids = append(ids, o.Entity.ID)
		}
		n, err := conclusiondetails.CountFieldsOfQueries(f.c, ids, true)
		must(t, err)
		return n
	}
	if small, big := count(2), count(8); small != big {
		t.Fatalf("queries: %d with 2 unrelated toponyms, %d with 8", small, big)
	}
}

func TestForEntityDropsSubjectValuedFields(t *testing.T) {
	f := newFixture(t)
	src := f.source("Register")
	ann, bob := f.personOn(src), f.personOn(src)
	f.cite(ann, f.nameIn(ann, "form:Ann Lee"))
	godparent, _, err := writes.Run(f.c, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
		return properties.Create(tx, userID, "Godparent", propertymeta.ValueTypeSubject, "", propertymeta.CardinalitySingle)
	})
	must(t, err)
	must(t, subjectvocab.EnsureBinding(f.c, f.person.ID, godparent.ID, 9))
	f.c.Graph().Drop()
	f.cite(ann, observations.Input{SubjectID: ann.ID, PropertyID: godparent.ID, ValueSubjectID: bob.ID})
	h := f.promote(ann, nil)
	for _, fld := range f.detail(h).Fields {
		if fld.ValueType == "subject" {
			t.Fatalf("field %q of value type subject returned for a Person", fld.PropertyKey)
		}
	}
}

func TestToponymLoadUsesSubjectIndex(t *testing.T) {
	f := newFixture(t)
	plan, err := conclusiondetails.ToponymPlan(f.c)
	must(t, err)
	if !strings.Contains(plan, "observations_subject_id_idx") || strings.Contains(plan, "observations_property_id_idx") {
		t.Fatalf("toponym load plan:\n%s", plan)
	}
}
