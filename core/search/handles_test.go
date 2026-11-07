package search

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/observations/observationstest"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/workspacelocation"
	"github.com/mendahu/provenencia/core/writes"
)

type handleFixture struct {
	*evidencelayertest.Records
	user     []byte
	source   []byte
	artifact []byte
	graph    []byte // the Source's one Evidence graph
}

func newHandleFixture(t *testing.T) *handleFixture {
	t.Helper()
	r := evidencelayertest.New(t)
	return &handleFixture{Records: r, user: evidencelayertest.UserID, source: r.Source.ID, artifact: r.Artifact.ID, graph: r.Graph(r.Source)}
}

// subject creates a Subject of kind with the given Observations.
func (f *handleFixture) subject(kind string, in func(s subjects.Subject) []observations.Input) subjects.Subject {
	f.T.Helper()
	s := subjectstest.On(f.T, f.C, f.user, f.graph, kind, "")
	if in != nil {
		obs := in(s)
		if _, _, err := observationstest.Cite(f.C, f.user, f.graph, f.artifact, evidencelayertest.Locator, obs...); err != nil {
			f.T.Fatal(err)
		}
	}
	return s
}

func (f *handleFixture) person(forms ...string) subjects.Subject {
	return f.subject("person", func(s subjects.Subject) []observations.Input {
		var in []observations.Input
		for _, form := range forms {
			in = append(in, observations.Input{SubjectID: s.ID, PropertyID: f.Prop("name").ID, Name: namevaluetest.Western(form)})
		}
		return in
	})
}

// place creates a place Subject with one toponym Observation per form.
func (f *handleFixture) place(toponyms ...string) subjects.Subject {
	return f.subject("place", func(s subjects.Subject) []observations.Input {
		var in []observations.Input
		for _, name := range toponyms {
			in = append(in, observations.Input{SubjectID: s.ID, PropertyID: f.Prop("toponym").ID, Scalar: propertymeta.TextScalar(name)})
		}
		return in
	})
}

func (f *handleFixture) promote(s subjects.Subject, onto []byte) promote.Result {
	f.T.Helper()
	return promotetest.Save(f.T, f.C, f.user, promote.Input{SubjectID: s.ID, EntityID: onto})
}

func (f *handleFixture) search(text string, kinds ...string) []Hit {
	f.T.Helper()
	hits, err := DefaultEngine().Search(context.Background(), f.C, Query{Text: text, Kinds: kinds})
	if err != nil {
		f.T.Fatal(err)
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
	james := f.promote(f.person("James Robins", "Kenneth Robins"), nil)
	f.promote(f.person("Mary Smith"), nil)

	t.Run("the omnibar default includes handles", func(t *testing.T) {
		hits := f.search("Robins")
		if refsOf(hits) != fmt.Sprint([]string{"person:" + james.Entity.Ref}) {
			t.Fatalf("got %s", refsOf(hits))
		}
	})

	t.Run("by rank-1 name, with location and member count", func(t *testing.T) {
		hits := f.search("James Robins", KindPerson)
		if len(hits) != 1 || hits[0].Ref != james.Entity.Ref || hits[0].Title != "James Kenneth Robins" {
			t.Fatalf("got %+v", hits)
		}
		loc := hits[0].Location
		if loc.Section != workspacelocation.Persons || loc.EntityID != uuidString(james.Entity.ID) || loc.Ref != james.Entity.Ref {
			t.Fatalf("location %+v", loc)
		}
		if hits[0].MemberCount != 1 {
			t.Fatalf("members %d", hits[0].MemberCount)
		}
	})

	t.Run("by another name cluster", func(t *testing.T) {
		if hits := f.search("Kenneth", KindPerson); refsOf(hits) != fmt.Sprint([]string{"person:" + james.Entity.Ref}) {
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
	thomas := f.promote(f.person("Thomas Lee", "thomas lee", "Thommas Lee"), nil)
	if hits := f.search("Thommas", KindPerson); refsOf(hits) != fmt.Sprint([]string{"person:" + thomas.Entity.Ref}) {
		t.Fatalf("got %s", refsOf(hits))
	}
}

func TestHandleSearchFollowsEdits(t *testing.T) {
	f := newHandleFixture(t)
	s := f.subject("person", func(s subjects.Subject) []observations.Input {
		return []observations.Input{{SubjectID: s.ID, PropertyID: f.Prop("name").ID, Name: namevaluetest.Western("Ada Byron")}}
	})
	ada := f.promote(s, nil)
	obs, err := observations.ListBySubject(f.C, s.ID)
	if err != nil || len(obs) != 1 {
		t.Fatalf("%v %d", err, len(obs))
	}
	nameID := f.Prop("name").ID
	if _, err := writes.Call(f.C, writes.Op{Action: "update_observation", UserID: f.user}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
		return observations.Update(tx, f.user, observations.Input{
			ID: obs[0].ID, SubjectID: s.ID, PropertyID: nameID, Name: namevaluetest.Western("Ada Lovelace"),
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

// locatedBirth creates a person from forms, a birth Event (1817) and a place
// with toponym, their three positions, and the participation and location
// bridges. It stops before promote and fileBridges.
func (f *handleFixture) locatedBirth(toponym string, forms ...string) (person, event, place subjects.Subject) {
	f.T.Helper()
	t := f.T
	person = f.person(forms...)
	birth, err := propertyterms.Lookup(f.C, f.Prop("event_type").ID, "birth", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	event = f.subject("event", func(s subjects.Subject) []observations.Input {
		return []observations.Input{
			{SubjectID: s.ID, PropertyID: f.Prop("event_type").ID, ValueTermID: birth.ID},
			{SubjectID: s.ID, PropertyID: f.Prop("date").ID, Date: datevaluetest.YMD(1817, 0, 0)},
		}
	})
	place = f.place(toponym)
	for i, s := range []subjects.Subject{person, event, place} {
		if _, err := writes.Call(f.C, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
			return subjectpositions.Set(tx, s.ID, int64(i), 0)
		}); err != nil {
			t.Fatal(err)
		}
	}
	role, err := propertyterms.Lookup(f.C, f.Prop("role").ID, "subject", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	personProp := f.Prop("person").ID
	eventProp := f.Prop("event").ID
	roleProp := f.Prop("role").ID
	if _, err := writes.Call(f.C, writes.Op{Action: "create_cited_bridge", UserID: f.user}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, f.user, connect.CreateInput{
			FromSubjectID: person.ID, ToSubjectID: event.ID, BridgeTypeKey: "participation",
			Citation: citations.CreateInput{ArtifactID: f.artifact, LocatorJSON: evidencelayertest.Locator},
			Observations: []observations.Input{
				{PropertyID: personProp, ValueSubjectID: person.ID},
				{PropertyID: eventProp, ValueSubjectID: event.ID},
				{PropertyID: roleProp, ValueTermID: role.ID},
			},
		})
	}); err != nil {
		t.Fatal(err)
	}
	placeProp := f.Prop("place").ID
	if _, err := writes.Call(f.C, writes.Op{Action: "create_cited_bridge", UserID: f.user}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, f.user, connect.CreateInput{
			FromSubjectID: event.ID, ToSubjectID: place.ID, BridgeTypeKey: "location",
			Citation: citations.CreateInput{ArtifactID: f.artifact, LocatorJSON: evidencelayertest.Locator},
			Observations: []observations.Input{
				{PropertyID: eventProp, ValueSubjectID: event.ID},
				{PropertyID: placeProp, ValueSubjectID: place.ID},
			},
		})
	}); err != nil {
		t.Fatal(err)
	}
	return person, event, place
}

func TestSummarySearchFindsTheListTitle(t *testing.T) {
	f := newHandleFixture(t)
	person, event, place := f.locatedBirth("York", "James Robins", "Kenneth Robins")
	promoted := f.promote(person, nil)
	f.promote(event, nil)
	f.promote(place, nil)
	f.fileBridges()

	if hits := f.search("Kenneth Robins", KindPerson); len(hits) != 1 || hits[0].Ref != promoted.Entity.Ref {
		t.Fatalf("alternate name: %s", refsOf(hits))
	}
	if hits := f.search(promoted.Entity.Ref, KindPerson); len(hits) != 1 || hits[0].MatchReason != "ref" {
		t.Fatalf("ref: %+v", hits)
	}
	if hits := f.search("James", KindEvent); len(hits) != 1 || hits[0].Title != "Birth of James Kenneth Robins" {
		t.Fatalf("event title: %+v", hits)
	}
	if hits := f.search("York", KindPlace); len(hits) != 1 {
		t.Fatalf("place: %s", refsOf(hits))
	}

	obs, err := observations.ListBySubject(f.C, person.ID)
	if err != nil {
		t.Fatal(err)
	}
	var jamesObs observations.Listed
	for _, o := range obs {
		if o.ValueNameForm == "James Robins" || o.Text == "James Robins" {
			jamesObs = o
			break
		}
	}
	if len(jamesObs.ID) != 16 {
		t.Fatalf("james observation missing: %+v", obs)
	}
	nameID := f.Prop("name").ID
	if _, err := writes.Call(f.C, writes.Op{Action: "update_observation", UserID: f.user}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
		return observations.Update(tx, f.user, observations.Input{
			ID: jamesObs.ID, SubjectID: person.ID, PropertyID: nameID,
			Name: namevaluetest.Western("John Robins"),
		})
	}); err != nil {
		t.Fatal(err)
	}
	if hits := f.search("John", KindEvent); len(hits) != 1 || hits[0].Title != "Birth of John Kenneth Robins" {
		t.Fatalf("after edit: %+v", hits)
	}
	if hits := f.search("James", KindEvent); len(hits) != 0 {
		t.Fatalf("old name still titles the event: %+v", hits)
	}
}

func (f *handleFixture) fileBridges() {
	f.T.Helper()
	db, err := f.C.DB()
	if err != nil {
		f.T.Fatal(err)
	}
	var rev int64
	if err := db.QueryRow(`SELECT COALESCE(MAX(revision), 0) FROM audit_transactions`).Scan(&rev); err != nil {
		f.T.Fatal(err)
	}
	if _, err := writes.Call(f.C, writes.Op{Action: "promote_batch", UserID: f.user}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		return promote.SaveBatch(tx, f.user, promote.Batch{GraphID: f.graph, SeenRevision: rev})
	}); err != nil {
		f.T.Fatal(err)
	}
}

func TestAlternateToponymIsMatchContext(t *testing.T) {
	f := newHandleFixture(t)
	place := f.promote(f.place("Montréal", "Montreal"), nil)
	hits := f.search("Montreal", KindPlace)
	if len(hits) != 1 || hits[0].Ref != place.Entity.Ref || hits[0].Title != "Montréal" {
		t.Fatalf("title %+v", hits)
	}
	if hits[0].MatchReason != "title" || hits[0].MatchSnippet != "" {
		t.Fatalf("a toponym that reads like the title is a title hit: %+v", hits[0])
	}

	g := newHandleFixture(t)
	other := g.promote(g.place("Montréal", "Ville-Marie"), nil)
	hits = g.search("Ville Marie", KindPlace)
	if len(hits) != 1 || hits[0].Ref != other.Entity.Ref || hits[0].Title != "Montréal" {
		t.Fatalf("title %+v", hits)
	}
	if hits[0].MatchReason != "place" || hits[0].MatchSnippet != "Ville-Marie" {
		t.Fatalf("context %+v", hits[0])
	}
}

func TestHandleSearchEventsAndPlaces(t *testing.T) {
	f := newHandleFixture(t)
	birth, err := propertyterms.Lookup(f.C, f.Prop("event_type").ID, "birth", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	event := f.promote(f.subject("event", func(s subjects.Subject) []observations.Input {
		return []observations.Input{
			{SubjectID: s.ID, PropertyID: f.Prop("event_type").ID, ValueTermID: birth.ID},
			{SubjectID: s.ID, PropertyID: f.Prop("date").ID, Date: datevaluetest.YMD(1817, 5, 0)},
		}
	}), nil)
	place := f.promote(f.place("York, Upper Canada"), nil)
	f.promote(f.person("York Smith"), nil)

	if hits := f.search("Birth", KindEvent); len(hits) != 1 || hits[0].Ref != event.Entity.Ref || hits[0].Title != "Unspecified birth" {
		t.Fatalf("event by type: %+v", hits)
	}
	if hits := f.search("1817", KindEvent); len(hits) != 1 || hits[0].Location.Section != workspacelocation.Events {
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
	db, err := f.C.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := searchindex.ClearAll(db); err != nil {
		t.Fatal(err)
	}
	if err := searchindex.SetProjectionVersion(db, 0); err != nil {
		t.Fatal(err)
	}
	if err := searchindex.EnsureCatalog(f.C); err != nil {
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
		{Kind: KindEvidenceGraph, Title: "G", Score: 10},
	}
	for i := range equal {
		equal[i].Score = applyScoreMix(equal[i].Score, equal[i].Kind)
	}
	sortHits(equal)
	if got := kindsOf(equal); fmt.Sprint(got) != fmt.Sprint([]string{KindPerson, KindEvent, KindPlace, KindSource, KindEvidenceGraph, KindSourceType}) {
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
	place := f.promote(f.place("Robins"), nil)
	typeID := seedType(t, f.C, "Census", "")
	src, _, err := writes.Run(f.C, writes.Op{Action: "create_source", UserID: f.user}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, f.user, sources.CreateInput{
			SourceTypeID: typeID,
			Title:        "Robins parish",
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	hits, err := DefaultEngine().Search(context.Background(), f.C, Query{Text: "Robins"})
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
	db, err := f.C.DB()
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

// TestHandleKindsArePrimaryKinds keeps the registry's Handle flags in step with
// the primary kinds in vocabkey, the one list the rest of the app reads.
func TestHandleKindsArePrimaryKinds(t *testing.T) {
	var got []string
	for _, s := range Registry {
		if s.Handle {
			got = append(got, s.Kind)
		}
	}
	want := vocabkey.PrimaryKinds()
	sort.Strings(got)
	sort.Strings(want)
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("Handle kinds = %v, want vocabkey.PrimaryKinds() = %v", got, want)
	}
}

func TestFoldedSpellingsFindPlaces(t *testing.T) {
	f := newHandleFixture(t)
	for _, tt := range []struct{ title, query string }{
		{"Ærøskøbing", "Aeroskobing"},
		{"Søby", "Soby"},
		{"Łask", "Lask"},
		{"Œuilly", "Oeuilly"},
		{"Straße", "Strasse"},
	} {
		place := f.promote(f.place(tt.title), nil)
		hits := f.search(tt.query, KindPlace)
		if len(hits) == 0 || hits[0].Ref != place.Entity.Ref || hits[0].MatchReason != "title" {
			t.Errorf("%q: want %s first as a title hit, got %+v", tt.query, tt.title, hits)
		}
	}

	t.Run("a deduped alternate keeps the place findable both ways", func(t *testing.T) {
		f := newHandleFixture(t)
		place := f.promote(f.place("Søby", "Soby"), nil)
		for _, q := range []string{"Soby", "Søby"} {
			if hits := f.search(q, KindPlace); len(hits) != 1 || hits[0].Ref != place.Entity.Ref {
				t.Errorf("%q: %+v", q, hits)
			}
		}
	})

	t.Run("a place written without the letter is found with it", func(t *testing.T) {
		f := newHandleFixture(t)
		place := f.promote(f.place("Soby"), nil)
		if hits := f.search("Søby", KindPlace); len(hits) != 1 || hits[0].Ref != place.Entity.Ref {
			t.Errorf("%+v", hits)
		}
	})
}

func TestNicknameFindsTheFormalName(t *testing.T) {
	t.Run("a folded nickname finds the person", func(t *testing.T) {
		f := newHandleFixture(t)
		thomas := f.promote(f.person("Tom Smith", "Thomas Smith"), nil)
		hits := f.search("Tom", KindPerson)
		if len(hits) != 1 || hits[0].Ref != thomas.Entity.Ref || hits[0].Title != "Thomas Tom Smith" ||
			hits[0].MatchReason != "title" || hits[0].MatchSnippet != "" {
			t.Fatalf("got %+v", hits)
		}
	})
	t.Run("a nickname no record spells finds the person", func(t *testing.T) {
		f := newHandleFixture(t)
		thomas := f.promote(f.person("Thomas Smith"), nil)
		if hits := f.search("Tommy", KindPerson); len(hits) != 1 || hits[0].Ref != thomas.Entity.Ref {
			t.Fatalf("got %+v", hits)
		}
	})
	t.Run("a tag word finds nothing", func(t *testing.T) {
		f := newHandleFixture(t)
		f.promote(f.person("Thomas Smith"), nil)
		for _, q := range []string{"Al", "Ali", "Alias"} {
			if hits := f.search(q, KindPerson); len(hits) != 0 {
				t.Fatalf("%q: got %+v", q, hits)
			}
		}
	})
	t.Run("an alternate's tag word finds nothing", func(t *testing.T) {
		f := newHandleFixture(t)
		f.promote(f.person("Tom Smith", "Thomas Smith"), nil)
		f.promote(f.place("York", "Yorke"), nil)
		if hits := f.search("Nam", KindPerson); len(hits) != 0 {
			t.Fatalf("Nam: got %+v", hits)
		}
		if hits := f.search("Topo", KindPlace); len(hits) != 0 {
			t.Fatalf("Topo: got %+v", hits)
		}
	})
	t.Run("a name that does not fold into it finds nothing", func(t *testing.T) {
		f := newHandleFixture(t)
		f.promote(f.person("William Smith"), nil)
		if hits := f.search("Robert", KindPerson); len(hits) != 0 {
			t.Fatalf("got %+v", hits)
		}
	})
}

func TestJoinReprojectsDependentsOnlyWhenTheRootMoves(t *testing.T) {
	f := newHandleFixture(t)
	person, event, place := f.locatedBirth("York", "James Robins")
	personH := f.promote(person, nil)
	eventH := f.promote(event, nil)
	placeH := f.promote(place, nil)
	f.fileBridges()

	db, err := f.C.DB()
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range []struct {
		kind string
		id   []byte
		ref  string
	}{{KindPerson, personH.Entity.ID, personH.Entity.Ref}, {KindEvent, eventH.Entity.ID, eventH.Entity.Ref}} {
		if err := searchindex.Upsert(db, searchindex.Document{
			Kind: h.kind, EntityID: uuidString(h.id), DisplayRef: h.ref, Ref: h.ref, DisplayTitle: "stalemarker", Title: "stalemarker",
		}); err != nil {
			t.Fatal(err)
		}
	}
	found := func(text string) string {
		return refsOf(f.search(text, KindPerson, KindEvent))
	}
	both := fmt.Sprint([]string{"event:" + eventH.Entity.Ref, "person:" + personH.Entity.Ref})

	f.promote(f.place("York"), placeH.Entity.ID)
	if got := found("stalemarker"); got != both {
		t.Fatalf("a join that moves no shown value reprojected a dependent: %s", got)
	}

	f.promote(f.place("Eboracum"), placeH.Entity.ID)
	if got := found("stalemarker"); got != "[]" {
		t.Fatalf("marker survived a join that adds a toponym: %s", got)
	}
	if got := found("Eboracum"); got != both {
		t.Fatalf("Eboracum finds %s, want the Event and the Person", got)
	}
}
