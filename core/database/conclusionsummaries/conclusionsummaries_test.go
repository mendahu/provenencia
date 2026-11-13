package conclusionsummaries_test

import (
	"bytes"
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/observations/observationstest"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcecredibility"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

var userID = evidencelayertest.UserID

const locator = evidencelayertest.Locator

type fixture struct {
	*evidencelayertest.Records
	graph []byte // the Source's one Evidence graph
	name  properties.Property
}

func must(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	r := evidencelayertest.New(t)
	name, err := properties.Lookup(r.C, "name", vocabkey.OriginProvenencia)
	must(t, err)
	return &fixture{Records: r, graph: r.Graph(r.Source), name: name}
}

func (f *fixture) typeID(key string) []byte {
	st, err := subjecttypes.Lookup(f.C, key, vocabkey.OriginProvenencia)
	must(f.T, err)
	return st.ID
}

// person promotes a new person Subject named by forms (one Observation each).
func (f *fixture) person(forms ...string) (subjects.Subject, []observations.Observation, []byte) {
	f.T.Helper()
	personType := f.typeID("person")
	s := subjectstest.OfType(f.T, f.C, userID, f.graph, personType, "")
	var obs []observations.Observation
	if len(forms) > 0 {
		var in []observations.Input
		for _, form := range forms {
			in = append(in, observations.Input{SubjectID: s.ID, PropertyID: f.name.ID, Name: namevaluetest.Western(form)})
		}
		res, _, err := observationstest.Cite(f.C, userID, f.graph, f.Artifact.ID, locator, in...)
		must(f.T, err)
		obs = res.Observations
	}
	p := promotetest.Save(f.T, f.C, userID, promote.Input{SubjectID: s.ID})
	return s, obs, p.Entity.ID
}

func (f *fixture) list() []conclusionsummaries.PersonSummary {
	f.T.Helper()
	h, err := conclusionsummaries.ListPersons(f.C)
	must(f.T, err)
	return h
}

func TestListPersons(t *testing.T) {
	f := newFixture(t)
	if got := f.list(); len(got) != 0 {
		t.Fatalf("empty project listed %d", len(got))
	}

	f.person("Mary Smith")
	_, jamesObs, _ := f.person("James Robins", "Kenneth Robins", "james robins")
	p6a1 := canonicalentities.CreateInput{SubjectTypeID: f.typeID("person"), Label: "Mother of James"}
	labelled, err := writes.Call(f.C, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
		return canonicalentities.Create(tx, userID, p6a1)
	})
	must(t, err)
	p6a2 := canonicalentities.CreateInput{SubjectTypeID: f.typeID("person")}
	bare, err := writes.Call(f.C, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
		return canonicalentities.Create(tx, userID, p6a2)
	})
	must(t, err)
	p6a3 := canonicalentities.CreateInput{SubjectTypeID: f.typeID("place"), Label: "York"}
	_, err = writes.Call(f.C, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
		return canonicalentities.Create(tx, userID, p6a3)
	})
	must(t, err)

	got := f.list()
	if len(got) != 4 {
		t.Fatalf("listed %d Persons, want 4 (no Place)", len(got))
	}
	// No profile is installed, so parts stay in stored order (given, then
	// surname): James before Mary, then the label, then the ref-only row.
	// James, Kenneth and james reconcile into one name (S9-13b): Kenneth is a
	// different given name, so it joins rather than being outvoted.
	if got[0].Name == nil || got[0].Name.Form != "James Kenneth Robins" || got[0].NameValueCount != 1 {
		t.Fatalf("first %+v", got[0])
	}
	if got[1].Name == nil || got[1].Name.Form != "Mary Smith" || got[1].NameValueCount != 1 {
		t.Fatalf("second %+v", got[1])
	}
	if h := got[2]; h.Entity.Ref != labelled.Ref || h.Name != nil || h.NameValueCount != 0 || h.Entity.Label != "Mother of James" {
		t.Fatalf("labelled third %+v", h)
	}
	if h := got[3]; h.Entity.Ref != bare.Ref || h.Name != nil || h.Entity.Label != "" {
		t.Fatalf("bare last %+v", h)
	}

	t.Run("a name edit reaches the summary", func(t *testing.T) {
		// Kenneth → James merges the two values.
		_, err := writes.Call(f.C, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
			return observations.Update(tx, userID, observations.Input{
				ID: jamesObs[1].ID, SubjectID: jamesObs[1].SubjectID, PropertyID: f.name.ID,
				Name: namevaluetest.Western("James Robins"),
			})
		})
		must(t, err)
		if h := f.list()[0]; h.Name.Form != "James Robins" || h.NameValueCount != 1 {
			t.Fatalf("after edit %+v", h)
		}
	})
}

// The People list files by the sorted part order. Zachary Adams precedes
// Aaron Zimmerman because Adams precedes Zimmerman. A form with no parts is
// not a name the list files, so that person sorts last.
func TestListPersonsSortsBySortedPartsNotForm(t *testing.T) {
	f := newFixture(t)
	must(t, nameformats.Install(f.C))
	f.person("Aaron Zimmerman")
	f.person("Zachary Adams")
	formOnly := &namevalue.Value{Form: "Aaron Aardvark"}
	personType := f.typeID("person")
	s := subjectstest.OfType(t, f.C, userID, f.graph, personType, "")
	_, _, err := observationstest.Cite(f.C, userID, f.graph, f.Artifact.ID, locator, []observations.Input{{
		SubjectID: s.ID, PropertyID: f.name.ID, Name: formOnly,
	}}...)
	must(t, err)
	promotetest.Save(t, f.C, userID, promote.Input{SubjectID: s.ID})

	got := f.list()
	if len(got) != 3 {
		t.Fatalf("listed %d", len(got))
	}
	if got[0].Name == nil || got[0].Name.Form != "Zachary Adams" {
		t.Fatalf("first %+v", got[0].Name)
	}
	if got[1].Name == nil || got[1].Name.Form != "Aaron Zimmerman" {
		t.Fatalf("second %+v", got[1].Name)
	}
	if got[2].Name != nil {
		t.Fatalf("form-only name was filed %+v", got[2].Name)
	}
}

// +N counts displayed name values only. Names are one structure (S9-13b), so
// a named Person always counts one, whatever its records disagree on.
func TestPersonNameCountIsDisplayedOnly(t *testing.T) {
	f := newFixture(t)
	f.person("Ann Lee", "Anne Lee")
	f.person("Thomas Robins", "thomas robins", "Thomas Robbins")
	for _, h := range f.list() {
		if h.NameValueCount != 1 {
			t.Fatalf("%s counts %d", h.Name.Form, h.NameValueCount)
		}
	}
}

// A summary shows what the detail page displays: a kept value. When every
// value is held back (here, only a provisional member speaks), rank 1 is not
// kept, so the row falls back to the label or ref.
func TestSummariesShowKeptValuesOnly(t *testing.T) {
	tests := []struct {
		name  string
		kind  string
		cite  func(f *fixture) []observations.Input
		check func(t *testing.T, f *fixture, entityID []byte)
	}{
		{
			name: "a provisional member's name is not the Person's name",
			kind: "person",
			cite: func(f *fixture) []observations.Input {
				return []observations.Input{{PropertyID: f.name.ID, Name: namevaluetest.Western("Ghost Name")}}
			},
			check: func(t *testing.T, f *fixture, entityID []byte) {
				for _, h := range f.list() {
					if string(h.Entity.ID) == string(entityID) && h.Name != nil {
						t.Fatalf("listed provisional name %q", h.Name.Form)
					}
				}
			},
		},
		{
			name: "a provisional member's name, type, and date are not the Event's",
			kind: "event",
			cite: func(f *fixture) []observations.Input {
				return []observations.Input{
					{PropertyID: f.Prop("event_name").ID, Scalar: propertymeta.TextScalar("Phantom Fair")},
					{PropertyID: f.Prop("event_type").ID, ValueTermID: f.Term("event_type", "birth").ID},
					{PropertyID: f.Prop("date").ID, Date: datevaluetest.YMD(1801, 0, 0)},
				}
			},
			check: func(t *testing.T, f *fixture, entityID []byte) {
				for _, h := range f.events() {
					if string(h.Entity.ID) != string(entityID) {
						continue
					}
					if h.EventName != "" || h.Type.Term != nil || h.Date != nil {
						t.Fatalf("listed provisional values %+v", h)
					}
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			p6a4 := canonicalentities.CreateInput{SubjectTypeID: f.typeID(tt.kind)}
			handle, err := writes.Call(f.C, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
				return canonicalentities.Create(tx, userID, p6a4)
			})
			must(t, err)
			s := subjectstest.On(f.T, f.C, userID, f.graph, tt.kind, "")
			f.Cite(f.Artifact, s, tt.cite(f)...)
			_, err = writes.Call(f.C, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
				return identityclaims.Create(tx, userID, identityclaims.CreateInput{
					SubjectID: s.ID, EntityID: handle.ID, Status: identityclaims.StatusProvisional,
				})
			})
			must(t, err)
			db, err := f.C.DB()
			must(t, err)
			_, err = autoreconciler.RecomputeReported(db, [][]byte{handle.ID})
			must(t, err)
			tt.check(t, f, handle.ID)
		})
	}
}

func TestListPersonsQueryCountIsConstant(t *testing.T) {
	f := newFixture(t)
	// An open project holds its name format from open; a Drop leaves it.
	must(t, nameformats.Warm(f.C))
	gender := func(s subjects.Subject) {
		f.Cite(f.Artifact, s, observations.Input{PropertyID: f.Prop("gender").ID, ValueTermID: f.Term("gender", "woman").ID})
	}
	s, _, _ := f.person("Ada Lovelace")
	gender(s)
	f.C.Graph().Drop()
	one, err := conclusionsummaries.ListPersonsQueryCount(f.C)
	must(t, err)
	for i := 0; i < 50; i++ {
		s, _, _ := f.person(fmt.Sprintf("Person %d", i), fmt.Sprintf("Alias %d", i))
		gender(s)
	}
	f.C.Graph().Drop()
	many, err := conclusionsummaries.ListPersonsQueryCount(f.C)
	must(t, err)
	// The name row is one query; birth and death are one more, and the
	// likeness lead one more, each for the whole list. The count stays put as the list grows.
	if one < 1 || many != one {
		t.Fatalf("queries: %d for 1 Person, %d for 51", one, many)
	}
}

// subjectAndHandle creates and promotes a Subject of kind; returns subject and handle.
func (f *fixture) subjectAndHandle(kind string) (subjects.Subject, []byte) {
	f.T.Helper()
	s := subjectstest.OfType(f.T, f.C, userID, f.graph, f.typeID(kind), "")
	p := promotetest.Save(f.T, f.C, userID, promote.Input{SubjectID: s.ID})
	return s, p.Entity.ID
}

func pictureLeads[H any](rows []H, id, pic func(H) []byte) map[string][]byte {
	out := map[string][]byte{}
	for _, r := range rows {
		out[string(id(r))] = pic(r)
	}
	return out
}

type summaryPictureCase struct {
	name  string
	key   string
	setup func(f *fixture) (subjects.Subject, []byte)
	list  func(f *fixture) (map[string][]byte, error)
	byIDs func(f *fixture, ids [][]byte) (map[string][]byte, error)
}

func summaryPictureCases() []summaryPictureCase {
	personID := func(r conclusionsummaries.PersonSummary) []byte { return r.Entity.ID }
	personPic := func(r conclusionsummaries.PersonSummary) []byte { return r.PictureImageFileID }
	eventID := func(r conclusionsummaries.EventSummary) []byte { return r.Entity.ID }
	eventPic := func(r conclusionsummaries.EventSummary) []byte { return r.PictureImageFileID }
	placeID := func(r conclusionsummaries.PlaceSummary) []byte { return r.Entity.ID }
	placePic := func(r conclusionsummaries.PlaceSummary) []byte { return r.PictureImageFileID }
	return []summaryPictureCase{
		{
			name: "person",
			key:  vocabkey.PropertyLikeness,
			setup: func(f *fixture) (subjects.Subject, []byte) {
				s, _, h := f.person("Ann Lee")
				return s, h
			},
			list: func(f *fixture) (map[string][]byte, error) {
				return pictureLeads(f.list(), personID, personPic), nil
			},
			byIDs: func(f *fixture, ids [][]byte) (map[string][]byte, error) {
				rows, err := conclusionsummaries.PersonsByIDs(f.C, ids)
				return pictureLeads(rows, personID, personPic), err
			},
		},
		{
			name: "event",
			key:  vocabkey.PropertyEventPicture,
			setup: func(f *fixture) (subjects.Subject, []byte) {
				return f.subjectAndHandle("event")
			},
			list: func(f *fixture) (map[string][]byte, error) {
				return pictureLeads(f.events(), eventID, eventPic), nil
			},
			byIDs: func(f *fixture, ids [][]byte) (map[string][]byte, error) {
				rows, err := conclusionsummaries.EventsByIDs(f.C, ids)
				return pictureLeads(rows, eventID, eventPic), err
			},
		},
		{
			name: "place",
			key:  vocabkey.PropertyPlacePicture,
			setup: func(f *fixture) (subjects.Subject, []byte) {
				return f.subjectAndHandle("place")
			},
			list: func(f *fixture) (map[string][]byte, error) {
				return pictureLeads(f.places(), placeID, placePic), nil
			},
			byIDs: func(f *fixture, ids [][]byte) (map[string][]byte, error) {
				rows, err := conclusionsummaries.PlacesByIDs(f.C, ids)
				return pictureLeads(rows, placeID, placePic), err
			},
		},
	}
}

func TestSummaryLeadsWithEarliestPicture(t *testing.T) {
	for _, tc := range summaryPictureCases() {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			s, h := tc.setup(f)
			var other []byte
			if tc.name == "person" {
				_, _, other = f.person("Bob Lee")
			}
			imageOf := func(o observations.Observation) []byte { return observationstest.ImageOf(t, f.C, o.ID) }
			leads, err := tc.list(f)
			must(t, err)
			if got := leads[string(h)]; got != nil {
				t.Fatalf("lead before any picture: %x", got)
			}
			a := observationstest.SaveDepiction(t, f.C, userID, f.Source.ID, s.ID, tc.key, 100)
			leads, err = tc.list(f)
			must(t, err)
			if got := leads[string(h)]; !bytes.Equal(got, imageOf(a)) {
				t.Fatalf("lead after A: %x, want %x", got, imageOf(a))
			}
			observationstest.SaveDepiction(t, f.C, userID, f.Source.ID, s.ID, tc.key, 200)
			leads, err = tc.list(f)
			must(t, err)
			if got := leads[string(h)]; !bytes.Equal(got, imageOf(a)) {
				t.Fatalf("lead after B: %x, want %x", got, imageOf(a))
			}
			ids := [][]byte{h}
			if other != nil {
				ids = append(ids, other)
			}
			byIDs, err := tc.byIDs(f, ids)
			must(t, err)
			if got := byIDs[string(h)]; !bytes.Equal(got, imageOf(a)) {
				t.Fatalf("ByIDs lead %x", got)
			}
			if other != nil {
				if got := byIDs[string(other)]; got != nil {
					t.Fatalf("handle without a picture leads with %x", got)
				}
			}
			must(t, f.C.Graph().Verify())
		})
	}
}

func TestPersonSummarySkipsNegativeLikeness(t *testing.T) {
	f := newFixture(t)
	s, _, h := f.person("Ann Lee")
	observationstest.SaveNegativeLikeness(t, f.C, userID, f.Source.ID, s.ID, 100)
	b := observationstest.SaveLikeness(t, f.C, userID, f.Source.ID, s.ID, 200)
	observationstest.SaveLikeness(t, f.C, userID, f.Source.ID, s.ID, 100)
	var got []byte
	for _, r := range f.list() {
		if bytes.Equal(r.Entity.ID, h) {
			got = r.PictureImageFileID
		}
	}
	if want := observationstest.ImageOf(t, f.C, b.ID); !bytes.Equal(got, want) {
		t.Fatalf("lead %x, want %x", got, want)
	}
	must(t, f.C.Graph().Verify())
}

func TestSummaryLeadsWithChosenPicture(t *testing.T) {
	for _, tc := range summaryPictureCases() {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			s, h := tc.setup(f)
			a := observationstest.SaveDepiction(t, f.C, userID, f.Source.ID, s.ID, tc.key, 100)
			b := observationstest.SaveDepiction(t, f.C, userID, f.Source.ID, s.ID, tc.key, 200)
			imageOf := func(o observations.Observation) []byte { return observationstest.ImageOf(t, f.C, o.ID) }
			choose := func(obs []byte) {
				t.Helper()
				_, err := writes.Call(f.C, writes.Op{Action: "set_preferred_likeness", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					ch, err := canonicalentities.SetPreferredLikeness(tx, userID, h, obs)
					return struct{}{}, ch, err
				})
				must(t, err)
			}
			check := func(want observations.Observation, label string) {
				t.Helper()
				listLeads, err := tc.list(f)
				must(t, err)
				byIDLeads, err := tc.byIDs(f, [][]byte{h})
				must(t, err)
				for _, leads := range []map[string][]byte{listLeads, byIDLeads} {
					if got := leads[string(h)]; !bytes.Equal(got, imageOf(want)) {
						t.Fatalf("%s: lead %x, want %x", label, got, imageOf(want))
					}
				}
				must(t, f.C.Graph().Verify())
			}
			check(a, "no choice")
			choose(b.ID)
			check(b, "B chosen")
			choose(nil)
			check(a, "cleared")
		})
	}
}

func TestPersonsByIDs(t *testing.T) {
	f := newFixture(t)
	_, _, mary := f.person("Mary Smith")
	_, _, james := f.person("James Robins")
	f.person("Ada Lovelace")
	p6a9 := canonicalentities.CreateInput{SubjectTypeID: f.typeID("place")}
	place, err := writes.Call(f.C, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
		return canonicalentities.Create(tx, userID, p6a9)
	})
	must(t, err)

	got, err := conclusionsummaries.PersonsByIDs(f.C, [][]byte{mary, james, mary, place.ID, make([]byte, 16)})
	must(t, err)
	if len(got) != 2 || got[0].Name.Form != "James Robins" || got[1].Name.Form != "Mary Smith" {
		t.Fatalf("want James, Mary in list order: %+v", got)
	}
	if got, err := conclusionsummaries.PersonsByIDs(f.C, nil); err != nil || got != nil {
		t.Fatalf("no ids: %v %+v", err, got)
	}
}

// event promotes a new event Subject with the given Observations.
func (f *fixture) event(in ...observations.Input) []byte {
	f.T.Helper()
	eventType := f.typeID("event")
	s := subjectstest.OfType(f.T, f.C, userID, f.graph, eventType, "")
	if len(in) > 0 {
		for i := range in {
			in[i].SubjectID = s.ID
		}
		_, _, err := observationstest.Cite(f.C, userID, f.graph, f.Artifact.ID, locator, in...)
		must(f.T, err)
	}
	p := promotetest.Save(f.T, f.C, userID, promote.Input{SubjectID: s.ID})
	return p.Entity.ID
}

func (f *fixture) events() []conclusionsummaries.EventSummary {
	f.T.Helper()
	h, err := conclusionsummaries.ListEvents(f.C)
	must(f.T, err)
	return h
}

func TestListEvents(t *testing.T) {
	f := newFixture(t)
	if got := f.events(); len(got) != 0 {
		t.Fatalf("empty project listed %d", len(got))
	}
	name := f.Prop("event_name")
	date := f.Prop("date")
	start := f.Prop("start_date")
	end := f.Prop("end_date")
	eventType := f.Prop("event_type")
	birth := f.Term("event_type", "birth")

	fire := f.event(
		observations.Input{PropertyID: name.ID, Scalar: propertymeta.TextScalar(" The Great Fire ")},
		observations.Input{PropertyID: date.ID, Date: datevaluetest.YMD(1849, 0, 0)},
	)
	birthID := f.event(
		observations.Input{PropertyID: eventType.ID, ValueTermID: birth.ID},
		observations.Input{PropertyID: date.ID, Date: datevaluetest.YMD(1985, 0, 0)},
		observations.Input{PropertyID: start.ID, Date: datevaluetest.YMD(1900, 0, 0)},
	)
	span := f.event(
		observations.Input{PropertyID: start.ID, Date: datevaluetest.YMD(1900, 0, 0)},
		observations.Input{PropertyID: end.ID, Date: datevaluetest.YMD(1910, 0, 0)},
	)
	undated := f.event()
	f.person("Ada Lovelace")

	got := f.events()
	if len(got) != 4 {
		t.Fatalf("listed %d Events, want 4 (no Person)", len(got))
	}
	byID := map[string]conclusionsummaries.EventSummary{}
	for _, h := range got {
		byID[string(h.Entity.ID)] = h
	}
	if string(got[0].Entity.ID) != string(fire) || got[0].EventName != "The Great Fire" || got[0].Date == nil || *got[0].Date.StartYear != 1849 {
		t.Fatalf("fire %+v", got[0])
	}
	born := byID[string(birthID)]
	if born.Type.Term == nil || born.Type.Term.Key != "birth" || born.Type.Term.Label != "Birth" || born.Type.Term.Origin != "provenencia" || born.Title.Parts.TypeOrigin != "provenencia" || born.Type.Count != 1 {
		t.Fatalf("type %+v", born.Type)
	}
	if born.Date == nil || *born.Date.StartYear != 1985 || born.StartDate != nil || born.EndDate != nil || born.StartDateCount != 1 {
		t.Fatalf("date wins over span %+v", born)
	}
	spanned := byID[string(span)]
	if spanned.Date != nil || spanned.StartDate == nil || *spanned.StartDate.StartYear != 1900 || spanned.EndDate == nil || *spanned.EndDate.StartYear != 1910 {
		t.Fatalf("span %+v", spanned)
	}
	if byID[string(undated)].Date != nil || byID[string(undated)].EventName != "" {
		t.Fatalf("undated %+v", byID[string(undated)])
	}
	want := [][]byte{fire, span, birthID, undated}
	for i, id := range want {
		if string(got[i].Entity.ID) != string(id) {
			t.Fatalf("order %d got %x want %x", i, got[i].Entity.ID, id)
		}
	}
}

func TestListEventsQueryCountIsConstant(t *testing.T) {
	f := newFixture(t)
	f.event()
	f.C.Graph().Drop()
	one, err := conclusionsummaries.ListEventsQueryCount(f.C)
	must(t, err)
	date := f.Prop("date")
	for i := 0; i < 50; i++ {
		f.event(observations.Input{PropertyID: date.ID, Date: datevaluetest.YMD(1800+i, 0, 0)})
	}
	f.C.Graph().Drop()
	many, err := conclusionsummaries.ListEventsQueryCount(f.C)
	must(t, err)
	// The event row is one query; subjects and places are one each, for the
	// whole list. The count stays put as the list grows.
	if one < 1 || many != one {
		t.Fatalf("queries: %d for 1 Event, %d for 51", one, many)
	}
}

func TestEventsByIDs(t *testing.T) {
	f := newFixture(t)
	name := f.Prop("event_name")
	fire := f.event(observations.Input{PropertyID: name.ID, Scalar: propertymeta.TextScalar("Fire")})
	bare := f.event()
	f.event(observations.Input{PropertyID: name.ID, Scalar: propertymeta.TextScalar("Other")})
	_, _, person := f.person("Ada Lovelace")
	got, err := conclusionsummaries.EventsByIDs(f.C, [][]byte{bare, fire, fire, person, make([]byte, 16)})
	must(t, err)
	if len(got) != 2 {
		t.Fatalf("%+v", got)
	}
	names := map[string]bool{got[0].EventName: true, got[1].EventName: true}
	if !names["Fire"] || !names[""] {
		t.Fatalf("%+v", got)
	}
	if got, err := conclusionsummaries.EventsByIDs(f.C, nil); err != nil || got != nil {
		t.Fatalf("no ids: %v %+v", err, got)
	}
}

func (f *fixture) lowTrust(src sources.Source) {
	f.T.Helper()
	g, err := sourcecredibilitygrades.Lookup(f.C, "low_trust", vocabkey.OriginProvenencia)
	must(f.T, err)
	p6in11 := sourcecredibility.UpsertInput{
		SourceID: src.ID, CredibilityGradeID: g.ID,
	}
	p6act10 := "create_source_credibility_assessment"

	// place promotes a new Place whose toponyms are cited, in order, on art.
	if _, p6look12 := sourcecredibility.GetBySource(f.C, p6in11.SourceID); p6look12 == nil {
		p6act10 = "update_source_credibility_assessment"
	}
	_, err = writes.Call(f.C, writes.Op{Action: p6act10, UserID: userID}, func(tx *database.Tx) (sourcecredibility.Assessment, []rowchange.Change, error) {
		return sourcecredibility.Upsert(tx, userID, p6in11)
	})

	must(f.T, err)
}

func (f *fixture) place(art artifacts.Artifact, names ...string) []byte {
	f.T.Helper()
	placeType := f.typeID("place")
	graph := evidencegraphstest.SoleGraph(f.T, f.C, art.SourceID).ID
	s := subjectstest.OfType(f.T, f.C, userID, graph, placeType, "")
	f.citeToponyms(art, s.ID, names...)
	p := promotetest.Save(f.T, f.C, userID, promote.Input{SubjectID: s.ID})
	return p.Entity.ID
}

func (f *fixture) citeToponyms(art artifacts.Artifact, subjectID []byte, names ...string) {
	f.T.Helper()
	if len(names) == 0 {
		return
	}
	toponym := f.Prop("toponym")
	in := make([]observations.Input, len(names))
	for i, name := range names {
		in[i] = observations.Input{SubjectID: subjectID, PropertyID: toponym.ID, Scalar: propertymeta.TextScalar(name)}
	}
	graph := evidencegraphstest.SoleGraph(f.T, f.C, art.SourceID).ID
	_, _, err := observationstest.Cite(f.C, userID, graph, art.ID, locator, in...)
	must(f.T, err)
}

func (f *fixture) joinPlace(art artifacts.Artifact, entityID []byte, names ...string) {
	f.T.Helper()
	placeType := f.typeID("place")
	graph := evidencegraphstest.SoleGraph(f.T, f.C, art.SourceID).ID
	s := subjectstest.OfType(f.T, f.C, userID, graph, placeType, "")
	f.citeToponyms(art, s.ID, names...)
	promotetest.Save(f.T, f.C, userID, promote.Input{SubjectID: s.ID, EntityID: entityID})
}

func (f *fixture) places() []conclusionsummaries.PlaceSummary {
	f.T.Helper()
	h, err := conclusionsummaries.ListPlaces(f.C)
	must(f.T, err)
	return h
}

func TestPlaceKeepsEveryDistinctToponym(t *testing.T) {
	f := newFixture(t)
	must(t, sourcecredibilitygrades.Install(f.C))
	census, censusArt := f.NewSource("Census")
	gazette, gazetteArt := f.NewSource("Gazette")
	f.lowTrust(gazette)

	york := f.place(f.Artifact, "York", "york")
	f.joinPlace(censusArt, york, "Toronto")
	f.joinPlace(gazetteArt, york, "Muddy York")

	got := f.places()
	if len(got) != 1 || len(got[0].Names) != 2 || got[0].Names[0] != "York" || got[0].Names[1] != "Toronto" {
		t.Fatalf("names %+v", got)
	}
	if got[0].StartDate != nil || got[0].EndDate != nil || got[0].Kind != "" || len(got[0].Parents) != 0 {
		t.Fatalf("no period or parents without evidence: %+v", got[0])
	}

	detail, err := conclusiondetails.ForEntity(f.C, york)
	must(t, err)
	var weak string
	for _, field := range detail.Fields {
		if field.PropertyKey != "toponym" {
			continue
		}
		for _, value := range field.Values {
			if value.Reason == "weak" {
				weak = value.Value.Text
			}
		}
	}
	if weak != "Muddy York" {
		t.Fatalf("weak spelling %+v", detail.Fields)
	}
	_ = census
}

func TestPlaceSort(t *testing.T) {
	f := newFixture(t)
	york := f.place(f.Artifact, "York")
	montreal := f.place(f.Artifact, "montreal")
	bareA := f.place(f.Artifact)
	bareB := f.place(f.Artifact)
	f.event()

	got := f.places()
	if len(got) != 4 {
		t.Fatalf("listed %d, want 4 (no Event)", len(got))
	}
	if string(got[0].Entity.ID) != string(montreal) || got[0].Names[0] != "montreal" {
		t.Fatalf("named first by folded toponym: %+v", got[0])
	}
	if string(got[1].Entity.ID) != string(york) || got[1].Names[0] != "York" {
		t.Fatalf("York second: %+v", got[1])
	}
	unnamed := []string{string(bareA), string(bareB)}
	if string(got[2].Entity.ID) == unnamed[1] {
		unnamed[0], unnamed[1] = unnamed[1], unnamed[0]
	}
	// Refs increase with creation, so the earlier unnamed Place sorts first only
	// when its ref does. Compare the two trailing rows to the two bare ids,
	// ordered by the refs the query already applied.
	if string(got[2].Entity.ID) != unnamed[0] || string(got[3].Entity.ID) != unnamed[1] ||
		len(got[2].Names) != 0 || len(got[3].Names) != 0 || got[2].Entity.Ref > got[3].Entity.Ref {
		t.Fatalf("unnamed last by ref: %s %s then %s %s", got[2].Entity.Ref, got[2].Entity.ID, got[3].Entity.Ref, got[3].Entity.ID)
	}
}

func TestListPlacesQueryCountIsConstant(t *testing.T) {
	f := newFixture(t)
	f.place(f.Artifact, "York")
	f.C.Graph().Drop()
	one, err := conclusionsummaries.ListPlacesQueryCount(f.C)
	must(t, err)
	for i := 0; i < 20; i++ {
		f.place(f.Artifact, fmt.Sprintf("Place %d", i))
	}
	f.C.Graph().Drop()
	many, err := conclusionsummaries.ListPlacesQueryCount(f.C)
	must(t, err)
	// Place list is one summary scan plus a fixed place-relationship graph
	// walk for today's chains (S9-39), whatever the row count.
	if one != many {
		t.Fatalf("queries: %d for 1 Place, %d for 21", one, many)
	}
}

func TestPlacesByIDs(t *testing.T) {
	f := newFixture(t)
	york := f.place(f.Artifact, "York")
	bare := f.place(f.Artifact)
	f.place(f.Artifact, "Toronto")
	_, _, person := f.person("Ada Lovelace")
	got, err := conclusionsummaries.PlacesByIDs(f.C, [][]byte{bare, york, york, person, make([]byte, 16)})
	must(t, err)
	if len(got) != 2 {
		t.Fatalf("%+v", got)
	}
	names := map[string]bool{}
	for _, h := range got {
		if len(h.Names) == 0 {
			names[""] = true
		} else {
			names[h.Names[0]] = true
		}
	}
	if !names["York"] || !names[""] {
		t.Fatalf("%+v", got)
	}
	if got, err := conclusionsummaries.PlacesByIDs(f.C, nil); err != nil || got != nil {
		t.Fatalf("no ids: %v %+v", err, got)
	}
}

func TestListOrderIsByShownTitle(t *testing.T) {
	place := func(ref, label string, names ...string) conclusionsummaries.PlaceSummary {
		return conclusionsummaries.PlaceSummary{Entity: canonicalentities.Entity{Ref: ref, Label: label}, Names: names}
	}
	tests := []struct {
		name string
		in   []conclusionsummaries.PlaceSummary
		want []string // refs
	}{
		{
			name: "case and diacritics do not split the alphabet",
			in:   []conclusionsummaries.PlaceSummary{place("P3", "", "york"), place("P1", "", "Montréal"), place("P2", "", "Mumbai"), place("P4", "", "Évora")},
			want: []string{"P4", "P1", "P2", "P3"},
		},
		{
			name: "a working label sorts among the names",
			in:   []conclusionsummaries.PlaceSummary{place("P1", "", "York"), place("P2", "the old mill"), place("P3", "", "Kingston")},
			want: []string{"P3", "P2", "P1"},
		},
		{
			name: "a ref-only row sorts last, by ref",
			in:   []conclusionsummaries.PlaceSummary{place("PLC-B", ""), place("P1", "", "York"), place("PLC-A", "  ")},
			want: []string{"P1", "PLC-A", "PLC-B"},
		},
		{
			name: "the order the Places list showed when Swift sorted it",
			in: []conclusionsummaries.PlaceSummary{
				place("PLC-9", "", "York"), place("PLC-2", "Home"), place("PLC-B", ""),
				place("PLC-A", ""), place("PLC-1", "", "Montréal"), place("PLC-3", "", "montreal"),
			},
			want: []string{"PLC-2", "PLC-1", "PLC-3", "PLC-9", "PLC-A", "PLC-B"},
		},
		{
			name: "the same title breaks by ref",
			in:   []conclusionsummaries.PlaceSummary{place("P2", "", "Montreal"), place("P1", "", "Montréal")},
			want: []string{"P1", "P2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := append([]conclusionsummaries.PlaceSummary(nil), tt.in...)
			conclusionsummaries.SortPlaces(rows)
			var got []string
			for _, r := range rows {
				got = append(got, r.Entity.Ref)
			}
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestPersonListOrder(t *testing.T) {
	f := newFixture(t)
	must(t, nameformats.Install(f.C))
	loaded, err := nameformats.LoadFor(f.C.Graph())
	must(t, err)
	western := loaded.Profile
	part := func(typ, v string) namevalue.Part { return namevalue.Part{Type: typ, Value: v} }
	row := func(ref string, parts ...namevalue.Part) conclusionsummaries.PersonSummary {
		return conclusionsummaries.PersonSummary{Entity: canonicalentities.Entity{Ref: ref}, Name: &namevalue.Value{Parts: parts}}
	}
	tests := []struct {
		name    string
		profile nameformats.Profile
		in      []conclusionsummaries.PersonSummary
		want    []string
	}{
		{"a prefix does not file a name", western, []conclusionsummaries.PersonSummary{
			row("Smith", part("prefix", "Mrs."), part("given", "Smith")),
			row("Adams", part("given", "Zachary"), part("surname", "Adams")),
			row("Nolan", part("given", "Pat"), part("surname", "Nolan")),
		}, []string{"Adams", "Nolan", "Smith"}},
		{"within a surname by given name", western, []conclusionsummaries.PersonSummary{
			row("Zed", part("prefix", "Dr."), part("given", "Zed"), part("surname", "Smith")),
			row("Jane", part("given", "Jane"), part("surname", "Smith")),
			row("John", part("prefix", "Mr."), part("given", "John"), part("surname", "Smith")),
		}, []string{"Jane", "John", "Zed"}},
		{"stored order", nameformats.Profile{}, []conclusionsummaries.PersonSummary{
			row("Bell", part("prefix", "Dr."), part("given", "Ann"), part("surname", "Bell")),
			row("Ames", part("given", "Carl"), part("surname", "Ames")),
		}, []string{"Bell", "Ames"}},
		{"a prefix-only name keeps its key", western, []conclusionsummaries.PersonSummary{
			row("Rev", part("prefix", "Rev.")),
			row("Quinn", part("given", "Zoe"), part("surname", "Quinn")),
			row("Aaa"),
		}, []string{"Quinn", "Rev", "Aaa"}},
		{"a surname prefix stays in the key", western, []conclusionsummaries.PersonSummary{
			row("Young", part("given", "Ada"), part("surname", "Young")),
			row("Gogh", part("surname_prefix", "van"), part("surname", "Gogh"), part("given", "Vincent")),
		}, []string{"Gogh", "Young"}},
		{"names differing only by prefix tie by ref", western, []conclusionsummaries.PersonSummary{
			row("P2", part("prefix", "Dr."), part("given", "John"), part("surname", "Smith")),
			row("P1", part("prefix", "Mr."), part("given", "John"), part("surname", "Smith")),
		}, []string{"P1", "P2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := append([]conclusionsummaries.PersonSummary(nil), tt.in...)
			conclusionsummaries.SortPersons(rows, tt.profile)
			var got []string
			for _, r := range rows {
				got = append(got, r.Entity.Ref)
			}
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

// A person typed "Mrs. Smith" files with the Smiths; the listed name still
// holds its prefix.
func TestListPersonsSortIgnoresHonorifics(t *testing.T) {
	f := newFixture(t)
	must(t, nameformats.Install(f.C))
	add := func(parts ...namevalue.Part) {
		var words []string
		for i := range parts {
			parts[i].Idx = i
			words = append(words, parts[i].Value)
		}
		s := subjectstest.OfType(t, f.C, userID, f.graph, f.typeID("person"), "")
		_, _, err := observationstest.Cite(f.C, userID, f.graph, f.Artifact.ID, locator, observations.Input{
			SubjectID: s.ID, PropertyID: f.name.ID,
			Name: &namevalue.Value{Form: strings.Join(words, " "), Parts: parts},
		})
		must(t, err)
		promotetest.Save(t, f.C, userID, promote.Input{SubjectID: s.ID})
	}
	add(namevalue.Part{Type: "prefix", Value: "Mrs."}, namevalue.Part{Type: "given", Value: "Smith"})
	add(namevalue.Part{Type: "given", Value: "Zachary"}, namevalue.Part{Type: "surname", Value: "Adams"})
	add(namevalue.Part{Type: "given", Value: "Pat"}, namevalue.Part{Type: "surname", Value: "Nolan"})
	got := f.list()
	var forms []string
	for _, h := range got {
		forms = append(forms, h.Name.Form)
	}
	if want := "Zachary Adams,Pat Nolan,Mrs. Smith"; strings.Join(forms, ",") != want {
		t.Fatalf("got %v want %s", forms, want)
	}
	if n := got[2].Name; len(n.Parts) == 0 || n.Parts[0].Type != "prefix" {
		t.Fatalf("listed name lost its prefix: %+v", n)
	}
}

// SortByTitle asks for each row's title once, not once per comparison.
func TestSortByTitleKeysOnce(t *testing.T) {
	var rows []conclusionsummaries.PlaceSummary
	for i := 0; i < 50; i++ {
		rows = append(rows, conclusionsummaries.PlaceSummary{
			Entity: canonicalentities.Entity{Ref: fmt.Sprintf("P%02d", i)},
			Names:  []string{fmt.Sprintf("Town %02d", (i*17)%50)},
		})
	}
	want := append([]conclusionsummaries.PlaceSummary(nil), rows...)
	sort.SliceStable(want, func(i, j int) bool { return want[i].Names[0] < want[j].Names[0] })
	calls := 0
	conclusionsummaries.SortByTitle(rows, func(h conclusionsummaries.PlaceSummary) string {
		calls++
		return h.Names[0]
	}, func(h conclusionsummaries.PlaceSummary) string { return h.Entity.Ref })
	if calls != 50 {
		t.Fatalf("title called %d times, want 50", calls)
	}
	for i := range rows {
		if rows[i].Entity.Ref != want[i].Entity.Ref {
			t.Fatalf("row %d is %s, want %s", i, rows[i].Entity.Ref, want[i].Entity.Ref)
		}
	}
}

// TestListSummariesCarryUpdated checks the three lists return each handle's
// newest canonical_entity revision and its time, read fresh from the scope
// rows and not from the display cache.
func TestListSummariesCarryUpdated(t *testing.T) {
	f := newFixture(t)
	db, err := f.C.DB()
	must(t, err)
	// A handle with no scope row, inserted before the graph is first loaded.
	personType, err := subjecttypes.Lookup(f.C, "person", vocabkey.OriginProvenencia)
	must(t, err)
	raw, err := uuid.NewV7()
	must(t, err)
	r, err := ref.Mint(personType.RefPrefix)
	must(t, err)
	_, err = db.Exec(`INSERT INTO canonical_entities (id, subject_type_id, ref) VALUES (?, ?, ?)`,
		raw[:], personType.ID, r)
	must(t, err)

	bobSubject, bobObs, bob := f.person("Bob Smith")
	_, _, ann := f.person("Ann Lee")
	fire := f.event(observations.Input{PropertyID: f.Prop("event_name").ID, Scalar: propertymeta.TextScalar("Fire")})
	_, art := f.NewSource("Gazetteer")
	boston := f.place(art, "Boston")

	at := func(rev int64) string {
		t.Helper()
		var s string
		must(t, db.QueryRow(`SELECT created_at FROM audit_transactions WHERE revision = ?`, rev).Scan(&s))
		return s
	}
	check := func(who string, got effects.Activity) {
		t.Helper()
		if got.Revision <= 0 || got.At == "" || got.At != at(got.Revision) {
			t.Fatalf("%s updated %+v", who, got)
		}
	}
	byID := map[string]effects.Activity{}
	people := f.list()
	for _, h := range people {
		byID[string(h.Entity.ID)] = h.Updated
	}
	check("Bob", byID[string(bob)])
	check("Ann", byID[string(ann)])
	if got, ok := byID[string(raw[:])]; !ok || got != (effects.Activity{}) {
		t.Fatalf("handle without scope rows: listed %v, updated %+v", ok, got)
	}
	evs, err := conclusionsummaries.ListEvents(f.C)
	must(t, err)
	if len(evs) != 1 || string(evs[0].Entity.ID) != string(fire) {
		t.Fatalf("events %+v", evs)
	}
	check("Fire", evs[0].Updated)
	pls := f.places()
	if len(pls) != 1 || string(pls[0].Entity.ID) != string(boston) {
		t.Fatalf("places %+v", pls)
	}
	check("Boston", pls[0].Updated)

	// A note moves the scope without dropping the cached display row.
	annBefore := byID[string(ann)]
	bobBefore := byID[string(bob)]
	_, err = writes.Call(f.C, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
		return observations.Update(tx, userID, observations.Input{
			ID: bobObs[0].ID, SubjectID: bobSubject.ID, PropertyID: f.name.ID,
			Name: namevaluetest.Western("Bob Smith"), Notes: []string{"seen"},
		})
	})
	must(t, err)
	byID = map[string]effects.Activity{}
	for _, h := range f.list() {
		byID[string(h.Entity.ID)] = h.Updated
	}
	if got := byID[string(bob)]; got.Revision <= bobBefore.Revision {
		t.Fatalf("Bob did not move: %+v → %+v", bobBefore, got)
	} else {
		check("Bob", got)
	}
	if got := byID[string(ann)]; got != annBefore {
		t.Fatalf("Ann moved: %+v → %+v", annBefore, got)
	}
	if got := byID[string(raw[:])]; got != (effects.Activity{}) {
		t.Fatalf("handle without scope rows updated %+v", got)
	}

	byIDs, err := conclusionsummaries.PersonsByIDs(f.C, [][]byte{bob})
	must(t, err)
	if len(byIDs) != 1 || byIDs[0].Updated != (effects.Activity{}) {
		t.Fatalf("PersonsByIDs %+v", byIDs)
	}
}

func TestByIDs(t *testing.T) {
	f := newFixture(t)
	_, _, person := f.person("Ada Lovelace")
	event := f.event()
	h, err := conclusionsummaries.ByIDs(f.C, conclusionsummaries.Want{Persons: [][]byte{person, event}, Events: [][]byte{event, person}})
	must(t, err)
	if p, ok := h.Person(person); !ok || p.Name.Form != "Ada Lovelace" {
		t.Errorf("Person(person) = %+v, %v", p, ok)
	}
	if e, ok := h.Event(event); !ok || !bytes.Equal(e.Entity.ID, event) {
		t.Errorf("Event(event) = %+v, %v", e, ok)
	}
	if _, ok := h.Person(event); ok {
		t.Error("Person(event) found")
	}
	if _, ok := h.Place(person); ok {
		t.Error("Place(person) found")
	}
	if _, ok := h.Person(make([]byte, 16)); ok {
		t.Error("an id of no entity was found")
	}
}

// termObs cites the term of a product Property.
func (f *fixture) termObs(propertyKey, termKey string) observations.Input {
	return observations.Input{PropertyID: f.Prop(propertyKey).ID, ValueTermID: f.Term(propertyKey, termKey).ID}
}

func (f *fixture) personSummary(id []byte) conclusionsummaries.PersonSummary {
	f.T.Helper()
	got, err := conclusionsummaries.PersonsByIDs(f.C, [][]byte{id})
	must(f.T, err)
	if len(got) != 1 {
		f.T.Fatalf("summaries %d", len(got))
	}
	return got[0]
}

func TestPersonSummaryTerms(t *testing.T) {
	termKey := func(f conclusionsummaries.TermFacts) string {
		if f.Term == nil {
			return ""
		}
		return f.Term.Key
	}
	tests := []struct {
		name                      string
		build                     func(f *fixture) []byte
		gender, sex               string
		genderCount, sexCount     int
		genderTermSet, sexTermSet bool
	}{
		{
			name: "one member",
			build: func(f *fixture) []byte {
				s, _, id := f.person("Ada")
				f.Cite(f.Artifact, s, f.termObs("gender", "woman"))
				return id
			},
			gender: "woman", genderCount: 1, genderTermSet: true,
		},
		{
			name: "sex only",
			build: func(f *fixture) []byte {
				s, _, id := f.person("Ada")
				f.Cite(f.Artifact, s, f.termObs("sex_at_birth", "male"))
				return id
			},
			sex: "male", sexCount: 1, sexTermSet: true,
		},
		{
			name: "two Sources disagree",
			build: func(f *fixture) []byte {
				s, _, id := f.person("Ada")
				f.Cite(f.Artifact, s, f.termObs("gender", "woman"))
				_, art := f.NewSource("Census")
				s2 := subjectstest.OfType(f.T, f.C, userID, evidencegraphstest.SoleGraph(f.T, f.C, art.SourceID).ID, f.typeID("person"), "")
				f.Cite(art, s2, f.termObs("gender", "man"))
				promotetest.Save(f.T, f.C, userID, promote.Input{SubjectID: s2.ID, EntityID: id})
				return id
			},
			genderCount: 2, genderTermSet: true,
		},
		{
			name: "provisional",
			build: func(f *fixture) []byte {
				personType := f.typeID("person")
				handle, err := writes.Call(f.C, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
					return canonicalentities.Create(tx, userID, canonicalentities.CreateInput{SubjectTypeID: personType})
				})
				must(t, err)
				s := subjectstest.On(f.T, f.C, userID, f.graph, "person", "")
				f.Cite(f.Artifact, s, f.termObs("gender", "woman"))
				_, err = writes.Call(f.C, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
					return identityclaims.Create(tx, userID, identityclaims.CreateInput{
						SubjectID: s.ID, EntityID: handle.ID, Status: identityclaims.StatusProvisional,
					})
				})
				must(t, err)
				db, err := f.C.DB()
				must(t, err)
				_, err = autoreconciler.RecomputeReported(db, [][]byte{handle.ID})
				must(t, err)
				return handle.ID
			},
		},
		{
			name: "user Property",
			build: func(f *fixture) []byte {
				s, _, id := f.person("Ada")
				// Researchers cannot create term Properties, so the user's own
				// "gender" is a text Property: it still must not feed the summary.
				prop, err := writes.Call(f.C, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
					return properties.Create(tx, userID, "Gender", propertymeta.ValueTypeText, "", propertymeta.CardinalityMultiple)
				})
				must(f.T, err)
				personType := f.typeID("person")
				must(f.T, subjectvocab.EnsureBinding(f.C, personType, prop.ID, 99))
				f.Cite(f.Artifact, s, observations.Input{PropertyID: prop.ID, Scalar: propertymeta.TextScalar("Woman")})
				return id
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			h := f.personSummary(tt.build(f))
			if h.Gender.Count != tt.genderCount || h.SexAtBirth.Count != tt.sexCount {
				t.Fatalf("counts gender %d sex %d", h.Gender.Count, h.SexAtBirth.Count)
			}
			if (h.Gender.Term != nil) != tt.genderTermSet || (h.SexAtBirth.Term != nil) != tt.sexTermSet {
				t.Fatalf("terms gender %+v sex %+v", h.Gender.Term, h.SexAtBirth.Term)
			}
			if tt.gender != "" && termKey(h.Gender) != tt.gender || tt.sex != "" && termKey(h.SexAtBirth) != tt.sex {
				t.Fatalf("keys gender %q sex %q", termKey(h.Gender), termKey(h.SexAtBirth))
			}
		})
	}
}

func TestPersonSummaryTermLabels(t *testing.T) {
	t.Run("labels", func(t *testing.T) {
		f := newFixture(t)
		s, _, id := f.person("Ada")
		f.Cite(f.Artifact, s, f.termObs("gender", "woman"))
		h := f.personSummary(id)
		if h.Gender.Term == nil || h.Gender.Term.Label != "Woman" || h.Gender.Term.Origin != vocabkey.OriginProvenencia {
			t.Fatalf("term %+v", h.Gender.Term)
		}
		if h.Gender.PropertyLabel != "Gender" || h.SexAtBirth.PropertyLabel != "Sex at birth" {
			t.Fatalf("property labels %q %q", h.Gender.PropertyLabel, h.SexAtBirth.PropertyLabel)
		}
	})
	t.Run("term relabel", func(t *testing.T) {
		f := newFixture(t)
		s, _, id := f.person("Ada")
		genderID := f.Prop("gender").ID
		term, err := writes.Call(f.C, writes.Op{Action: "create_property_term", UserID: userID}, func(tx *database.Tx) (propertyterms.Term, []rowchange.Change, error) {
			return propertyterms.Create(tx, userID, genderID, "Agender", "")
		})
		must(t, err)
		f.Cite(f.Artifact, s, observations.Input{PropertyID: genderID, ValueTermID: term.ID})
		if h := f.personSummary(id); h.Gender.Term == nil || h.Gender.Term.Label != "Agender" {
			t.Fatalf("before %+v", h.Gender.Term)
		}
		_, err = writes.Call(f.C, writes.Op{Action: "update_property_term", UserID: userID}, func(tx *database.Tx) (propertyterms.Term, []rowchange.Change, error) {
			return propertyterms.Update(tx, userID, term.ID, "No gender", "")
		})
		must(t, err)
		h := f.personSummary(id)
		if h.Gender.Term == nil || h.Gender.Term.Label != "No gender" || h.Gender.Term.Origin != "user" {
			t.Fatalf("after %+v", h.Gender.Term)
		}
		must(t, f.C.Graph().Verify())
	})
	t.Run("Property relabel", func(t *testing.T) {
		f := newFixture(t)
		s, _, id := f.person("Ada")
		f.Cite(f.Artifact, s, f.termObs("gender", "woman"))
		if h := f.personSummary(id); h.Gender.PropertyLabel != "Gender" {
			t.Fatalf("before %q", h.Gender.PropertyLabel)
		}
		p := f.Prop("gender")
		_, err := writes.Call(f.C, writes.Op{Action: "update_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
			return properties.Update(tx, userID, p.ID, "Gender identity", p.ValueType, p.Description, p.Cardinality)
		})
		must(t, err)
		if h := f.personSummary(id); h.Gender.PropertyLabel != "Gender identity" {
			t.Fatalf("after %q", h.Gender.PropertyLabel)
		}
		must(t, f.C.Graph().Verify())
	})
}

func TestSortByBirth(t *testing.T) {
	person := func(ref string, year int) conclusionsummaries.PersonSummary {
		h := conclusionsummaries.PersonSummary{Entity: canonicalentities.Entity{Ref: ref}}
		if year != 0 {
			h.Birth.Date = datevaluetest.YMD(year, 0, 0)
		}
		return h
	}
	personWith := func(ref string, d *datevalue.Value) conclusionsummaries.PersonSummary {
		h := person(ref, 0)
		h.Birth.Date = d
		return h
	}
	cases := []struct {
		name string
		in   []conclusionsummaries.PersonSummary
		want []string
	}{
		{"dated before undated", []conclusionsummaries.PersonSummary{person("PER-A", 0), person("PER-B", 1820)}, []string{"PER-B", "PER-A"}},
		{"earliest first", []conclusionsummaries.PersonSummary{person("PER-A", 1820), person("PER-B", 1815)}, []string{"PER-B", "PER-A"}},
		{"undated by ref", []conclusionsummaries.PersonSummary{person("PER-B", 0), person("PER-A", 0)}, []string{"PER-A", "PER-B"}},
		{"equal dates by ref", []conclusionsummaries.PersonSummary{person("PER-B", 1820), person("PER-A", 1820)}, []string{"PER-A", "PER-B"}},
		{"open start sits at its known end", []conclusionsummaries.PersonSummary{person("PER-D", 0), person("PER-C", 1820), personWith("PER-B", datevaluetest.Before(1817, 0, 0)), person("PER-A", 1815)}, []string{"PER-A", "PER-B", "PER-C", "PER-D"}},
		{"refs differing only in case keep input order", []conclusionsummaries.PersonSummary{person("PER-a", 1820), person("PER-A", 1820)}, []string{"PER-a", "PER-A"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conclusionsummaries.SortByBirth(tc.in, func(h conclusionsummaries.PersonSummary) conclusionsummaries.PersonSummary { return h })
			var got []string
			for _, h := range tc.in {
				got = append(got, h.Entity.Ref)
			}
			if fmt.Sprint(got) != fmt.Sprint(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestEventSpan(t *testing.T) {
	noYear := &datevalue.Value{Kind: datevalue.KindPoint, Calendar: "gregorian"}
	noStart := &datevalue.Value{Kind: datevalue.KindRange, Calendar: "gregorian", EndYear: datevaluetest.Part(1849)}
	bothOpen := &datevalue.Value{Kind: datevalue.KindRange, Calendar: "gregorian"}
	march := datevaluetest.YMD(1817, 3, 12)
	year := func(from, to float64, openStart, openEnd bool) conclusionsummaries.TimelineSpan {
		return conclusionsummaries.TimelineSpan{From: from, To: to, OpenStart: openStart, OpenEnd: openEnd}
	}
	cases := []struct {
		name   string
		in     conclusionsummaries.EventSummary
		want   conclusionsummaries.TimelineSpan
		wantOK bool
	}{
		{"day", conclusionsummaries.EventSummary{Date: march}, year(1817+70.0/365, 1817+71.0/365, false, false), true},
		{"year", conclusionsummaries.EventSummary{Date: datevaluetest.YMD(1835, 0, 0)}, year(1835, 1836, false, false), true},
		{"range", conclusionsummaries.EventSummary{Date: datevaluetest.Between(1844, 0, 0, 1849, 0, 0)}, year(1844, 1850, false, false), true},
		{"after", conclusionsummaries.EventSummary{Date: datevaluetest.After(1870, 0, 0)}, year(1870, 1871, false, true), true},
		{"about is not widened", conclusionsummaries.EventSummary{Date: datevaluetest.About(1846, 0, 0)}, year(1846, 1847, false, false), true},
		{"before", conclusionsummaries.EventSummary{Date: datevaluetest.Before(1880, 0, 0)}, year(1880, 1881, true, false), true},
		{"range with no start year", conclusionsummaries.EventSummary{Date: noStart}, year(1849, 1850, true, false), true},
		{"start date only", conclusionsummaries.EventSummary{StartDate: datevaluetest.YMD(1860, 0, 0)}, year(1860, 1861, false, false), true},
		{"end date only", conclusionsummaries.EventSummary{EndDate: datevaluetest.YMD(1870, 0, 0)}, conclusionsummaries.TimelineSpan{}, false},
		{"no date", conclusionsummaries.EventSummary{}, conclusionsummaries.TimelineSpan{}, false},
		{"date with no year", conclusionsummaries.EventSummary{Date: noYear}, conclusionsummaries.TimelineSpan{}, false},
		{"both ends open", conclusionsummaries.EventSummary{Date: bothOpen}, conclusionsummaries.TimelineSpan{}, false},
		{"date with no year, dated start", conclusionsummaries.EventSummary{Date: noYear, StartDate: datevaluetest.YMD(1860, 0, 0)}, year(1860, 1861, false, false), true},
		{"date wins over start", conclusionsummaries.EventSummary{Date: datevaluetest.YMD(1850, 0, 0), StartDate: datevaluetest.YMD(1860, 0, 0)}, year(1850, 1851, false, false), true},
		{"year above the window", conclusionsummaries.EventSummary{Date: datevaluetest.YMD(1<<31-1, 0, 0)}, conclusionsummaries.TimelineSpan{}, false},
		{"year below the window", conclusionsummaries.EventSummary{Date: datevaluetest.YMD(-1<<31, 0, 0)}, conclusionsummaries.TimelineSpan{}, false},
		{"range from -2^31 to 2^31", conclusionsummaries.EventSummary{Date: datevaluetest.Between(-1<<31, 0, 0, 1<<31-1, 0, 0)}, conclusionsummaries.TimelineSpan{}, false},
		{"range with one end past the window", conclusionsummaries.EventSummary{Date: datevaluetest.Between(1844, 0, 0, 10000, 0, 0)}, conclusionsummaries.TimelineSpan{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := conclusionsummaries.EventSpanAndAge(tc.in, conclusionsummaries.PlacedBirth{}, false)
			got, ok := p.Span, p.Dated
			if ok != tc.wantOK || got != tc.want {
				t.Fatalf("EventSpan = %+v, %v; want %+v, %v", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestEventSpanDays(t *testing.T) {
	noYear := &datevalue.Value{Kind: datevalue.KindPoint, Calendar: "gregorian"}
	days := func(d *datevalue.Value) valuecompare.Window {
		w, ok := valuecompare.DateWindow(*d)
		if !ok {
			t.Fatalf("DateWindow(%+v) has no days", d)
		}
		return w
	}
	cases := []struct {
		name      string
		in        conclusionsummaries.EventSummary
		want      valuecompare.Window
		wantDated bool
	}{
		{"about is not widened", conclusionsummaries.EventSummary{Date: datevaluetest.About(2011, 0, 0)}, days(datevaluetest.YMD(2011, 0, 0)), true},
		{"after is open at its end", conclusionsummaries.EventSummary{Date: datevaluetest.After(2011, 3, 29)}, days(datevaluetest.After(2011, 3, 29)), true},
		{"before is open at its start", conclusionsummaries.EventSummary{Date: datevaluetest.Before(1880, 0, 0)}, days(datevaluetest.Before(1880, 0, 0)), true},
		{"start date only", conclusionsummaries.EventSummary{StartDate: datevaluetest.YMD(1860, 0, 0)}, days(datevaluetest.YMD(1860, 0, 0)), true},
		{"date with no year, dated start", conclusionsummaries.EventSummary{Date: noYear, StartDate: datevaluetest.YMD(1860, 3, 4)}, days(datevaluetest.YMD(1860, 3, 4)), true},
		{"no date", conclusionsummaries.EventSummary{}, valuecompare.Window{}, false},
	}
	same := func(a, b *int) bool { return (a == nil) == (b == nil) && (a == nil || *a == *b) }
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := conclusionsummaries.EventSpanAndAge(tc.in, conclusionsummaries.PlacedBirth{}, false)
			got, dated := p.Days, p.Dated
			if dated != tc.wantDated || !same(got.Lo, tc.want.Lo) || !same(got.Hi, tc.want.Hi) {
				t.Fatalf("days = %+v, %v; want %+v, %v", got, dated, tc.want, tc.wantDated)
			}
		})
	}
}

func TestEventOrderAtKnownEnd(t *testing.T) {
	f := newFixture(t)
	date := f.Prop("date")
	dated := func(d *datevalue.Value) []byte {
		return f.event(observations.Input{PropertyID: date.ID, Date: d})
	}
	undated := f.event()
	y1881 := dated(datevaluetest.YMD(1881, 0, 0))
	bef1880 := dated(datevaluetest.Before(1880, 0, 0))
	y1880 := dated(datevaluetest.YMD(1880, 0, 0))
	y1879 := dated(datevaluetest.YMD(1879, 0, 0))
	// On the same day, before sorts first and after last, whatever the refs.
	aft2011 := dated(datevaluetest.After(2011, 3, 29))
	day2011 := dated(datevaluetest.YMD(2011, 3, 29))
	bef2011 := dated(datevaluetest.Before(2011, 3, 29))

	want := [][]byte{y1879, bef1880, y1880, y1881, bef2011, day2011, aft2011, undated}

	check := func(name string, got []conclusionsummaries.EventSummary) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("%s: %d rows, want %d", name, len(got), len(want))
		}
		for i := range want {
			if !bytes.Equal(got[i].Entity.ID, want[i]) {
				t.Fatalf("%s: row %d is %s, want %s", name, i, got[i].Entity.Ref, refOf(t, f, want[i]))
			}
		}
	}
	check("ListEvents", f.events())
	byIDs, err := conclusionsummaries.EventsByIDs(f.C, [][]byte{undated, aft2011, y1881, bef1880, day2011, y1880, bef2011, y1879})
	must(t, err)
	check("EventsByIDs", byIDs)
}

func refOf(t *testing.T, f *fixture, id []byte) string {
	t.Helper()
	got, err := conclusionsummaries.EventsByIDs(f.C, [][]byte{id})
	must(t, err)
	if len(got) != 1 {
		t.Fatalf("no event %x", id)
	}
	return got[0].Entity.Ref
}

func TestTermLabeler(t *testing.T) {
	c := graphcachetest.OpenCatalog(t)
	label, err := conclusionsummaries.TermLabeler(c)
	if err != nil {
		t.Fatal(err)
	}
	var id []byte
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT id FROM property_terms WHERE key = 'witness' AND origin = 'provenencia'`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	got, ok := label(id, "witness")
	if !ok || got.Key != "witness" || got.Label != "Witness" || got.Origin != vocabkey.OriginProvenencia || !bytes.Equal(got.ID, id) {
		t.Fatalf("label(witness) = %+v, %v", got, ok)
	}
	if _, ok := label(bytes.Repeat([]byte{0x7e}, 16), "witness"); ok {
		t.Fatal("an unknown id was labelled")
	}
}

func TestEventAge(t *testing.T) {
	born := datevaluetest.YMD(1817, 3, 12)
	age := func(lo, hi int) conclusionsummaries.AgeRange { return conclusionsummaries.AgeRange{Lo: lo, Hi: hi} }
	ev := func(d *datevalue.Value) conclusionsummaries.EventSummary {
		return conclusionsummaries.EventSummary{Date: d}
	}
	cases := []struct {
		name   string
		in     conclusionsummaries.EventSummary
		birth  *datevalue.Value
		want   conclusionsummaries.AgeRange
		wantOK bool
	}{
		{"the birth day", ev(born), born, age(0, 0), true},
		{"a day", ev(datevaluetest.YMD(1844, 6, 4)), born, age(27, 27), true},
		{"a year", ev(datevaluetest.YMD(1835, 0, 0)), born, age(17, 18), true},
		{"a range", ev(datevaluetest.Between(1844, 0, 0, 1849, 0, 0)), born, age(26, 32), true},
		{"start date only", conclusionsummaries.EventSummary{StartDate: datevaluetest.YMD(1844, 6, 4)}, born, age(27, 27), true},
		{"before: at most, the year included", ev(datevaluetest.Before(1880, 0, 0)), born, conclusionsummaries.AgeRange{Hi: 63, OpenLo: true}, true},
		{"after: at least, the year included", ev(datevaluetest.After(1850, 0, 0)), born, conclusionsummaries.AgeRange{Lo: 32, OpenHi: true}, true},
		{"a range with a start only", ev(datevaluetest.Between(1850, 0, 0, 0, 0, 0)), born, conclusionsummaries.AgeRange{Lo: 32, OpenHi: true}, true},
		{"a range with an end only", ev(datevaluetest.Between(0, 0, 0, 1860, 0, 0)), born, conclusionsummaries.AgeRange{Hi: 43, OpenLo: true}, true},
		{"a start date before a year", conclusionsummaries.EventSummary{StartDate: datevaluetest.Before(1850, 0, 0)}, born, conclusionsummaries.AgeRange{Hi: 33, OpenLo: true}, true},
		{"before a day before the birth", ev(datevaluetest.Before(1810, 0, 0)), born, conclusionsummaries.AgeRange{}, false},
		{"after with a birth after: no bound", ev(datevaluetest.After(1850, 0, 0)), datevaluetest.After(1817, 0, 0), conclusionsummaries.AgeRange{}, false},
		{"a birth after a year", ev(datevaluetest.YMD(1844, 6, 4)), datevaluetest.After(1817, 0, 0), conclusionsummaries.AgeRange{Hi: 27, OpenLo: true}, true},
		{"an about birth", ev(datevaluetest.YMD(1844, 6, 4)), datevaluetest.About(1817, 0, 0), age(26, 27), true},
		{"a birth before a year", ev(datevaluetest.YMD(1844, 6, 4)), datevaluetest.Before(1817, 0, 0), conclusionsummaries.AgeRange{Lo: 26, OpenHi: true}, true},
		{"no birth", ev(datevaluetest.YMD(1844, 6, 4)), nil, conclusionsummaries.AgeRange{}, false},
		{"no event date", conclusionsummaries.EventSummary{}, born, conclusionsummaries.AgeRange{}, false},
		{"event before the birth", ev(datevaluetest.YMD(1800, 0, 0)), born, conclusionsummaries.AgeRange{}, false},
		{"year above the window", ev(datevaluetest.YMD(1<<31-1, 0, 0)), born, conclusionsummaries.AgeRange{}, false},
		{"birth year above the window", ev(datevaluetest.YMD(1844, 6, 4)), datevaluetest.YMD(1<<31-1, 0, 0), conclusionsummaries.AgeRange{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			placed, hasBirth := conclusionsummaries.PlaceBirth(tc.birth)
			p := conclusionsummaries.EventSpanAndAge(tc.in, placed, hasBirth)
			got, ok := p.Age, p.Aged
			if ok != tc.wantOK || got != tc.want {
				t.Errorf("EventAge = %+v, %v; want %+v, %v", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

// TestEventParticipants checks that an event row keeps every participant by id
// with the distinct roles their links carry, in link order, and that the
// subjects stay the title's name summaries.
func TestEventParticipants(t *testing.T) {
	type part struct {
		who   string
		roles string
	}
	tests := []struct {
		name         string
		build        func(k *graphcachetest.Canon, ned, pat []byte) (event []byte)
		want         []part
		wantSubjects []string
	}{
		{"a subject and a witness", func(k *graphcachetest.Canon, ned, pat []byte) []byte {
			ev := k.Event(vocabkey.EventBurial)
			k.Join(ned, ev, string(vocabkey.RoleSubject))
			k.Join(pat, ev, string(vocabkey.RoleWitness))
			return ev
		}, []part{{"ned", "subject"}, {"pat", "witness"}}, []string{"ned"}},
		{"one person with two roles is one participant", func(k *graphcachetest.Canon, ned, pat []byte) []byte {
			ev := k.Event(vocabkey.EventBurial)
			k.Join(ned, ev, string(vocabkey.RoleSubject))
			k.Join(ned, ev, string(vocabkey.RoleWitness))
			return ev
		}, []part{{"ned", "subject+witness"}}, []string{"ned"}},
		{"a participant with no accepted member is listed", func(k *graphcachetest.Canon, ned, pat []byte) []byte {
			ev := k.Event(vocabkey.EventBurial)
			k.Join(ned, ev, string(vocabkey.RoleSubject))
			k.Join(pat, ev, string(vocabkey.RoleWitness))
			k.Provisional(pat)
			return ev
		}, []part{{"ned", "subject"}, {"pat", "witness"}}, []string{"ned"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := graphcachetest.OpenCatalog(t)
			k := graphcachetest.NewCanon(t, c)
			ned, pat := k.Person(), k.Person()
			ev := tc.build(k, ned, pat)
			k.Commit(ned, pat, ev)
			rows, err := conclusionsummaries.EventsByIDs(c, [][]byte{ev})
			must(t, err)
			if len(rows) != 1 {
				t.Fatalf("%d rows", len(rows))
			}
			who := map[string]string{string(ned): "ned", string(pat): "pat"}
			var got []part
			for _, p := range rows[0].Participants {
				var keys []string
				for _, r := range p.Roles {
					keys = append(keys, r.Key)
				}
				got = append(got, part{who[string(p.ID)], strings.Join(keys, "+")})
			}
			if !slices.Equal(got, tc.want) {
				t.Fatalf("participants %v, want %v", got, tc.want)
			}
			var subjects []string
			for _, s := range rows[0].Subjects {
				subjects = append(subjects, who[string(s.Entity.ID)])
			}
			if !slices.Equal(subjects, tc.wantSubjects) {
				t.Fatalf("subjects %v, want %v", subjects, tc.wantSubjects)
			}
		})
	}
}
