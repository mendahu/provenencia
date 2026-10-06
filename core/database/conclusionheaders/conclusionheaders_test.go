package conclusionheaders_test

import (
	"fmt"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
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
	src, err := sources.Create(c, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	must(t, err)
	art, err := artifacts.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
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
	s, err := subjects.Create(f.c, userID, subjects.CreateInput{SourceID: f.source.ID, SubjectTypeID: f.typeID("person")}, nil)
	must(f.t, err)
	var obs []observations.Observation
	if len(forms) > 0 {
		var in []observations.Input
		for _, form := range forms {
			in = append(in, observations.Input{SubjectID: s.ID, PropertyID: f.name.ID, Name: namevaluestest.Western(form)})
		}
		res, err := citations.CreateWithObservations(f.c, userID, citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: locator}, in)
		must(f.t, err)
		obs = res.Observations
	}
	p, err := promote.Save(f.c, userID, promote.Input{SubjectID: s.ID})
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
	labelled, err := canonicalentities.Create(f.c, userID, canonicalentities.CreateInput{SubjectTypeID: f.typeID("person"), Label: "Mother of James"})
	must(t, err)
	bare, err := canonicalentities.Create(f.c, userID, canonicalentities.CreateInput{SubjectTypeID: f.typeID("person")})
	must(t, err)
	_, err = canonicalentities.Create(f.c, userID, canonicalentities.CreateInput{SubjectTypeID: f.typeID("place"), Label: "York"})
	must(t, err)

	got := f.list()
	if len(got) != 4 {
		t.Fatalf("listed %d Persons, want 4 (no Place)", len(got))
	}
	// Named first by sort key (james < mary), then unnamed by ref. James,
	// Jim and james reconcile into one name (S9-13b): Jim is a different
	// given name, so it joins rather than being outvoted.
	if got[0].Name == nil || got[0].Name.Form != "James Jim Robins" || got[0].NameValueCount != 1 {
		t.Fatalf("first %+v", got[0])
	}
	if got[1].Name == nil || got[1].Name.Form != "Mary Smith" || got[1].NameValueCount != 1 {
		t.Fatalf("second %+v", got[1])
	}
	unnamed := map[string]conclusionheaders.PersonHeader{got[2].Entity.Ref: got[2], got[3].Entity.Ref: got[3]}
	if h := unnamed[labelled.Ref]; h.Name != nil || h.NameValueCount != 0 || h.Entity.Label != "Mother of James" {
		t.Fatalf("labelled %+v", h)
	}
	if h := unnamed[bare.Ref]; h.Name != nil || h.Entity.Label != "" {
		t.Fatalf("bare %+v", h)
	}
	if got[2].Entity.Ref > got[3].Entity.Ref {
		t.Fatalf("unnamed not ordered by ref: %s, %s", got[2].Entity.Ref, got[3].Entity.Ref)
	}

	t.Run("a name edit reaches the header", func(t *testing.T) {
		// Jim → James merges the two values.
		_, err := observations.Update(f.c, userID, observations.Input{
			ID: jamesObs[1].ID, SubjectID: jamesObs[1].SubjectID, PropertyID: f.name.ID,
			Name: namevaluestest.Western("James Robins"),
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
	if one != 1 || many != one {
		t.Fatalf("queries: %d for 1 Person, %d for 51", one, many)
	}
}

func TestPersonsByIDs(t *testing.T) {
	f := newFixture(t)
	_, _, mary := f.person("Mary Smith")
	_, _, james := f.person("James Robins")
	f.person("Ada Lovelace")
	place, err := canonicalentities.Create(f.c, userID, canonicalentities.CreateInput{SubjectTypeID: f.typeID("place")})
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
	s, err := subjects.Create(f.c, userID, subjects.CreateInput{SourceID: f.source.ID, SubjectTypeID: f.typeID("event")}, nil)
	must(f.t, err)
	if len(in) > 0 {
		for i := range in {
			in[i].SubjectID = s.ID
		}
		_, err = citations.CreateWithObservations(f.c, userID, citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: locator}, in)
		must(f.t, err)
	}
	p, err := promote.Save(f.c, userID, promote.Input{SubjectID: s.ID})
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
	if one != 1 || many != one {
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
