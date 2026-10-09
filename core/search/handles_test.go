package search

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/writes"
)

const handleLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

type handleFixture struct {
	t        *testing.T
	c        *database.Catalog
	user     []byte
	source   []byte
	artifact []byte
}

func newHandleFixture(t *testing.T) *handleFixture {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	user := seedUser(t, c)
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	typeID := seedType(t, c, "Book", "")
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: user}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, user, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	art, err := runArtifactCreate(c, user, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}
	return &handleFixture{t: t, c: c, user: user, source: src.ID, artifact: art.ID}
}

func (f *handleFixture) prop(key string) properties.Property {
	f.t.Helper()
	p, err := properties.Lookup(f.c, key, properties.OriginProvenencia)
	if err != nil {
		f.t.Fatal(err)
	}
	return p
}

// subject creates a Subject of kind with the given Observations.
func (f *handleFixture) subject(kind string, in func(s subjects.Subject) []observations.Input) subjects.Subject {
	f.t.Helper()
	st, err := subjecttypes.Lookup(f.c, kind, subjecttypes.OriginProvenencia)
	if err != nil {
		f.t.Fatal(err)
	}
	s, err := writes.Call(f.c, writes.Op{Action: "create_subject", UserID: f.user}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, f.user, subjects.CreateInput{SourceID: f.source, SubjectTypeID: st.ID}, nil)
	})
	if err != nil {
		f.t.Fatal(err)
	}
	if in != nil {
		obs := in(s)
		if _, err := writes.Call(f.c, writes.Op{Action: "create_citation_with_observations", UserID: f.user}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, f.user, citations.CreateInput{ArtifactID: f.artifact, LocatorJSON: handleLocator}, obs)
		}); err != nil {
			f.t.Fatal(err)
		}
	}
	return s
}

func (f *handleFixture) person(forms ...string) subjects.Subject {
	return f.subject("person", func(s subjects.Subject) []observations.Input {
		var in []observations.Input
		for _, form := range forms {
			in = append(in, observations.Input{SubjectID: s.ID, PropertyID: f.prop("name").ID, Name: namevaluestest.Western(form)})
		}
		return in
	})
}

func (f *handleFixture) promote(s subjects.Subject, onto []byte) promote.Result {
	f.t.Helper()
	res, err := promote.Save(f.c, f.user, promote.Input{SubjectID: s.ID, EntityID: onto})
	if err != nil {
		f.t.Fatal(err)
	}
	return res
}

func (f *handleFixture) search(text string, kinds ...string) []Hit {
	f.t.Helper()
	hits, err := DefaultEngine().Search(context.Background(), f.c, Query{Text: text, Kinds: kinds})
	if err != nil {
		f.t.Fatal(err)
	}
	return hits
}

func refsOf(hits []Hit) string {
	var out []string
	for _, h := range hits {
		out = append(out, h.Kind+":"+h.Ref)
	}
	sort.Strings(out)
	return fmt.Sprint(out)
}

func TestHandleSearch(t *testing.T) {
	f := newHandleFixture(t)
	james := f.promote(f.person("James Robins", "Jim Robins"), nil)
	f.promote(f.person("Mary Smith"), nil)

	t.Run("the omnibar default includes handles", func(t *testing.T) {
		hits := f.search("Robins")
		if refsOf(hits) != fmt.Sprint([]string{"person:" + james.Entity.Ref}) {
			t.Fatalf("got %s", refsOf(hits))
		}
	})

	t.Run("by rank-1 name, with location and member count", func(t *testing.T) {
		hits := f.search("James Robins", KindPerson)
		if len(hits) != 1 || hits[0].Ref != james.Entity.Ref || hits[0].Title != "James Jim Robins" {
			t.Fatalf("got %+v", hits)
		}
		loc := hits[0].Location
		if loc.Section != SectionPersons || loc.EntityID != uuidString(james.Entity.ID) || loc.Ref != james.Entity.Ref {
			t.Fatalf("location %+v", loc)
		}
		if hits[0].MemberCount != 1 {
			t.Fatalf("members %d", hits[0].MemberCount)
		}
	})

	t.Run("by another name cluster", func(t *testing.T) {
		if hits := f.search("Jim", KindPerson); refsOf(hits) != fmt.Sprint([]string{"person:" + james.Entity.Ref}) {
			t.Fatalf("got %s", refsOf(hits))
		}
	})

	t.Run("by ref", func(t *testing.T) {
		hits := f.search(james.Entity.Ref, KindPerson)
		if len(hits) == 0 || hits[0].Ref != james.Entity.Ref || hits[0].MatchReason != "ref" {
			t.Fatalf("got %+v", hits)
		}
	})

	t.Run("a join raises the member count and adds its names", func(t *testing.T) {
		f.promote(f.person("J. Robbins"), james.Entity.ID)
		hits := f.search("Robbins", KindPerson)
		if len(hits) != 1 || hits[0].Ref != james.Entity.Ref || hits[0].MemberCount != 2 {
			t.Fatalf("got %+v", hits)
		}
	})

	t.Run("unknown kinds return nothing", func(t *testing.T) {
		if hits := f.search("Robins", "nope"); len(hits) != 0 {
			t.Fatalf("got %s", refsOf(hits))
		}
	})
}

// A name outvoted by the other records isn't displayed, but it stays a
// cached value, so search still finds the handle by it (S9-13).
func TestHandleSearchFindsOutvotedNames(t *testing.T) {
	f := newHandleFixture(t)
	thomas := f.promote(f.person("Thomas Lee", "thomas lee", "Tom Lee"), nil)
	if hits := f.search("Tom", KindPerson); refsOf(hits) != fmt.Sprint([]string{"person:" + thomas.Entity.Ref}) {
		t.Fatalf("got %s", refsOf(hits))
	}
}

func TestHandleSearchFollowsEdits(t *testing.T) {
	f := newHandleFixture(t)
	s := f.subject("person", func(s subjects.Subject) []observations.Input {
		return []observations.Input{{SubjectID: s.ID, PropertyID: f.prop("name").ID, Name: namevaluestest.Western("Ada Byron")}}
	})
	ada := f.promote(s, nil)
	obs, err := observations.ListBySubject(f.c, s.ID)
	if err != nil || len(obs) != 1 {
		t.Fatalf("%v %d", err, len(obs))
	}
	nameID := f.prop("name").ID
	if _, err := writes.Call(f.c, writes.Op{Action: "update_observation", UserID: f.user}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
		return observations.Update(tx, f.user, observations.Input{
			ID: obs[0].ID, SubjectID: s.ID, PropertyID: nameID, Name: namevaluestest.Western("Ada Lovelace"),
		})
	}); err != nil {
		t.Fatal(err)
	}
	if hits := f.search("Lovelace", KindPerson); len(hits) != 1 || hits[0].Ref != ada.Entity.Ref {
		t.Fatalf("new name: %s", refsOf(hits))
	}
	if hits := f.search("Byron", KindPerson); len(hits) != 0 {
		t.Fatalf("old name still found: %s", refsOf(hits))
	}
}

func TestHeaderSearchFindsTheListTitle(t *testing.T) {
	f := newHandleFixture(t)
	person := f.person("James Robins", "Jim Robins")
	birth, err := propertyterms.Lookup(f.c, f.prop("event_type").ID, "birth", propertyterms.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	y := 1817
	event := f.subject("event", func(s subjects.Subject) []observations.Input {
		return []observations.Input{
			{SubjectID: s.ID, PropertyID: f.prop("event_type").ID, ValueTermID: birth.ID},
			{SubjectID: s.ID, PropertyID: f.prop("date").ID, Date: &datevalues.Value{Kind: datevalues.KindPoint, Calendar: "gregorian", StartYear: &y}},
		}
	})
	place := f.subject("place", func(s subjects.Subject) []observations.Input {
		return []observations.Input{{SubjectID: s.ID, PropertyID: f.prop("toponym").ID, ValueText: "York", HasText: true}}
	})
	for i, s := range []subjects.Subject{person, event, place} {
		if _, err := writes.Call(f.c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
			return subjectpositions.Set(tx, s.ID, int64(i), 0)
		}); err != nil {
			t.Fatal(err)
		}
	}
	role, err := propertyterms.Lookup(f.c, f.prop("role").ID, "subject", propertyterms.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	personProp := f.prop("person").ID
	eventProp := f.prop("event").ID
	roleProp := f.prop("role").ID
	if _, err := writes.Call(f.c, writes.Op{Action: "create_cited_bridge", UserID: f.user}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, f.user, connect.CreateInput{
			SourceID: f.source, FromSubjectID: person.ID, ToSubjectID: event.ID, BridgeTypeKey: "participation",
			Citation: citations.CreateInput{ArtifactID: f.artifact, LocatorJSON: handleLocator},
			Observations: []observations.Input{
				{PropertyID: personProp, ValueSubjectID: person.ID},
				{PropertyID: eventProp, ValueSubjectID: event.ID},
				{PropertyID: roleProp, ValueTermID: role.ID},
			},
		})
	}); err != nil {
		t.Fatal(err)
	}
	placeProp := f.prop("place").ID
	if _, err := writes.Call(f.c, writes.Op{Action: "create_cited_bridge", UserID: f.user}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, f.user, connect.CreateInput{
			SourceID: f.source, FromSubjectID: event.ID, ToSubjectID: place.ID, BridgeTypeKey: "location",
			Citation: citations.CreateInput{ArtifactID: f.artifact, LocatorJSON: handleLocator},
			Observations: []observations.Input{
				{PropertyID: eventProp, ValueSubjectID: event.ID},
				{PropertyID: placeProp, ValueSubjectID: place.ID},
			},
		})
	}); err != nil {
		t.Fatal(err)
	}
	promoted := f.promote(person, nil)
	f.promote(event, nil)
	f.promote(place, nil)
	f.fileBridges()

	if hits := f.search("Jim Robins", KindPerson); len(hits) != 1 || hits[0].Ref != promoted.Entity.Ref {
		t.Fatalf("alternate name: %s", refsOf(hits))
	}
	if hits := f.search(promoted.Entity.Ref, KindPerson); len(hits) != 1 || hits[0].MatchReason != "ref" {
		t.Fatalf("ref: %+v", hits)
	}
	if hits := f.search("James", KindEvent); len(hits) != 1 || hits[0].Title != "Birth of James Jim Robins" {
		t.Fatalf("event title: %+v", hits)
	}
	if hits := f.search("York", KindPlace); len(hits) != 1 {
		t.Fatalf("place: %s", refsOf(hits))
	}

	obs, err := observations.ListBySubject(f.c, person.ID)
	if err != nil {
		t.Fatal(err)
	}
	var jamesObs observations.Listed
	for _, o := range obs {
		if o.ValueNameForm == "James Robins" || o.ValueText == "James Robins" {
			jamesObs = o
			break
		}
	}
	if len(jamesObs.ID) != 16 {
		t.Fatalf("james observation missing: %+v", obs)
	}
	nameID := f.prop("name").ID
	if _, err := writes.Call(f.c, writes.Op{Action: "update_observation", UserID: f.user}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
		return observations.Update(tx, f.user, observations.Input{
			ID: jamesObs.ID, SubjectID: person.ID, PropertyID: nameID,
			Name: namevaluestest.Western("John Robins"),
		})
	}); err != nil {
		t.Fatal(err)
	}
	if hits := f.search("John", KindEvent); len(hits) != 1 || hits[0].Title != "Birth of John Jim Robins" {
		t.Fatalf("after edit: %+v", hits)
	}
	if hits := f.search("James", KindEvent); len(hits) != 0 {
		t.Fatalf("old name still titles the event: %+v", hits)
	}
}

func (f *handleFixture) fileBridges() {
	f.t.Helper()
	db, err := f.c.DB()
	if err != nil {
		f.t.Fatal(err)
	}
	var rev int64
	if err := db.QueryRow(`SELECT COALESCE(MAX(revision), 0) FROM audit_transactions`).Scan(&rev); err != nil {
		f.t.Fatal(err)
	}
	if _, err := promote.SaveBatch(f.c, f.user, promote.Batch{SourceID: f.source, SeenRevision: rev}); err != nil {
		f.t.Fatal(err)
	}
}

func TestAlternateToponymIsMatchContext(t *testing.T) {
	f := newHandleFixture(t)
	place := f.promote(f.subject("place", func(s subjects.Subject) []observations.Input {
		return []observations.Input{
			{SubjectID: s.ID, PropertyID: f.prop("toponym").ID, ValueText: "Montréal", HasText: true},
			{SubjectID: s.ID, PropertyID: f.prop("toponym").ID, ValueText: "Montreal", HasText: true},
		}
	}), nil)
	hits := f.search("Montreal", KindPlace)
	if len(hits) != 1 || hits[0].Ref != place.Entity.Ref || hits[0].Title != "Montréal" {
		t.Fatalf("title %+v", hits)
	}
	if hits[0].MatchReason != "place" || hits[0].MatchSnippet != "Montreal" {
		t.Fatalf("context %+v", hits[0])
	}
}

func TestHandleSearchEventsAndPlaces(t *testing.T) {
	f := newHandleFixture(t)
	birth, err := propertyterms.Lookup(f.c, f.prop("event_type").ID, "birth", propertyterms.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	y, m := 1817, 5
	event := f.promote(f.subject("event", func(s subjects.Subject) []observations.Input {
		return []observations.Input{
			{SubjectID: s.ID, PropertyID: f.prop("event_type").ID, ValueTermID: birth.ID},
			{SubjectID: s.ID, PropertyID: f.prop("date").ID, Date: &datevalues.Value{Kind: datevalues.KindPoint, Calendar: "gregorian", StartYear: &y, StartMonth: &m}},
		}
	}), nil)
	place := f.promote(f.subject("place", func(s subjects.Subject) []observations.Input {
		return []observations.Input{{SubjectID: s.ID, PropertyID: f.prop("toponym").ID, ValueText: "York, Upper Canada", HasText: true}}
	}), nil)
	f.promote(f.person("York Smith"), nil)

	if hits := f.search("Birth", KindEvent); len(hits) != 1 || hits[0].Ref != event.Entity.Ref || hits[0].Title != "Unspecified birth" {
		t.Fatalf("event by type: %+v", hits)
	}
	if hits := f.search("1817", KindEvent); len(hits) != 1 || hits[0].Location.Section != SectionEvents {
		t.Fatalf("event by year: %+v", hits)
	}
	if hits := f.search("York", KindPlace); refsOf(hits) != fmt.Sprint([]string{"place:" + place.Entity.Ref}) {
		t.Fatalf("place only: %s", refsOf(hits))
	}
	if hits := f.search("York", KindPlace, KindPerson); len(hits) != 2 {
		t.Fatalf("both kinds: %s", refsOf(hits))
	}
}

func TestHandleWithoutValuesIsFoundByRef(t *testing.T) {
	f := newHandleFixture(t)
	bare := f.promote(f.person(), nil)
	hits := f.search(bare.Entity.Ref, KindPerson)
	if len(hits) != 1 || hits[0].Title != bare.Entity.Ref {
		t.Fatalf("got %+v", hits)
	}
}

func TestRebuildIndexesHandles(t *testing.T) {
	f := newHandleFixture(t)
	james := f.promote(f.person("James Robins"), nil)
	db, err := f.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := searchindex.ClearAll(db); err != nil {
		t.Fatal(err)
	}
	if err := searchindex.SetProjectionVersion(db, 0); err != nil {
		t.Fatal(err)
	}
	if err := searchindex.EnsureCatalog(f.c); err != nil {
		t.Fatal(err)
	}
	if hits := f.search("Robins", KindPerson); len(hits) != 1 || hits[0].Ref != james.Entity.Ref {
		t.Fatalf("got %s", refsOf(hits))
	}
}

func TestKindPriorityIsATenthOfTheScore(t *testing.T) {
	equal := []Hit{
		{Kind: KindSourceType, Title: "T", Score: 10},
		{Kind: KindSource, Title: "S", Score: 10},
		{Kind: KindPlace, Title: "P", Score: 10},
		{Kind: KindEvent, Title: "E", Score: 10},
		{Kind: KindPerson, Title: "N", Score: 10},
	}
	for i := range equal {
		equal[i].Score = applyScoreMix(equal[i].Score, equal[i].Kind)
	}
	sortHits(equal)
	if got := kindsOf(equal); fmt.Sprint(got) != fmt.Sprint([]string{KindPerson, KindEvent, KindPlace, KindSource, KindSourceType}) {
		t.Fatalf("equal text, got %v", got)
	}

	// A text score more than the priority share ahead still wins.
	strongerSource := applyScoreMix(12, KindSource)
	person := applyScoreMix(10, KindPerson)
	if strongerSource <= person {
		t.Fatalf("source text 12 should beat person text 10, got %v vs %v", strongerSource, person)
	}
}

func kindsOf(hits []Hit) []string {
	out := make([]string, len(hits))
	for i, h := range hits {
		out[i] = h.Kind
	}
	return out
}

func TestSimilarTitlesRankThePersonAheadOfTheSource(t *testing.T) {
	f := newHandleFixture(t)
	person := f.promote(f.person("James Robins"), nil)
	place := f.promote(f.subject("place", func(s subjects.Subject) []observations.Input {
		return []observations.Input{{SubjectID: s.ID, PropertyID: f.prop("toponym").ID, ValueText: "Robins", HasText: true}}
	}), nil)
	typeID := seedType(t, f.c, "Census", "")
	src, _, err := writes.Run(f.c, writes.Op{Action: "create_source", UserID: f.user}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, f.user, sources.CreateInput{
			SourceTypeID: typeID,
			Title:        "Robins parish",
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	hits, err := DefaultEngine().Search(context.Background(), f.c, Query{Text: "Robins"})
	if err != nil {
		t.Fatal(err)
	}
	index := map[string]int{}
	byID := map[string]Hit{}
	for i, h := range hits {
		index[h.ID] = i
		byID[h.ID] = h
	}
	personID := uuidString(person.Entity.ID)
	placeID := uuidString(place.Entity.ID)
	sourceID := uuidString(src.ID)
	for _, id := range []string{personID, placeID, sourceID} {
		if _, ok := index[id]; !ok {
			t.Fatalf("missing %s in %+v", id, hits)
		}
	}
	if index[personID] > index[sourceID] {
		t.Fatalf("similar titles should put the person ahead of the source: %+v", hits)
	}
	switch {
	case byID[placeID].Score == byID[personID].Score && index[personID] > index[placeID]:
		t.Fatalf("tied person should precede the place: %+v", hits)
	case byID[placeID].Score != byID[personID].Score && (byID[placeID].Score > byID[personID].Score) != (index[placeID] < index[personID]):
		t.Fatalf("place and person should follow score: %+v", hits)
	}
}

func TestMergedHandleLeavesTheIndex(t *testing.T) {
	f := newHandleFixture(t)
	a := f.promote(f.person("James Robins"), nil)
	b := f.promote(f.person("James Robins"), nil)
	db, err := f.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, a.Entity.ID, b.Entity.ID); err != nil {
		t.Fatal(err)
	}
	if err := searchindex.ReprojectHandles(db, [][]byte{b.Entity.ID}); err != nil {
		t.Fatal(err)
	}
	if hits := f.search("Robins", KindPerson); refsOf(hits) != fmt.Sprint([]string{"person:" + a.Entity.Ref}) {
		t.Fatalf("got %s", refsOf(hits))
	}
}

func runArtifactCreate(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error) {
	a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, in)
		})
	return a, err
}
