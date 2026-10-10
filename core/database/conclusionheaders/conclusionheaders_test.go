package conclusionheaders_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcecredibility"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

const locator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

type fixture struct {
	t        *testing.T
	c        *database.Catalog
	source   sources.Source
	artifact artifacts.Artifact
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
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	must(t, err)
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	must(t, err)
	art, err := runArtifactCreate(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	must(t, err)
	name, err := properties.Lookup(c, "name", properties.OriginProvenencia)
	must(t, err)
	return &fixture{t: t, c: c, source: src, artifact: art, name: name}
}

func (f *fixture) typeID(key string) []byte {
	st, err := subjecttypes.Lookup(f.c, key, subjecttypes.OriginProvenencia)
	must(f.t, err)
	return st.ID
}

// person promotes a new person Subject named by forms (one Observation each).
func (f *fixture) person(forms ...string) (subjects.Subject, []observations.Observation, []byte) {
	f.t.Helper()
	personType := f.typeID("person")
	s, err := writes.Call(f.c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{SourceID: f.source.ID, SubjectTypeID: personType}, nil)
	})
	must(f.t, err)
	var obs []observations.Observation
	if len(forms) > 0 {
		var in []observations.Input
		for _, form := range forms {
			in = append(in, observations.Input{SubjectID: s.ID, PropertyID: f.name.ID, Name: namevaluestest.Western(form)})
		}
		res, err := writes.Call(f.c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: locator}, in)
		})
		must(f.t, err)
		obs = res.Observations
	}
	p, err := writes.Call(f.c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, promote.Input{SubjectID: s.ID})
	})
	must(f.t, err)
	return s, obs, p.Entity.ID
}

func (f *fixture) list() []conclusionheaders.PersonHeader {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	h, err := conclusionheaders.ListPersons(db)
	must(f.t, err)
	return h
}

func TestListPersons(t *testing.T) {
	f := newFixture(t)
	if got := f.list(); len(got) != 0 {
		t.Fatalf("empty project listed %d", len(got))
	}

	f.person("Mary Smith")
	_, jamesObs, _ := f.person("James Robins", "Jim Robins", "james robins")
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
	// By the title the row shows (james < mary < "Mother of James"), then
	// the ref-only row. James, Jim and james reconcile into one name
	// (S9-13b): Jim is a different given name, so it joins rather than
	// being outvoted.
	if got[0].Name == nil || got[0].Name.Form != "James Jim Robins" || got[0].NameValueCount != 1 {
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
		// Jim → James merges the two values.
		_, err := writes.Call(f.c, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
			return observations.Update(tx, userID, observations.Input{
				ID: jamesObs[1].ID, SubjectID: jamesObs[1].SubjectID, PropertyID: f.name.ID,
				Name: namevaluestest.Western("James Robins"),
			})
		})
		must(t, err)
		if h := f.list()[0]; h.Name.Form != "James Robins" || h.NameValueCount != 1 {
			t.Fatalf("after edit %+v", h)
		}
	})
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
				return []observations.Input{{PropertyID: f.name.ID, Name: namevaluestest.Western("Ghost Name")}}
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
					if h.EventName != "" || h.EventType != nil || h.Date != nil {
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
			s := f.bare(tt.kind)
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
	f.person("Ada Lovelace")
	db, err := f.c.DB()
	must(t, err)
	one, err := conclusionheaders.ListPersonsQueryCount(db)
	must(t, err)
	for i := 0; i < 50; i++ {
		f.person(fmt.Sprintf("Person %d", i), fmt.Sprintf("Alias %d", i))
	}
	many, err := conclusionheaders.ListPersonsQueryCount(db)
	must(t, err)
	// The name row is one query; birth and death are one more, for the whole
	// list. The count stays put as the list grows.
	if one < 1 || many != one {
		t.Fatalf("queries: %d for 1 Person, %d for 51", one, many)
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
	db, err := f.c.DB()
	must(t, err)

	got, err := conclusionheaders.PersonsByIDs(db, [][]byte{mary, james, mary, place.ID, make([]byte, 16)})
	must(t, err)
	if len(got) != 2 || got[0].Name.Form != "James Robins" || got[1].Name.Form != "Mary Smith" {
		t.Fatalf("want James, Mary in list order: %+v", got)
	}
	if got, err := conclusionheaders.PersonsByIDs(db, nil); err != nil || got != nil {
		t.Fatalf("no ids: %v %+v", err, got)
	}
}

func (f *fixture) prop(key string) properties.Property {
	f.t.Helper()
	p, err := properties.Lookup(f.c, key, properties.OriginProvenencia)
	must(f.t, err)
	return p
}

func (f *fixture) term(propertyKey, termKey string) propertyterms.Term {
	f.t.Helper()
	p := f.prop(propertyKey)
	term, err := propertyterms.Lookup(f.c, p.ID, termKey, propertyterms.OriginProvenencia)
	must(f.t, err)
	return term
}

func pointYear(year int) *datevalues.Value {
	y := year
	return &datevalues.Value{Kind: datevalues.KindPoint, Calendar: "gregorian", StartYear: &y}
}

// event promotes a new event Subject with the given Observations.
func (f *fixture) event(in ...observations.Input) []byte {
	f.t.Helper()
	eventType := f.typeID("event")
	s, err := writes.Call(f.c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{SourceID: f.source.ID, SubjectTypeID: eventType}, nil)
	})
	must(f.t, err)
	if len(in) > 0 {
		for i := range in {
			in[i].SubjectID = s.ID
		}
		_, err = writes.Call(f.c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: locator}, in)
		})
		must(f.t, err)
	}
	p, err := writes.Call(f.c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, promote.Input{SubjectID: s.ID})
	})
	must(f.t, err)
	return p.Entity.ID
}

func (f *fixture) events() []conclusionheaders.EventHeader {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	h, err := conclusionheaders.ListEvents(db)
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
	if born.EventType == nil || born.EventType.Key != "birth" || born.EventType.Label != "Birth" || born.EventTypeCount != 1 {
		t.Fatalf("type %+v", born.EventType)
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
	db, err := f.c.DB()
	must(t, err)
	one, err := conclusionheaders.ListEventsQueryCount(db)
	must(t, err)
	date := f.prop("date")
	for i := 0; i < 50; i++ {
		f.event(observations.Input{PropertyID: date.ID, Date: pointYear(1800 + i)})
	}
	many, err := conclusionheaders.ListEventsQueryCount(db)
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
	db, err := f.c.DB()
	must(t, err)
	got, err := conclusionheaders.EventsByIDs(db, [][]byte{bare, fire, fire, person, make([]byte, 16)})
	must(t, err)
	if len(got) != 2 {
		t.Fatalf("%+v", got)
	}
	names := map[string]bool{got[0].EventName: true, got[1].EventName: true}
	if !names["Fire"] || !names[""] {
		t.Fatalf("%+v", got)
	}
	if got, err := conclusionheaders.EventsByIDs(db, nil); err != nil || got != nil {
		t.Fatalf("no ids: %v %+v", err, got)
	}
}

func (f *fixture) book(title string) (sources.Source, artifacts.Artifact) {
	f.t.Helper()
	src, _, err := writes.Run(f.c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: f.source.SourceTypeID, Title: title})
	})
	must(f.t, err)
	art, err := runArtifactCreate(f.c, userID, artifacts.CreateInput{SourceID: src.ID, Label: title})
	must(f.t, err)
	return src, art
}

func (f *fixture) lowTrust(src sources.Source) {
	f.t.Helper()
	g, err := sourcecredibilitygrades.Lookup(f.c, "low_trust", sourcecredibilitygrades.OriginProvenencia)
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
	s, err := writes.Call(f.c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{SourceID: art.SourceID, SubjectTypeID: placeType}, nil)
	})
	must(f.t, err)
	f.citeToponyms(art, s.ID, names...)
	p, err := writes.Call(f.c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, promote.Input{SubjectID: s.ID})
	})
	must(f.t, err)
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
	_, err := writes.Call(f.c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, userID, citations.CreateInput{ArtifactID: art.ID, LocatorJSON: locator}, in)
	})
	must(f.t, err)
}

func (f *fixture) joinPlace(art artifacts.Artifact, entityID []byte, names ...string) {
	f.t.Helper()
	placeType := f.typeID("place")
	s, err := writes.Call(f.c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{SourceID: art.SourceID, SubjectTypeID: placeType}, nil)
	})
	must(f.t, err)
	f.citeToponyms(art, s.ID, names...)
	_, err = writes.Call(f.c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, promote.Input{SubjectID: s.ID, EntityID: entityID})
	})
	must(f.t, err)
}

func (f *fixture) places() []conclusionheaders.PlaceHeader {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	h, err := conclusionheaders.ListPlaces(db)
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

	db, err := f.c.DB()
	must(t, err)
	detail, err := conclusiondetails.ForEntity(db, york)
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
	db, err := f.c.DB()
	must(t, err)
	one, err := conclusionheaders.ListPlacesQueryCount(db)
	must(t, err)
	for i := 0; i < 20; i++ {
		f.place(f.artifact, fmt.Sprintf("Place %d", i))
	}
	many, err := conclusionheaders.ListPlacesQueryCount(db)
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
	db, err := f.c.DB()
	must(t, err)
	got, err := conclusionheaders.PlacesByIDs(db, [][]byte{bare, york, york, person, make([]byte, 16)})
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
	if got, err := conclusionheaders.PlacesByIDs(db, nil); err != nil || got != nil {
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

func runArtifactCreate(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error) {
	a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, in)
		})
	return a, err
}
