package conclusionheaders_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
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
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

const locator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

type fixture struct {
	t        *testing.T
	c        *database.Catalog
	source   sources.Source
	artifact artifacts.Artifact
	graph    []byte // the Source's one Evidence graph
	name     properties.Property
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	must(t, err)
	t.Cleanup(func() { _ = c.Close() })
	r, err := ref.Mint(ref.PrefixUser)
	must(t, err)
	must(t, users.Upsert(c, userID, "Tester", r))
	must(t, subjectvocab.Install(c))
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	must(t, err)
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	must(t, err)
	art, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	must(t, err)
	name, err := properties.Lookup(c, "name", vocabkey.OriginProvenencia)
	must(t, err)
	return &fixture{t: t, c: c, source: src, artifact: art, graph: evidencegraphstest.SoleGraph(t, c, src.ID).ID, name: name}
}

func (f *fixture) typeID(key string) []byte {
	st, err := subjecttypes.Lookup(f.c, key, vocabkey.OriginProvenencia)
	must(f.t, err)
	return st.ID
}

// person promotes a new person Subject named by forms (one Observation each).
func (f *fixture) person(forms ...string) (subjects.Subject, []observations.Observation, []byte) {
	f.t.Helper()
	personType := f.typeID("person")
	s := subjectstest.OfType(f.t, f.c, userID, f.graph, personType, "")
	var obs []observations.Observation
	if len(forms) > 0 {
		var in []observations.Input
		for _, form := range forms {
			in = append(in, observations.Input{SubjectID: s.ID, PropertyID: f.name.ID, Name: namevaluetest.Western(form)})
		}
		res, _, err := observationstest.Cite(f.c, userID, f.graph, f.artifact.ID, locator, in...)
		must(f.t, err)
		obs = res.Observations
	}
	p := promotetest.Save(f.t, f.c, userID, promote.Input{SubjectID: s.ID})
	return s, obs, p.Entity.ID
}

func (f *fixture) list() []conclusionheaders.PersonHeader {
	f.t.Helper()
	h, err := conclusionheaders.ListPersons(f.c)
	must(f.t, err)
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
	labelled, err := writes.Call(f.c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
		return canonicalentities.Create(tx, userID, p6a1)
	})
	must(t, err)
	p6a2 := canonicalentities.CreateInput{SubjectTypeID: f.typeID("person")}
	bare, err := writes.Call(f.c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
		return canonicalentities.Create(tx, userID, p6a2)
	})
	must(t, err)
	p6a3 := canonicalentities.CreateInput{SubjectTypeID: f.typeID("place"), Label: "York"}
	_, err = writes.Call(f.c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
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

	t.Run("a name edit reaches the header", func(t *testing.T) {
		// Kenneth → James merges the two values.
		_, err := writes.Call(f.c, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
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
	must(t, nameformats.Install(f.c))
	f.person("Aaron Zimmerman")
	f.person("Zachary Adams")
	formOnly := &namevalue.Value{Form: "Aaron Aardvark"}
	personType := f.typeID("person")
	s := subjectstest.OfType(t, f.c, userID, f.graph, personType, "")
	_, _, err := observationstest.Cite(f.c, userID, f.graph, f.artifact.ID, locator, []observations.Input{{
		SubjectID: s.ID, PropertyID: f.name.ID, Name: formOnly,
	}}...)
	must(t, err)
	promotetest.Save(t, f.c, userID, promote.Input{SubjectID: s.ID})

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

// A header shows what the detail page displays: a kept value. When every
// value is held back (here, only a provisional member speaks), rank 1 is not
// kept, so the row falls back to the label or ref.
func TestHeadersShowKeptValuesOnly(t *testing.T) {
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
					{PropertyID: f.prop("event_name").ID, ValueText: "Phantom Fair", HasText: true},
					{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID},
					{PropertyID: f.prop("date").ID, Date: pointYear(1801)},
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
			handle, err := writes.Call(f.c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
				return canonicalentities.Create(tx, userID, p6a4)
			})
			must(t, err)
			s := subjectstest.On(f.t, f.c, userID, f.graph, tt.kind, "")
			f.cite(s, tt.cite(f)...)
			_, err = writes.Call(f.c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
				return identityclaims.Create(tx, userID, identityclaims.CreateInput{
					SubjectID: s.ID, EntityID: handle.ID, Status: identityclaims.StatusProvisional,
				})
			})
			must(t, err)
			db, err := f.c.DB()
			must(t, err)
			must(t, autoreconciler.RecomputeTx(db, [][]byte{handle.ID}))
			tt.check(t, f, handle.ID)
		})
	}
}

func TestListPersonsQueryCountIsConstant(t *testing.T) {
	f := newFixture(t)
	gender := func(s subjects.Subject) {
		f.cite(s, observations.Input{PropertyID: f.prop("gender").ID, ValueTermID: f.term("gender", "woman").ID})
	}
	s, _, _ := f.person("Ada Lovelace")
	gender(s)
	f.c.Graph().Drop()
	one, err := conclusionheaders.ListPersonsQueryCount(f.c)
	must(t, err)
	for i := 0; i < 50; i++ {
		s, _, _ := f.person(fmt.Sprintf("Person %d", i), fmt.Sprintf("Alias %d", i))
		gender(s)
	}
	f.c.Graph().Drop()
	many, err := conclusionheaders.ListPersonsQueryCount(f.c)
	must(t, err)
	// The name row is one query; birth and death are one more, and the
	// likeness lead one more, each for the whole list. The count stays put as the list grows.
	if one < 1 || many != one {
		t.Fatalf("queries: %d for 1 Person, %d for 51", one, many)
	}
}

// subjectAndHandle creates and promotes a Subject of kind; returns subject and handle.
func (f *fixture) subjectAndHandle(kind string) (subjects.Subject, []byte) {
	f.t.Helper()
	s := subjectstest.OfType(f.t, f.c, userID, f.graph, f.typeID(kind), "")
	p := promotetest.Save(f.t, f.c, userID, promote.Input{SubjectID: s.ID})
	return s, p.Entity.ID
}

func pictureLeads[H any](rows []H, id, pic func(H) []byte) map[string][]byte {
	out := map[string][]byte{}
	for _, r := range rows {
		out[string(id(r))] = pic(r)
	}
	return out
}

type headerPictureCase struct {
	name  string
	key   string
	setup func(f *fixture) (subjects.Subject, []byte)
	list  func(f *fixture) (map[string][]byte, error)
	byIDs func(f *fixture, ids [][]byte) (map[string][]byte, error)
}

func headerPictureCases() []headerPictureCase {
	personID := func(r conclusionheaders.PersonHeader) []byte { return r.Entity.ID }
	personPic := func(r conclusionheaders.PersonHeader) []byte { return r.PictureImageFileID }
	eventID := func(r conclusionheaders.EventHeader) []byte { return r.Entity.ID }
	eventPic := func(r conclusionheaders.EventHeader) []byte { return r.PictureImageFileID }
	placeID := func(r conclusionheaders.PlaceHeader) []byte { return r.Entity.ID }
	placePic := func(r conclusionheaders.PlaceHeader) []byte { return r.PictureImageFileID }
	return []headerPictureCase{
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
				rows, err := conclusionheaders.PersonsByIDs(f.c, ids)
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
				rows, err := conclusionheaders.EventsByIDs(f.c, ids)
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
				rows, err := conclusionheaders.PlacesByIDs(f.c, ids)
				return pictureLeads(rows, placeID, placePic), err
			},
		},
	}
}

func TestHeaderLeadsWithEarliestPicture(t *testing.T) {
	for _, tc := range headerPictureCases() {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			s, h := tc.setup(f)
			var other []byte
			if tc.name == "person" {
				_, _, other = f.person("Bob Lee")
			}
			imageOf := func(o observations.Observation) []byte { return observationstest.ImageOf(t, f.c, o.ID) }
			leads, err := tc.list(f)
			must(t, err)
			if got := leads[string(h)]; got != nil {
				t.Fatalf("lead before any picture: %x", got)
			}
			a := observationstest.SaveDepiction(t, f.c, userID, f.source.ID, s.ID, tc.key, 100)
			leads, err = tc.list(f)
			must(t, err)
			if got := leads[string(h)]; !bytes.Equal(got, imageOf(a)) {
				t.Fatalf("lead after A: %x, want %x", got, imageOf(a))
			}
			observationstest.SaveDepiction(t, f.c, userID, f.source.ID, s.ID, tc.key, 200)
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
			must(t, f.c.Graph().Verify())
		})
	}
}

func TestPersonHeaderSkipsNegativeLikeness(t *testing.T) {
	f := newFixture(t)
	s, _, h := f.person("Ann Lee")
	observationstest.SaveNegativeLikeness(t, f.c, userID, f.source.ID, s.ID, 100)
	b := observationstest.SaveLikeness(t, f.c, userID, f.source.ID, s.ID, 200)
	observationstest.SaveLikeness(t, f.c, userID, f.source.ID, s.ID, 100)
	var got []byte
	for _, r := range f.list() {
		if bytes.Equal(r.Entity.ID, h) {
			got = r.PictureImageFileID
		}
	}
	if want := observationstest.ImageOf(t, f.c, b.ID); !bytes.Equal(got, want) {
		t.Fatalf("lead %x, want %x", got, want)
	}
	must(t, f.c.Graph().Verify())
}

func TestHeaderLeadsWithChosenPicture(t *testing.T) {
	for _, tc := range headerPictureCases() {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			s, h := tc.setup(f)
			a := observationstest.SaveDepiction(t, f.c, userID, f.source.ID, s.ID, tc.key, 100)
			b := observationstest.SaveDepiction(t, f.c, userID, f.source.ID, s.ID, tc.key, 200)
			imageOf := func(o observations.Observation) []byte { return observationstest.ImageOf(t, f.c, o.ID) }
			choose := func(obs []byte) {
				t.Helper()
				_, err := writes.Call(f.c, writes.Op{Action: "set_preferred_likeness", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
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
				must(t, f.c.Graph().Verify())
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
	place, err := writes.Call(f.c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
		return canonicalentities.Create(tx, userID, p6a9)
	})
	must(t, err)

	got, err := conclusionheaders.PersonsByIDs(f.c, [][]byte{mary, james, mary, place.ID, make([]byte, 16)})
	must(t, err)
	if len(got) != 2 || got[0].Name.Form != "James Robins" || got[1].Name.Form != "Mary Smith" {
		t.Fatalf("want James, Mary in list order: %+v", got)
	}
	if got, err := conclusionheaders.PersonsByIDs(f.c, nil); err != nil || got != nil {
		t.Fatalf("no ids: %v %+v", err, got)
	}
}

func (f *fixture) prop(key string) properties.Property {
	f.t.Helper()
	p, err := properties.Lookup(f.c, key, vocabkey.OriginProvenencia)
	must(f.t, err)
	return p
}

func (f *fixture) term(propertyKey, termKey string) propertyterms.Term {
	f.t.Helper()
	p := f.prop(propertyKey)
	term, err := propertyterms.Lookup(f.c, p.ID, termKey, vocabkey.OriginProvenencia)
	must(f.t, err)
	return term
}

func pointYear(year int) *datevalue.Value {
	y := year
	return &datevalue.Value{Kind: datevalue.KindPoint, Calendar: "gregorian", StartYear: &y}
}

// event promotes a new event Subject with the given Observations.
func (f *fixture) event(in ...observations.Input) []byte {
	f.t.Helper()
	eventType := f.typeID("event")
	s := subjectstest.OfType(f.t, f.c, userID, f.graph, eventType, "")
	if len(in) > 0 {
		for i := range in {
			in[i].SubjectID = s.ID
		}
		_, _, err := observationstest.Cite(f.c, userID, f.graph, f.artifact.ID, locator, in...)
		must(f.t, err)
	}
	p := promotetest.Save(f.t, f.c, userID, promote.Input{SubjectID: s.ID})
	return p.Entity.ID
}

func (f *fixture) events() []conclusionheaders.EventHeader {
	f.t.Helper()
	h, err := conclusionheaders.ListEvents(f.c)
	must(f.t, err)
	return h
}

func TestListEvents(t *testing.T) {
	f := newFixture(t)
	if got := f.events(); len(got) != 0 {
		t.Fatalf("empty project listed %d", len(got))
	}
	name := f.prop("event_name")
	date := f.prop("date")
	start := f.prop("start_date")
	end := f.prop("end_date")
	eventType := f.prop("event_type")
	birth := f.term("event_type", "birth")

	fire := f.event(
		observations.Input{PropertyID: name.ID, ValueText: " The Great Fire ", HasText: true},
		observations.Input{PropertyID: date.ID, Date: pointYear(1849)},
	)
	birthID := f.event(
		observations.Input{PropertyID: eventType.ID, ValueTermID: birth.ID},
		observations.Input{PropertyID: date.ID, Date: pointYear(1985)},
		observations.Input{PropertyID: start.ID, Date: pointYear(1900)},
	)
	span := f.event(
		observations.Input{PropertyID: start.ID, Date: pointYear(1900)},
		observations.Input{PropertyID: end.ID, Date: pointYear(1910)},
	)
	undated := f.event()
	f.person("Ada Lovelace")

	got := f.events()
	if len(got) != 4 {
		t.Fatalf("listed %d Events, want 4 (no Person)", len(got))
	}
	byID := map[string]conclusionheaders.EventHeader{}
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
	f.c.Graph().Drop()
	one, err := conclusionheaders.ListEventsQueryCount(f.c)
	must(t, err)
	date := f.prop("date")
	for i := 0; i < 50; i++ {
		f.event(observations.Input{PropertyID: date.ID, Date: pointYear(1800 + i)})
	}
	f.c.Graph().Drop()
	many, err := conclusionheaders.ListEventsQueryCount(f.c)
	must(t, err)
	// The event row is one query; subjects and places are one each, for the
	// whole list. The count stays put as the list grows.
	if one < 1 || many != one {
		t.Fatalf("queries: %d for 1 Event, %d for 51", one, many)
	}
}

func TestEventsByIDs(t *testing.T) {
	f := newFixture(t)
	name := f.prop("event_name")
	fire := f.event(observations.Input{PropertyID: name.ID, ValueText: "Fire", HasText: true})
	bare := f.event()
	f.event(observations.Input{PropertyID: name.ID, ValueText: "Other", HasText: true})
	_, _, person := f.person("Ada Lovelace")
	got, err := conclusionheaders.EventsByIDs(f.c, [][]byte{bare, fire, fire, person, make([]byte, 16)})
	must(t, err)
	if len(got) != 2 {
		t.Fatalf("%+v", got)
	}
	names := map[string]bool{got[0].EventName: true, got[1].EventName: true}
	if !names["Fire"] || !names[""] {
		t.Fatalf("%+v", got)
	}
	if got, err := conclusionheaders.EventsByIDs(f.c, nil); err != nil || got != nil {
		t.Fatalf("no ids: %v %+v", err, got)
	}
}

func (f *fixture) book(title string) (sources.Source, artifacts.Artifact) {
	f.t.Helper()
	src, _, err := writes.Run(f.c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: f.source.SourceTypeID, Title: title})
	})
	must(f.t, err)
	art, err := artifactstest.Create(f.c, userID, artifacts.CreateInput{SourceID: src.ID, Label: title})
	must(f.t, err)
	return src, art
}

func (f *fixture) lowTrust(src sources.Source) {
	f.t.Helper()
	g, err := sourcecredibilitygrades.Lookup(f.c, "low_trust", vocabkey.OriginProvenencia)
	must(f.t, err)
	p6in11 := sourcecredibility.UpsertInput{
		SourceID: src.ID, CredibilityGradeID: g.ID,
	}
	p6act10 := "create_source_credibility_assessment"

	// place promotes a new Place whose toponyms are cited, in order, on art.
	if _, p6look12 := sourcecredibility.GetBySource(f.c, p6in11.SourceID); p6look12 == nil {
		p6act10 = "update_source_credibility_assessment"
	}
	_, err = writes.Call(f.c, writes.Op{Action: p6act10, UserID: userID}, func(tx *database.Tx) (sourcecredibility.Assessment, []rowchange.Change, error) {
		return sourcecredibility.Upsert(tx, userID, p6in11)
	})

	must(f.t, err)
}

func (f *fixture) place(art artifacts.Artifact, names ...string) []byte {
	f.t.Helper()
	placeType := f.typeID("place")
	graph := evidencegraphstest.SoleGraph(f.t, f.c, art.SourceID).ID
	s := subjectstest.OfType(f.t, f.c, userID, graph, placeType, "")
	f.citeToponyms(art, s.ID, names...)
	p := promotetest.Save(f.t, f.c, userID, promote.Input{SubjectID: s.ID})
	return p.Entity.ID
}

func (f *fixture) citeToponyms(art artifacts.Artifact, subjectID []byte, names ...string) {
	f.t.Helper()
	if len(names) == 0 {
		return
	}
	toponym := f.prop("toponym")
	in := make([]observations.Input, len(names))
	for i, name := range names {
		in[i] = observations.Input{SubjectID: subjectID, PropertyID: toponym.ID, ValueText: name, HasText: true}
	}
	graph := evidencegraphstest.SoleGraph(f.t, f.c, art.SourceID).ID
	_, _, err := observationstest.Cite(f.c, userID, graph, art.ID, locator, in...)
	must(f.t, err)
}

func (f *fixture) joinPlace(art artifacts.Artifact, entityID []byte, names ...string) {
	f.t.Helper()
	placeType := f.typeID("place")
	graph := evidencegraphstest.SoleGraph(f.t, f.c, art.SourceID).ID
	s := subjectstest.OfType(f.t, f.c, userID, graph, placeType, "")
	f.citeToponyms(art, s.ID, names...)
	promotetest.Save(f.t, f.c, userID, promote.Input{SubjectID: s.ID, EntityID: entityID})
}

func (f *fixture) places() []conclusionheaders.PlaceHeader {
	f.t.Helper()
	h, err := conclusionheaders.ListPlaces(f.c)
	must(f.t, err)
	return h
}

func TestPlaceKeepsEveryDistinctToponym(t *testing.T) {
	f := newFixture(t)
	must(t, sourcecredibilitygrades.Install(f.c))
	census, censusArt := f.book("Census")
	gazette, gazetteArt := f.book("Gazette")
	f.lowTrust(gazette)

	york := f.place(f.artifact, "York", "york")
	f.joinPlace(censusArt, york, "Toronto")
	f.joinPlace(gazetteArt, york, "Muddy York")

	got := f.places()
	if len(got) != 1 || len(got[0].Names) != 2 || got[0].Names[0] != "York" || got[0].Names[1] != "Toronto" {
		t.Fatalf("names %+v", got)
	}
	if got[0].StartDate != nil || got[0].EndDate != nil || got[0].Kind != "" || len(got[0].Parents) != 0 {
		t.Fatalf("no period or parents without evidence: %+v", got[0])
	}

	detail, err := conclusiondetails.ForEntity(f.c, york)
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
	york := f.place(f.artifact, "York")
	montreal := f.place(f.artifact, "montreal")
	bareA := f.place(f.artifact)
	bareB := f.place(f.artifact)
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
	f.place(f.artifact, "York")
	f.c.Graph().Drop()
	one, err := conclusionheaders.ListPlacesQueryCount(f.c)
	must(t, err)
	for i := 0; i < 20; i++ {
		f.place(f.artifact, fmt.Sprintf("Place %d", i))
	}
	f.c.Graph().Drop()
	many, err := conclusionheaders.ListPlacesQueryCount(f.c)
	must(t, err)
	// Place list is one header scan plus a fixed place-relationship graph
	// walk for today's chains (S9-39), whatever the row count.
	if one != many {
		t.Fatalf("queries: %d for 1 Place, %d for 21", one, many)
	}
}

func TestPlacesByIDs(t *testing.T) {
	f := newFixture(t)
	york := f.place(f.artifact, "York")
	bare := f.place(f.artifact)
	f.place(f.artifact, "Toronto")
	_, _, person := f.person("Ada Lovelace")
	got, err := conclusionheaders.PlacesByIDs(f.c, [][]byte{bare, york, york, person, make([]byte, 16)})
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
	if got, err := conclusionheaders.PlacesByIDs(f.c, nil); err != nil || got != nil {
		t.Fatalf("no ids: %v %+v", err, got)
	}
}

func TestListOrderIsByShownTitle(t *testing.T) {
	place := func(ref, label string, names ...string) conclusionheaders.PlaceHeader {
		return conclusionheaders.PlaceHeader{Entity: canonicalentities.Entity{Ref: ref, Label: label}, Names: names}
	}
	tests := []struct {
		name string
		in   []conclusionheaders.PlaceHeader
		want []string // refs
	}{
		{
			name: "case and diacritics do not split the alphabet",
			in:   []conclusionheaders.PlaceHeader{place("P3", "", "york"), place("P1", "", "Montréal"), place("P2", "", "Mumbai"), place("P4", "", "Évora")},
			want: []string{"P4", "P1", "P2", "P3"},
		},
		{
			name: "a working label sorts among the names",
			in:   []conclusionheaders.PlaceHeader{place("P1", "", "York"), place("P2", "the old mill"), place("P3", "", "Kingston")},
			want: []string{"P3", "P2", "P1"},
		},
		{
			name: "a ref-only row sorts last, by ref",
			in:   []conclusionheaders.PlaceHeader{place("PLC-B", ""), place("P1", "", "York"), place("PLC-A", "  ")},
			want: []string{"P1", "PLC-A", "PLC-B"},
		},
		{
			name: "the order the Places list showed when Swift sorted it",
			in: []conclusionheaders.PlaceHeader{
				place("PLC-9", "", "York"), place("PLC-2", "Home"), place("PLC-B", ""),
				place("PLC-A", ""), place("PLC-1", "", "Montréal"), place("PLC-3", "", "montreal"),
			},
			want: []string{"PLC-2", "PLC-1", "PLC-3", "PLC-9", "PLC-A", "PLC-B"},
		},
		{
			name: "the same title breaks by ref",
			in:   []conclusionheaders.PlaceHeader{place("P2", "", "Montreal"), place("P1", "", "Montréal")},
			want: []string{"P1", "P2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := append([]conclusionheaders.PlaceHeader(nil), tt.in...)
			conclusionheaders.SortPlaces(rows)
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

// TestListHeadersCarryUpdated checks the three lists return each handle's
// newest canonical_entity revision and its time, read fresh from the scope
// rows and not from the display cache.
func TestListHeadersCarryUpdated(t *testing.T) {
	f := newFixture(t)
	db, err := f.c.DB()
	must(t, err)
	// A handle with no scope row, inserted before the graph is first loaded.
	personType, err := subjecttypes.Lookup(f.c, "person", vocabkey.OriginProvenencia)
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
	fire := f.event(observations.Input{PropertyID: f.prop("event_name").ID, ValueText: "Fire", HasText: true})
	_, art := f.book("Gazetteer")
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
	evs, err := conclusionheaders.ListEvents(f.c)
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
	_, err = writes.Call(f.c, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
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

	byIDs, err := conclusionheaders.PersonsByIDs(f.c, [][]byte{bob})
	must(t, err)
	if len(byIDs) != 1 || byIDs[0].Updated != (effects.Activity{}) {
		t.Fatalf("PersonsByIDs %+v", byIDs)
	}
}

func TestByIDs(t *testing.T) {
	f := newFixture(t)
	_, _, person := f.person("Ada Lovelace")
	event := f.event()
	h, err := conclusionheaders.ByIDs(f.c, [][]byte{person, event}, [][]byte{event, person}, nil)
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
	return observations.Input{PropertyID: f.prop(propertyKey).ID, ValueTermID: f.term(propertyKey, termKey).ID}
}

func (f *fixture) personHeader(id []byte) conclusionheaders.PersonHeader {
	f.t.Helper()
	got, err := conclusionheaders.PersonsByIDs(f.c, [][]byte{id})
	must(f.t, err)
	if len(got) != 1 {
		f.t.Fatalf("headers %d", len(got))
	}
	return got[0]
}

func TestPersonHeaderTerms(t *testing.T) {
	termKey := func(f conclusionheaders.TermFacts) string {
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
				f.cite(s, f.termObs("gender", "woman"))
				return id
			},
			gender: "woman", genderCount: 1, genderTermSet: true,
		},
		{
			name: "sex only",
			build: func(f *fixture) []byte {
				s, _, id := f.person("Ada")
				f.cite(s, f.termObs("sex_at_birth", "male"))
				return id
			},
			sex: "male", sexCount: 1, sexTermSet: true,
		},
		{
			name: "two Sources disagree",
			build: func(f *fixture) []byte {
				s, _, id := f.person("Ada")
				f.cite(s, f.termObs("gender", "woman"))
				_, art := f.book("Census")
				s2 := subjectstest.OfType(f.t, f.c, userID, evidencegraphstest.SoleGraph(f.t, f.c, art.SourceID).ID, f.typeID("person"), "")
				f.citeOn(art, s2, f.termObs("gender", "man"))
				promotetest.Save(f.t, f.c, userID, promote.Input{SubjectID: s2.ID, EntityID: id})
				return id
			},
			genderCount: 2, genderTermSet: true,
		},
		{
			name: "provisional",
			build: func(f *fixture) []byte {
				personType := f.typeID("person")
				handle, err := writes.Call(f.c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
					return canonicalentities.Create(tx, userID, canonicalentities.CreateInput{SubjectTypeID: personType})
				})
				must(t, err)
				s := subjectstest.On(f.t, f.c, userID, f.graph, "person", "")
				f.cite(s, f.termObs("gender", "woman"))
				_, err = writes.Call(f.c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
					return identityclaims.Create(tx, userID, identityclaims.CreateInput{
						SubjectID: s.ID, EntityID: handle.ID, Status: identityclaims.StatusProvisional,
					})
				})
				must(t, err)
				db, err := f.c.DB()
				must(t, err)
				must(t, autoreconciler.RecomputeTx(db, [][]byte{handle.ID}))
				return handle.ID
			},
		},
		{
			name: "user Property",
			build: func(f *fixture) []byte {
				s, _, id := f.person("Ada")
				// Researchers cannot create term Properties, so the user's own
				// "gender" is a text Property: it still must not feed the header.
				prop, err := writes.Call(f.c, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
					return properties.Create(tx, userID, "Gender", propertymeta.ValueTypeText, "", propertymeta.CardinalityMultiple)
				})
				must(f.t, err)
				personType := f.typeID("person")
				must(f.t, subjectvocab.EnsureBinding(f.c, personType, prop.ID, 99))
				f.cite(s, observations.Input{PropertyID: prop.ID, ValueText: "Woman", HasText: true})
				return id
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			h := f.personHeader(tt.build(f))
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

func TestPersonHeaderTermLabels(t *testing.T) {
	t.Run("labels", func(t *testing.T) {
		f := newFixture(t)
		s, _, id := f.person("Ada")
		f.cite(s, f.termObs("gender", "woman"))
		h := f.personHeader(id)
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
		genderID := f.prop("gender").ID
		term, err := writes.Call(f.c, writes.Op{Action: "create_property_term", UserID: userID}, func(tx *database.Tx) (propertyterms.Term, []rowchange.Change, error) {
			return propertyterms.Create(tx, userID, genderID, "Agender", "")
		})
		must(t, err)
		f.cite(s, observations.Input{PropertyID: genderID, ValueTermID: term.ID})
		if h := f.personHeader(id); h.Gender.Term == nil || h.Gender.Term.Label != "Agender" {
			t.Fatalf("before %+v", h.Gender.Term)
		}
		_, err = writes.Call(f.c, writes.Op{Action: "update_property_term", UserID: userID}, func(tx *database.Tx) (propertyterms.Term, []rowchange.Change, error) {
			return propertyterms.Update(tx, userID, term.ID, "No gender", "")
		})
		must(t, err)
		h := f.personHeader(id)
		if h.Gender.Term == nil || h.Gender.Term.Label != "No gender" || h.Gender.Term.Origin != "user" {
			t.Fatalf("after %+v", h.Gender.Term)
		}
		must(t, f.c.Graph().Verify())
	})
	t.Run("Property relabel", func(t *testing.T) {
		f := newFixture(t)
		s, _, id := f.person("Ada")
		f.cite(s, f.termObs("gender", "woman"))
		if h := f.personHeader(id); h.Gender.PropertyLabel != "Gender" {
			t.Fatalf("before %q", h.Gender.PropertyLabel)
		}
		p := f.prop("gender")
		_, err := writes.Call(f.c, writes.Op{Action: "update_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
			return properties.Update(tx, userID, p.ID, "Gender identity", p.ValueType, p.Description, p.Cardinality)
		})
		must(t, err)
		if h := f.personHeader(id); h.Gender.PropertyLabel != "Gender identity" {
			t.Fatalf("after %q", h.Gender.PropertyLabel)
		}
		must(t, f.c.Graph().Verify())
	})
}
