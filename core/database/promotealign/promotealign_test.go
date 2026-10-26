package promotealign_test

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	_ "github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/observations/observationstest"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/segmentcompare/segmentcomparetest"
	"github.com/mendahu/provenencia/core/valuecodec"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

const locator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

type fixture struct {
	t        testing.TB
	c        *database.Catalog
	source   sources.Source
	artifact artifacts.Artifact
	y        int64
}

func must(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func newFixture(t testing.TB) *fixture {
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
	return &fixture{t: t, c: c, source: src, artifact: art}
}

func (f *fixture) newSource(title string) (sources.Source, artifacts.Artifact) {
	f.t.Helper()
	typeID, err := sourcetypes.Upsert(f.c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	must(f.t, err)
	src, _, err := writes.Run(f.c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: title})
	})
	must(f.t, err)
	art, err := artifactstest.Create(f.c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	must(f.t, err)
	return src, art
}

func (f *fixture) prop(key string) properties.Property {
	f.t.Helper()
	p, err := properties.Lookup(f.c, key, vocabkey.OriginProvenencia)
	must(f.t, err)
	return p
}

func (f *fixture) term(propKey, termKey string) propertyterms.Term {
	f.t.Helper()
	t, err := propertyterms.Lookup(f.c, f.prop(propKey).ID, termKey, vocabkey.OriginProvenencia)
	must(f.t, err)
	return t
}

func (f *fixture) bare(src sources.Source, kind string) subjects.Subject {
	f.t.Helper()
	s := subjectstest.On(f.t, f.c, userID, f.graph(src), kind, "")
	_, err := writes.Call(f.c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
		return subjectpositions.Set(tx, s.ID, 0, f.y)
	})
	must(f.t, err)
	f.y += 2
	return s
}

func (f *fixture) cite(art artifacts.Artifact, s subjects.Subject, in ...observations.Input) {
	f.t.Helper()
	for i := range in {
		in[i].SubjectID = s.ID
	}
	_, err := writes.Call(f.c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, userID, citations.CreateInput{
			ArtifactID: art.ID, LocatorJSON: locator,
		}, in)
	})
	must(f.t, err)
}

func (f *fixture) person(src sources.Source, art artifacts.Artifact, name string) subjects.Subject {
	f.t.Helper()
	s := f.bare(src, "person")
	f.cite(art, s, observations.Input{
		PropertyID: f.prop("name").ID, Name: namevaluetest.Western(name),
	})
	return s
}

func (f *fixture) birth(src sources.Source, art artifacts.Artifact) subjects.Subject {
	f.t.Helper()
	s := f.bare(src, "event")
	f.cite(art, s, observations.Input{
		PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID,
	})
	return s
}

func (f *fixture) place(src sources.Source, art artifacts.Artifact, name string) subjects.Subject {
	f.t.Helper()
	s := f.bare(src, "place")
	f.cite(art, s, observations.Input{
		PropertyID: f.prop("toponym").ID, ValueText: name, HasText: true,
	})
	return s
}

func (f *fixture) participation(src sources.Source, art artifacts.Artifact, person, event subjects.Subject, role string) {
	f.t.Helper()
	personProp := f.prop("person").ID
	eventProp := f.prop("event").ID
	roleProp := f.prop("role").ID
	roleTerm := f.term("role", role).ID
	_, err := writes.Call(f.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, userID, connect.CreateInput{
			FromSubjectID: person.ID, ToSubjectID: event.ID, BridgeTypeKey: "participation",
			Citation: citations.CreateInput{ArtifactID: art.ID, LocatorJSON: locator},
			Observations: []observations.Input{
				{PropertyID: personProp, ValueSubjectID: person.ID},
				{PropertyID: eventProp, ValueSubjectID: event.ID},
				{PropertyID: roleProp, ValueTermID: roleTerm},
			},
		})
	})
	must(f.t, err)
}

func (f *fixture) locate(src sources.Source, art artifacts.Artifact, event, place subjects.Subject) {
	f.t.Helper()
	eventProp := f.prop("event").ID
	placeProp := f.prop("place").ID
	_, err := writes.Call(f.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, userID, connect.CreateInput{
			FromSubjectID: event.ID, ToSubjectID: place.ID, BridgeTypeKey: "location",
			Citation: citations.CreateInput{ArtifactID: art.ID, LocatorJSON: locator},
			Observations: []observations.Input{
				{PropertyID: eventProp, ValueSubjectID: event.ID},
				{PropertyID: placeProp, ValueSubjectID: place.ID},
			},
		})
	})
	must(f.t, err)
}

func (f *fixture) placeRel(src sources.Source, art artifacts.Artifact, from, to subjects.Subject, kind string) {
	f.t.Helper()
	fromProp := f.prop("from").ID
	toProp := f.prop("to").ID
	typeProp := f.prop("place_relationship_type").ID
	kindTerm := f.term("place_relationship_type", kind).ID
	_, err := writes.Call(f.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, userID, connect.CreateInput{
			FromSubjectID: from.ID, ToSubjectID: to.ID, BridgeTypeKey: "place_relationship",
			Citation: citations.CreateInput{ArtifactID: art.ID, LocatorJSON: locator},
			Observations: []observations.Input{
				{PropertyID: fromProp, ValueSubjectID: from.ID},
				{PropertyID: toProp, ValueSubjectID: to.ID},
				{PropertyID: typeProp, ValueTermID: kindTerm},
			},
		})
	})
	must(f.t, err)
}

// relationship cites "person is typ of related" on src.
func (f *fixture) relationship(src sources.Source, art artifacts.Artifact, person, related subjects.Subject, typ string) {
	f.t.Helper()
	personProp := f.prop("person").ID
	relatedProp := f.prop("related_to").ID
	typeProp := f.prop("relationship_type").ID
	typTerm := f.term("relationship_type", typ).ID
	_, err := writes.Call(f.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, userID, connect.CreateInput{
			FromSubjectID: person.ID, ToSubjectID: related.ID, BridgeTypeKey: "relationship",
			Citation: citations.CreateInput{ArtifactID: art.ID, LocatorJSON: locator},
			Observations: []observations.Input{
				{PropertyID: personProp, ValueSubjectID: person.ID},
				{PropertyID: relatedProp, ValueSubjectID: related.ID},
				{PropertyID: typeProp, ValueTermID: typTerm},
			},
		})
	})
	must(f.t, err)
}

// sexOnly is a person whose record gives no name, only a sex at birth.
func (f *fixture) sexOnly(src sources.Source, art artifacts.Artifact, sex string) subjects.Subject {
	f.t.Helper()
	s := f.bare(src, "person")
	f.cite(art, s, observations.Input{
		PropertyID: f.prop("sex_at_birth").ID, ValueTermID: f.term("sex_at_birth", sex).ID,
	})
	return s
}

func (f *fixture) promote(s subjects.Subject) promote.Result {
	f.t.Helper()
	return promotetest.Save(f.t, f.c, userID, promote.Input{SubjectID: s.ID})
}

// graph is the one Evidence graph of src. Call it outside a write closure.
func (f *fixture) graph(src sources.Source) []byte {
	f.t.Helper()
	return evidencegraphstest.SoleGraph(f.t, f.c, src.ID).ID
}

func (f *fixture) propose(graphID []byte, fixed []graphalign.Fixed) graphalign.Proposal {
	f.t.Helper()
	p, _, err := promotealign.Propose(f.c, graphID, fixed)
	must(f.t, err)
	return p
}

func rowFor(p graphalign.Proposal, subjectID []byte) graphalign.Row {
	for _, r := range p.Rows {
		if bytes.Equal(r.SubjectID, subjectID) {
			return r
		}
	}
	return graphalign.Row{}
}

// A Source whose only Subject is fixed has nothing to seed, so Propose does not
// read every node to build the shortlist index.
func TestProposeWithOnlyFixedSubjectsBuildsNoIndex(t *testing.T) {
	f := newFixture(t)
	aPerson := f.person(f.source, f.artifact, "Gracie")
	per := f.promote(aPerson)
	srcB, artB := f.newSource("Obituary")
	bPerson := f.person(srcB, artB, "Gracie")

	f.propose(f.graph(srcB), []graphalign.Fixed{{SubjectID: bPerson.ID, HandleID: per.Entity.ID}})
	if _, ok := f.c.Graph().MatchIndex(); ok {
		t.Error("Propose built the match index with no Subject to seed")
	}
	f.propose(f.graph(srcB), nil)
	if _, ok := f.c.Graph().MatchIndex(); !ok {
		t.Error("Propose with an unfixed Subject did not build the match index")
	}
}

// Seed from an already-promoted birth: the unpromoted person on a new Source
// should accept the matching Gracie handle through the participation edge.
func TestProposeSeedFromPromotedNeighbor(t *testing.T) {
	f := newFixture(t)

	// Canon: Gracie ↔ birth filed on Source A.
	aPerson := f.person(f.source, f.artifact, "Gracie")
	aBirth := f.birth(f.source, f.artifact)
	f.participation(f.source, f.artifact, aPerson, aBirth, "subject")
	per := f.promote(aPerson)
	evt := f.promote(aBirth)

	// Layer B: matching person + birth; fix the birth to the known handle.
	srcB, artB := f.newSource("Obituary")
	bPerson := f.person(srcB, artB, "Gracie")
	bBirth := f.birth(srcB, artB)
	f.participation(srcB, artB, bPerson, bBirth, "subject")

	got := f.propose(f.graph(srcB), []graphalign.Fixed{{
		SubjectID: bBirth.ID, HandleID: evt.Entity.ID,
	}})
	row := rowFor(got, bPerson.ID)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, per.Entity.ID) {
		t.Fatalf("person row %+v, want handle %x", row, per.Entity.ID)
	}
	if row.Via == nil || !bytes.Equal(row.Via.NeighborSubjectID, bBirth.ID) || row.Via.Key.Type != vocabkey.BridgeParticipation {
		t.Fatalf("via %+v, want birth participation", row.Via)
	}
	birthRow := rowFor(got, bBirth.ID)
	if birthRow.Target != graphalign.TargetHandle || !bytes.Equal(birthRow.HandleID, evt.Entity.ID) {
		t.Fatalf("birth row %+v, want fixed %x", birthRow, evt.Entity.ID)
	}
}

func TestProposePlacePartOfChain(t *testing.T) {
	f := newFixture(t)

	york := f.place(f.source, f.artifact, "York")
	ontario := f.place(f.source, f.artifact, "Ontario")
	f.placeRel(f.source, f.artifact, york, ontario, "part_of")
	yorkH := f.promote(york)
	ontarioH := f.promote(ontario)

	srcB, artB := f.newSource("Gazetteer")
	york2 := f.place(srcB, artB, "York")
	ontario2 := f.place(srcB, artB, "Ontario")
	f.placeRel(srcB, artB, york2, ontario2, "part_of")

	got := f.propose(f.graph(srcB), []graphalign.Fixed{{
		SubjectID: ontario2.ID, HandleID: ontarioH.Entity.ID,
	}})
	row := rowFor(got, york2.ID)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, yorkH.Entity.ID) {
		t.Fatalf("york row %+v, want handle %x", row, yorkH.Entity.ID)
	}
}

func TestProposeOverlappingBoundShowsTheDateAndResembles(t *testing.T) {
	f := newFixture(t)
	bef := func(year, month, day int) *datevalue.Value {
		y, m, d := year, month, day
		return &datevalue.Value{
			Kind: datevalue.KindPoint, Qualifier: datevalue.QualifierBEF,
			StartYear: &y, StartMonth: &m, StartDay: &d,
		}
	}
	// A death, not a birth: two births "before" different dates would put the
	// person born on two dates, and the lifespan would rule the pair out. A
	// death is a one-cardinality link, so the event files through its subject.
	earlier, later := bef(1985, 2, 1), bef(1985, 3, 31)

	aPerson := f.person(f.source, f.artifact, "Doug")
	aEvent := f.bare(f.source, "event")
	f.cite(f.artifact, aEvent,
		observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "death").ID},
		observations.Input{PropertyID: f.prop("start_date").ID, Date: earlier},
	)
	f.participation(f.source, f.artifact, aPerson, aEvent, "subject")
	per := f.promote(aPerson)
	evt := f.promote(aEvent)

	srcB, artB := f.newSource("Letter")
	bPerson := f.person(srcB, artB, "Doug")
	bEvent := f.bare(srcB, "event")
	f.cite(artB, bEvent,
		observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "death").ID},
		observations.Input{PropertyID: f.prop("start_date").ID, Date: later},
	)
	f.participation(srcB, artB, bPerson, bEvent, "subject")

	got := f.propose(f.graph(srcB), []graphalign.Fixed{{SubjectID: bPerson.ID, HandleID: per.Entity.ID}})
	row := rowFor(got, bEvent.ID)
	if !bytes.Equal(row.HandleID, evt.Entity.ID) {
		t.Fatalf("event %+v, want %s", row, evt.Entity.Ref)
	}
}

func TestProposeInvalidGraph(t *testing.T) {
	f := newFixture(t)
	_, _, err := promotealign.Propose(f.c, make([]byte, 16), nil)
	if err != promote.ErrInvalid {
		t.Fatalf("err %v, want promote.ErrInvalid", err)
	}
}

// A Source with two graphs proposes one graph's Subjects at a time.
func TestProposeReadsOneGraph(t *testing.T) {
	f := newFixture(t)
	census, censusArt := f.newSource("Census")
	def, jones, page2 := artifactstest.Split(t, f.c, userID, census.ID, "Jones")
	smith := subjectstest.On(t, f.c, userID, def.ID, "person", "")
	f.cite(censusArt, smith, observations.Input{PropertyID: f.prop("name").ID, Name: namevaluetest.Western("John Smith")})
	mary := subjectstest.On(t, f.c, userID, jones.ID, "person", "")
	f.cite(page2, mary, observations.Input{PropertyID: f.prop("name").ID, Name: namevaluetest.Western("Mary Jones")})

	for _, tc := range []struct {
		name        string
		graph, want []byte
	}{
		{"Jones", jones.ID, mary.ID},
		{"Default", def.ID, smith.ID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := f.propose(tc.graph, nil)
			if len(p.Rows) != 1 || !bytes.Equal(p.Rows[0].SubjectID, tc.want) {
				t.Fatalf("rows %+v, want only %x", p.Rows, tc.want)
			}
		})
	}
	for name, id := range map[string][]byte{"no such graph": make([]byte, 16), "a short id": {1, 2, 3}} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := promotealign.Propose(f.c, id, nil); !errors.Is(err, promote.ErrInvalid) {
				t.Fatalf("err %v, want promote.ErrInvalid", err)
			}
		})
	}
}

func TestProposeHoldsNewAndSkipDecisions(t *testing.T) {
	tests := []struct {
		name    string
		fixed   func(handle []byte) graphalign.Fixed
		want    graphalign.Target
		wantErr bool
	}{
		{name: "a row decided New stays New", fixed: func([]byte) graphalign.Fixed { return graphalign.Fixed{Target: graphalign.TargetNew} }, want: graphalign.TargetNew},
		{name: "a row decided Skip stays Skip", fixed: func([]byte) graphalign.Fixed { return graphalign.Fixed{Target: graphalign.TargetSkip} }, want: graphalign.TargetSkip},
		{name: "a New decision with a handle is refused", fixed: func(h []byte) graphalign.Fixed {
			return graphalign.Fixed{Target: graphalign.TargetNew, HandleID: h}
		}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			per := f.promote(f.person(f.source, f.artifact, "Gracie"))
			srcB, artB := f.newSource("Obituary")
			again := f.person(srcB, artB, "Gracie")
			fixed := tt.fixed(per.Entity.ID)
			fixed.SubjectID = again.ID
			p, _, err := promotealign.Propose(f.c, f.graph(srcB), []graphalign.Fixed{fixed})
			if tt.wantErr {
				if err != promote.ErrInvalid {
					t.Fatalf("err %v, want promote.ErrInvalid", err)
				}
				return
			}
			must(t, err)
			if row := rowFor(p, again.ID); row.Target != tt.want {
				t.Fatalf("row %+v, want %s", row, tt.want)
			}
		})
	}
}

// Every graph read in Propose is batched, so a bigger layer and catalog cost
// no more queries than a small one of the same shape.
func TestProposeQueryCountDoesNotGrowWithTheLayer(t *testing.T) {
	queries := func(people int) int {
		f := newFixture(t)
		srcB, artB := f.newSource("Obituary")
		for i := 0; i < people; i++ {
			name := fmt.Sprintf("Person %c", 'A'+i)
			p := f.person(f.source, f.artifact, name)
			b := f.birth(f.source, f.artifact)
			f.participation(f.source, f.artifact, p, b, "subject")
			f.promote(p)
			f.promote(b)
			pb := f.person(srcB, artB, name)
			bb := f.birth(srcB, artB)
			f.participation(srcB, artB, pb, bb, "subject")
		}
		n := 0
		f.c.Graph().SetQueryCounterForTest(func() { n++ })
		_, _, err := promotealign.Propose(f.c, f.graph(srcB), nil)
		must(t, err)
		return n
	}
	small, large := queries(2), queries(6)
	t.Logf("%d queries per proposal", small)
	if small != large {
		t.Fatalf("2 people took %d queries, 6 took %d; a read grew with the layer", small, large)
	}
}

// A city, state, and country that agree on toponym and on part-of, with
// nothing decided yet, are strong. The hierarchy is one hop of context.
func TestPartOfChainIsAStrongMatch(t *testing.T) {
	f := newFixture(t)
	city := f.place(f.source, f.artifact, "Gumptiontown")
	state := f.place(f.source, f.artifact, "Provenance")
	country := f.place(f.source, f.artifact, "Arcadia")
	f.placeRel(f.source, f.artifact, city, state, "part_of")
	f.placeRel(f.source, f.artifact, state, country, "part_of")
	cityH := f.promote(city)
	stateH := f.promote(state)
	countryH := f.promote(country)

	srcB, artB := f.newSource("Gazetteer")
	city2 := f.place(srcB, artB, "Gumptiontown")
	state2 := f.place(srcB, artB, "Provenance")
	country2 := f.place(srcB, artB, "Arcadia")
	f.placeRel(srcB, artB, city2, state2, "part_of")
	f.placeRel(srcB, artB, state2, country2, "part_of")

	prop := f.propose(f.graph(srcB), nil)
	cfg := matchregistry.Default().AlignConfig()
	for _, pair := range []struct {
		sub    subjects.Subject
		handle []byte
	}{
		{city2, cityH.Entity.ID},
		{state2, stateH.Entity.ID},
		{country2, countryH.Entity.ID},
	} {
		row := rowFor(prop, pair.sub.ID)
		if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, pair.handle) ||
			row.Assessment != graphalign.AssessStrong || row.Score < cfg.Alignment.Bands.Strong {
			t.Fatalf("%s: %+v", pair.sub.Ref, row)
		}
	}
}

// A place that shares its only toponym with a handle is a medium match, even
// when another place in the catalog makes that name look common.
func TestSameToponymAloneIsAMediumMatch(t *testing.T) {
	f := newFixture(t)
	york := f.promote(f.place(f.source, f.artifact, "York"))
	f.promote(f.place(f.source, f.artifact, "Leeds"))

	srcB, artB := f.newSource("Gazetteer")
	again := f.place(srcB, artB, "York")
	row := rowFor(f.propose(f.graph(srcB), nil), again.ID)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, york.Entity.ID) ||
		row.Assessment != graphalign.AssessMedium {
		t.Fatalf("york row %+v, want %s as a weak match", row, york.Entity.Ref)
	}
}

func TestShorterStructuredNameIsAWeakMatch(t *testing.T) {
	cases := []struct {
		full, short string
	}{
		{"James Kenneth Robins", "James Robins"},
		{"Lee-Ellen Matilda Breakell", "Lee Breakell"},
	}
	for _, tc := range cases {
		t.Run(tc.short, func(t *testing.T) {
			f := newFixture(t)
			known := f.promote(f.person(f.source, f.artifact, tc.full))
			srcB, artB := f.newSource("Census")
			again := f.person(srcB, artB, tc.short)
			row := rowFor(f.propose(f.graph(srcB), nil), again.ID)
			cfg := matchregistry.Default().AlignConfig()
			if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, known.Entity.ID) ||
				row.Assessment != graphalign.AssessWeak ||
				row.Score < cfg.Alignment.Bands.Weak || row.Score >= cfg.Alignment.Bands.Medium {
				t.Fatalf("%s row %+v, want %s as a weak match", tc.short, row, known.Entity.Ref)
			}
		})
	}
}

// "Provenance" shares no word with "Provenanced", so the shortlist never offers
// the handle: the accepted gap for a spelling-only toponym (S11-15).
// dogfood: Spike 11 matching dogfood: if spelling-only names or toponyms go
// unmatched there, revisit the coarse misspelling key (one more prefix in
// valuecompare's wordKeys).
func TestSpellingOnlyToponymIsNotShortlisted(t *testing.T) {
	f := newFixture(t)
	f.promote(f.place(f.source, f.artifact, "Provenanced"))
	f.promote(f.place(f.source, f.artifact, "Leeds"))

	srcB, artB := f.newSource("Gazetteer")
	again := f.place(srcB, artB, "Provenance")
	row := rowFor(f.propose(f.graph(srcB), nil), again.ID)
	if row.Target != graphalign.TargetNew || row.Reason != graphalign.ReasonNoMatch {
		t.Fatalf("provenance row %+v, want a new handle with no_match", row)
	}
}

// A directed term is matched from the same end: the fixed place's part (not
// its whole, though both are named York) is what the layer's part matches.
func TestProposePartOfMatchesTheSameEnd(t *testing.T) {
	f := newFixture(t)
	part := f.place(f.source, f.artifact, "York")
	middle := f.place(f.source, f.artifact, "Ontario")
	whole := f.place(f.source, f.artifact, "York")
	f.placeRel(f.source, f.artifact, part, middle, "part_of")
	f.placeRel(f.source, f.artifact, middle, whole, "part_of")
	partH := f.promote(part)
	middleH := f.promote(middle)
	f.promote(whole)
	// Other places, so "York" is not a common name and the edge decides.
	for _, name := range []string{"Bath", "Leeds", "Hull", "Derby", "Ely", "Wells"} {
		f.promote(f.place(f.source, f.artifact, name))
	}

	srcB, artB := f.newSource("Gazetteer")
	york := f.place(srcB, artB, "York")
	ontario := f.place(srcB, artB, "Ontario")
	f.placeRel(srcB, artB, york, ontario, "part_of")

	row := rowFor(f.propose(f.graph(srcB), []graphalign.Fixed{{SubjectID: ontario.ID, HandleID: middleH.Entity.ID}}), york.ID)
	if row.Reason != graphalign.ReasonVia || !bytes.Equal(row.HandleID, partH.Entity.ID) {
		t.Fatalf("york row %+v, want the part %s via Ontario", row, partH.Entity.Ref)
	}
}

// "Mary parent of John" on one Source and "John child of Mary" on another are
// one relationship: the walk from Mary reaches John's handle either way.
func TestProposeInverseKinshipTermsCorrespond(t *testing.T) {
	f := newFixture(t)
	mary := f.person(f.source, f.artifact, "Mary Smith")
	john := f.person(f.source, f.artifact, "John Smith")
	f.relationship(f.source, f.artifact, mary, john, "parent")
	hMary := f.promote(mary).Entity.ID
	hJohn := f.promote(john).Entity.ID

	src2, art2 := f.newSource("Baptism")
	son := f.person(src2, art2, "John Smith")
	mother := f.person(src2, art2, "Mary Smith")
	f.relationship(src2, art2, son, mother, "child")

	p := f.propose(f.graph(src2), []graphalign.Fixed{{SubjectID: mother.ID, HandleID: hMary}})
	r := rowFor(p, son.ID)
	if r.Target != graphalign.TargetHandle || !bytes.Equal(r.HandleID, hJohn) || r.Reason != graphalign.ReasonVia {
		t.Fatalf("son: target=%s handle-is-John=%v reason=%s, want John's handle via his mother",
			r.Target, bytes.Equal(r.HandleID, hJohn), r.Reason)
	}
}

// A relationship is walked from either end: anchoring the child reaches the
// parent as anchoring the parent reaches the child.
//
// dogfood: a parent or child link is an Other link, so an end with no name
// scores under the accept bar and is not filed; the walk still offers the
// right handle first. Revisit when the whole algorithm is dogfooded (Spike 11
// matching dogfood).
func TestProposeRelationshipFromEitherEnd(t *testing.T) {
	f := newFixture(t)
	mary := f.person(f.source, f.artifact, "Mary Smith")
	john := f.person(f.source, f.artifact, "John Smith")
	f.relationship(f.source, f.artifact, mary, john, "parent")
	hMary := f.promote(mary).Entity.ID
	hJohn := f.promote(john).Entity.ID

	tests := []struct {
		name           string
		anchorIsParent bool
		namedOther     bool
	}{
		{name: "child anchored reaches the parent"},
		{name: "parent anchored reaches the child", anchorIsParent: true},
		{name: "named other end files via the anchor", anchorIsParent: true, namedOther: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, art := f.newSource(tt.name)
			var parent, child subjects.Subject
			var fixed graphalign.Fixed
			var want []byte
			if tt.anchorIsParent {
				parent = f.person(src, art, "Mary Smith")
				child = f.sexOnly(src, art, "male")
				if tt.namedOther {
					child = f.person(src, art, "John Smith")
				}
				fixed = graphalign.Fixed{SubjectID: parent.ID, HandleID: hMary}
				want = hJohn
			} else {
				parent = f.sexOnly(src, art, "female")
				child = f.person(src, art, "John Smith")
				fixed = graphalign.Fixed{SubjectID: child.ID, HandleID: hJohn}
				want = hMary
			}
			f.relationship(src, art, parent, child, "parent")
			p := f.propose(f.graph(src), []graphalign.Fixed{fixed})
			other := child
			if !tt.anchorIsParent {
				other = parent
			}
			r := rowFor(p, other.ID)
			if tt.namedOther {
				if r.Target != graphalign.TargetHandle || !bytes.Equal(r.HandleID, want) || r.Reason != graphalign.ReasonVia {
					t.Fatalf("target=%s right-handle=%v reason=%s, want the other end's handle via the anchor",
						r.Target, bytes.Equal(r.HandleID, want), r.Reason)
				}
				if r.Via == nil || r.Via.Key.Type != vocabkey.BridgeRelationship || r.Via.Key.Term != "child" {
					t.Fatalf("via %+v, want a relationship key with term child", r.Via)
				}
				return
			}
			if r.Target != graphalign.TargetSkip || r.Reason != graphalign.ReasonWeak {
				t.Fatalf("target=%s reason=%s, want skip and weak (under the accept bar)", r.Target, r.Reason)
			}
			if len(r.Alternatives) == 0 || !bytes.Equal(r.Alternatives[0].HandleID, want) || r.Alternatives[0].Reason != graphalign.ReasonVia {
				t.Fatalf("alternatives %+v, want the other end's handle first, via the anchor", r.Alternatives)
			}
		})
	}
}

// A place with several filed events still proposes one row per layer
// subject, and the place lands on its handle.
func TestPlaceWithManyEventsProposalStaysOnTheLayer(t *testing.T) {
	f := newFixture(t)
	city := f.place(f.source, f.artifact, "Gumptiontown")
	cityH := f.promote(city)
	for i := 0; i < 3; i++ {
		person := f.person(f.source, f.artifact, fmt.Sprintf("Neighbor %d", i))
		birth := f.birth(f.source, f.artifact)
		f.participation(f.source, f.artifact, person, birth, "subject")
		f.locate(f.source, f.artifact, birth, city)
		f.promote(person)
		f.promote(birth)
	}

	srcB, artB := f.newSource("Gazetteer")
	city2 := f.place(srcB, artB, "Gumptiontown")
	birth2 := f.birth(srcB, artB)
	f.locate(srcB, artB, birth2, city2)

	got := f.propose(f.graph(srcB), nil)
	if len(got.Rows) != 2 {
		t.Fatalf("rows %d, want the layer's place and birth", len(got.Rows))
	}
	row := rowFor(got, city2.ID)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, cityH.Entity.ID) {
		t.Fatalf("place row %+v, want the city's handle", row.Target)
	}
	if err := f.c.Graph().Verify(); err != nil {
		t.Fatal(err)
	}
}

// A sex at birth the catalog flags neutral is not compared: an "unknown"
// record of John Smith still lands on the male John Smith, while "female"
// conflicts and keeps it under the bar.
func TestProposeIgnoresANeutralSex(t *testing.T) {
	f := newFixture(t)
	named := func(src sources.Source, art artifacts.Artifact, sex string) subjects.Subject {
		s := f.person(src, art, "John Smith")
		f.cite(art, s, observations.Input{
			PropertyID: f.prop("sex_at_birth").ID, ValueTermID: f.term("sex_at_birth", sex).ID,
		})
		return s
	}
	handle := f.promote(named(f.source, f.artifact, "male")).Entity.ID

	src, art := f.newSource("Register")
	unknown := named(src, art, "unknown")
	r := rowFor(f.propose(f.graph(src), nil), unknown.ID)
	if r.Target != graphalign.TargetHandle || !bytes.Equal(r.HandleID, handle) {
		t.Fatalf("unknown sex: target %s score %.2f, want the male John Smith", r.Target, r.Score)
	}
	for _, c := range r.Comparisons {
		if c.Key == vocabkey.PropertySexAtBirth && c.Result.Outcome != propertycompare.OutcomeUnknown {
			t.Fatalf("unknown sex compared as %s", c.Result.Outcome)
		}
	}

	src2, art2 := f.newSource("Census")
	female := named(src2, art2, "female")
	if r := rowFor(f.propose(f.graph(src2), nil), female.ID); r.Target == graphalign.TargetHandle && bytes.Equal(r.HandleID, handle) {
		t.Fatalf("female John Smith landed on the male one (score %.2f)", r.Score)
	}
}

// Jake Robins against James Kenneth Jake Robins: Jake and Robins match, and
// James and Kenneth exist only on the conclusion, so the name is a resemblance
// (2.5 × 0.75), below a full match. Gender agrees as well (0.6). Together that
// is a weak match.
func TestProposeExtraGivenNamesAreAWeakMatch(t *testing.T) {
	f := newFixture(t)
	withGender := func(src sources.Source, art artifacts.Artifact, name string) subjects.Subject {
		s := f.person(src, art, name)
		f.cite(art, s, observations.Input{
			PropertyID: f.prop("gender").ID, ValueTermID: f.term("gender", "man").ID,
		})
		return s
	}
	james := withGender(f.source, f.artifact, "James Kenneth Robins")
	handle := f.promote(james).Entity.ID
	birthSrc, birthArt := f.newSource("Birth Announcement")
	f.join(withGender(birthSrc, birthArt, "Jake Robins"), handle)

	noteSrc, noteArt := f.newSource("Note")
	probe := withGender(noteSrc, noteArt, "Jake Robins")
	row := rowFor(f.propose(f.graph(noteSrc), nil), probe.ID)
	cfg := matchregistry.Default().AlignConfig()
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, handle) ||
		row.Assessment != graphalign.AssessWeak || row.Score < cfg.Alignment.Bands.Weak || row.Score >= cfg.Alignment.Bands.Medium ||
		math.Abs(row.Score-(2.5*0.75+0.6)) > 1e-9 {
		t.Fatalf("Jake Robins: target %s assessment %s score %.4f", row.Target, row.Assessment, row.Score)
	}
}

// namedMembers files n people called name, each on a Source of its own titled
// "Census 1851", "Census 1861", ...: the first on a new handle, the rest joined
// to it. It returns the handle and the members in filing order.
func (f *fixture) namedMembers(n int, name string) (handle []byte, members []subjects.Subject) {
	f.t.Helper()
	for i := 0; i < n; i++ {
		src, art := f.newSource(fmt.Sprintf("Census %d", 1851+10*i))
		p := f.person(src, art, name)
		if i == 0 {
			handle = f.promote(p).Entity.ID
		} else {
			f.join(p, handle)
		}
		members = append(members, p)
	}
	return handle, members
}

func (f *fixture) join(s subjects.Subject, entityID []byte) {
	f.t.Helper()
	promotetest.Save(f.t, f.c, userID, promote.Input{SubjectID: s.ID, EntityID: entityID})
}

func comparisonOutcome(t *testing.T, row graphalign.Row, key string) propertycompare.Outcome {
	t.Helper()
	for _, c := range row.Comparisons {
		if string(c.Key) == key {
			return c.Result.Outcome
		}
	}
	t.Fatalf("no %s comparison in %+v", key, row.Comparisons)
	return ""
}

func TestProposeJudgesAgainstTheMajority(t *testing.T) {
	t.Run("a stray name", func(t *testing.T) {
		f := newFixture(t)
		handle := f.promote(f.person(f.source, f.artifact, "Mary Smith")).Entity.ID
		blob, err := valuecodec.MarshalName(*namevaluetest.Western("James Robins"))
		must(t, err)
		db, err := f.c.DB()
		must(t, err)
		_, err = db.Exec(`UPDATE auto_reconciler_values SET support = 5 WHERE entity_id = ? AND property_id = ? AND rank = 1`, handle, f.prop("name").ID)
		must(t, err)
		_, err = db.Exec(`INSERT INTO auto_reconciler_values (entity_id, property_id, rank, value_name, support, against, reason)
			VALUES (?, ?, 2, ?, 1, 0, 'outvoted')`, handle, f.prop("name").ID, blob)
		must(t, err)
		f.c.Graph().Drop()

		src, art := f.newSource("Register")
		incoming := f.person(src, art, "James Robins")
		row := rowFor(f.propose(f.graph(src), []graphalign.Fixed{{SubjectID: incoming.ID, HandleID: handle}}), incoming.ID)
		if got := comparisonOutcome(t, row, "name"); got != propertycompare.OutcomeConflict {
			t.Fatalf("name comparison %s, want conflict", got)
		}
	})

	t.Run("a negative Observation", func(t *testing.T) {
		f := newFixture(t)
		sex := func(src sources.Source, art artifacts.Artifact, key string) subjects.Subject {
			s := f.person(src, art, "Mary Smith")
			f.cite(art, s, observations.Input{PropertyID: f.prop("sex_at_birth").ID, ValueTermID: f.term("sex_at_birth", key).ID})
			return s
		}
		handle := f.promote(sex(f.source, f.artifact, "male")).Entity.ID
		db, err := f.c.DB()
		must(t, err)
		_, err = db.Exec(`UPDATE auto_reconciler_values SET support = 3 WHERE entity_id = ? AND property_id = ?`, handle, f.prop("sex_at_birth").ID)
		must(t, err)
		f.c.Graph().Drop()

		src, art := f.newSource("Census")
		incoming := f.person(src, art, "Mary Smith")
		f.cite(art, incoming, observations.Input{
			PropertyID: f.prop("sex_at_birth").ID, ValueTermID: f.term("sex_at_birth", "male").ID,
			Polarity: observations.PolarityNegative,
		})
		row := rowFor(f.propose(f.graph(src), []graphalign.Fixed{{SubjectID: incoming.ID, HandleID: handle}}), incoming.ID)
		if got := comparisonOutcome(t, row, "sex_at_birth"); got != propertycompare.OutcomeConflict {
			t.Fatalf("sex comparison %s, want conflict", got)
		}
	})
}

// Likeness has no weight: a person proposed with likenesses on the Source and
// on the handle gets the band, score and comparisons it gets without them.
func TestProposeIgnoresLikenesses(t *testing.T) {
	score := func(t *testing.T, withLikeness bool) graphalign.Row {
		f := newFixture(t)
		a := f.person(f.source, f.artifact, "John Smith")
		if withLikeness {
			observationstest.SaveLikeness(t, f.c, userID, f.source.ID, a.ID, 40)
		}
		f.promote(a)
		src, art := f.newSource("Obituary")
		incoming := f.person(src, art, "John Smith")
		if withLikeness {
			observationstest.SaveLikeness(t, f.c, userID, src.ID, incoming.ID, 200)
		}
		_ = art
		return rowFor(f.propose(f.graph(src), nil), incoming.ID)
	}
	with, without := score(t, true), score(t, false)
	if with.Target != without.Target || with.Assessment != without.Assessment || with.Score != without.Score || with.Flags.Contested != without.Flags.Contested || with.Reason != without.Reason ||
		len(with.Comparisons) != len(without.Comparisons) {
		t.Fatalf("with likenesses %+v\nwithout %+v", with, without)
	}
}

// A Place handle keeps every toponym it was read as; the minority one still
// agrees, because a multiple-cardinality Property weighs no value down.
func TestProposeMultipleToponyms(t *testing.T) {
	f := newFixture(t)
	handle := f.promote(f.place(f.source, f.artifact, "York")).Entity.ID
	db, err := f.c.DB()
	must(t, err)
	_, err = db.Exec(`UPDATE auto_reconciler_values SET support = 5 WHERE entity_id = ? AND property_id = ? AND rank = 1`, handle, f.prop("toponym").ID)
	must(t, err)
	_, err = db.Exec(`INSERT INTO auto_reconciler_values (entity_id, property_id, rank, value_text, support, against, reason)
		VALUES (?, ?, 2, 'Eboracum', 1, 0, 'kept')`, handle, f.prop("toponym").ID)
	must(t, err)
	f.c.Graph().Drop()

	src, art := f.newSource("Roman itinerary")
	incoming := f.place(src, art, "Eboracum")
	row := rowFor(f.propose(f.graph(src), []graphalign.Fixed{{SubjectID: incoming.ID, HandleID: handle}}), incoming.ID)
	if got := comparisonOutcome(t, row, "toponym"); got != propertycompare.OutcomeAgree {
		t.Fatalf("toponym comparison %s, want agree", got)
	}
}

func datedEvent(f *fixture, src sources.Source, art artifacts.Artifact, eventType string, year int) subjects.Subject {
	return f.datedOn(src, art, eventType, year, 0, 0)
}

// datedOn is an event cited as an event type on a date (m and d 0 for a year).
func (f *fixture) datedOn(src sources.Source, art artifacts.Artifact, eventType string, y, m, d int) subjects.Subject {
	f.t.Helper()
	date := datevalue.Value{Kind: datevalue.KindPoint, Calendar: "gregorian", StartYear: &y}
	if m != 0 {
		date.StartMonth = &m
	}
	if d != 0 {
		date.StartDay = &d
	}
	return f.eventOn(src, art, eventType, &date)
}

// eventOn is an event cited as an event type, on a date when d is not nil.
func (f *fixture) eventOn(src sources.Source, art artifacts.Artifact, eventType string, d *datevalue.Value) subjects.Subject {
	f.t.Helper()
	s := f.bare(src, "event")
	in := []observations.Input{{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", eventType).ID}}
	if d != nil {
		in = append(in, observations.Input{PropertyID: f.prop("date").ID, Date: d})
	}
	f.cite(art, s, in...)
	return s
}

// lifeFixture is a person "James Robins" on the first Source, the subject of a
// dated death and promoted with it, and the same-named person on a second
// Source, the subject of a birth in 1850.
func lifeFixture(t *testing.T, deathYear int) (f *fixture, handle []byte, src sources.Source, person subjects.Subject) {
	t.Helper()
	f = newFixture(t)
	aPerson := f.person(f.source, f.artifact, "James Robins")
	aDeath := datedEvent(f, f.source, f.artifact, "death", deathYear)
	f.participation(f.source, f.artifact, aPerson, aDeath, "subject")
	handle = f.promote(aPerson).Entity.ID
	f.promote(aDeath)

	src, art := f.newSource("Letter")
	person = f.person(src, art, "James Robins")
	bBirth := datedEvent(f, src, art, "birth", 1850)
	f.participation(src, art, person, bBirth, "subject")
	return f, handle, src, person
}

// The slice check: a person who died before the Subject was born is never
// proposed for them, and one who could be them is.
func TestProposeLifespanFromLifeEvents(t *testing.T) {
	t.Run("AC-3.1", func(t *testing.T) {
		f, _, src, person := lifeFixture(t, 1830)
		row := rowFor(f.propose(f.graph(src), nil), person.ID)
		if row.Target != graphalign.TargetNew || row.Reason != graphalign.ReasonNoMatch {
			t.Fatalf("a man who died in 1830 for a man born in 1850: %+v", row)
		}
	})
	t.Run("AC-3.2", func(t *testing.T) {
		f, handle, src, person := lifeFixture(t, 1920)
		row := rowFor(f.propose(f.graph(src), nil), person.ID)
		if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, handle) ||
			row.Assessment != graphalign.AssessWeak || math.Abs(row.Score-2.6) > 1e-9 {
			t.Fatalf("a man who died in 1920 for a man born in 1850: %+v, want the handle at weak 2.6", row)
		}
	})
}

// The lifespan counts in the row's score and comparisons.
func TestProposeComparesTheLifespan(t *testing.T) {
	f, handle, src, person := lifeFixture(t, 1830)
	row := rowFor(f.propose(f.graph(src), []graphalign.Fixed{{SubjectID: person.ID, HandleID: handle}}), person.ID)
	var lifespans int
	for _, c := range row.Comparisons {
		if c.Key != vocabkey.PropertyLifespan {
			continue
		}
		lifespans++
		if c.Result.Outcome != propertycompare.OutcomeConflict || c.Points != -20 || !c.Derived {
			t.Fatalf("lifespan comparison %+v", c)
		}
	}
	if lifespans != 1 {
		t.Fatalf("%d lifespan comparisons in %+v, want 1", lifespans, row.Comparisons)
	}
}

// datedBirth is an event cited as a birth in a year.
func (f *fixture) datedBirth(src sources.Source, art artifacts.Artifact, y int) subjects.Subject {
	f.t.Helper()
	return f.datedOn(src, art, "birth", y, 0, 0)
}

// The loader reads a Source's links from the connect rules: one bridge for
// each registered bridge type, whatever its term.
func TestLayerBridgesCoverEveryConnectRule(t *testing.T) {
	f := newFixture(t)
	ann := f.person(f.source, f.artifact, "Ann Ames")
	bob := f.person(f.source, f.artifact, "Bob Ames")
	birth := f.birth(f.source, f.artifact)
	york := f.place(f.source, f.artifact, "York")
	ontario := f.place(f.source, f.artifact, "Ontario")
	f.participation(f.source, f.artifact, ann, birth, "subject")
	f.locate(f.source, f.artifact, birth, york)
	f.relationship(f.source, f.artifact, ann, bob, "parent")
	f.placeRel(f.source, f.artifact, york, ontario, "part_of")

	bridges, err := promotealign.LayerBridges(f.c, f.graph(f.source))
	must(t, err)
	var got []string
	for _, b := range bridges {
		got = append(got, string(b.KeyA.Type))
	}
	want := connectrules.BridgeTypeKeys()
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("bridge types %v, want %v", got, want)
	}
}

// A birth placed on another event than the handle's birth subtracts from the
// person: a person has one birth.
func TestProposeBirthOnAnotherPersonSubtracts(t *testing.T) {
	f := newFixture(t)
	aAnn := f.person(f.source, f.artifact, "Ann Ames")
	aBirth := f.datedBirth(f.source, f.artifact, 1817)
	f.participation(f.source, f.artifact, aAnn, aBirth, "subject")
	annH := f.promote(aAnn)
	f.promote(aBirth)

	srcC, artC := f.newSource("Register")
	cMary := f.person(srcC, artC, "Mary Ames")
	cBirth := f.datedBirth(srcC, artC, 1817)
	f.participation(srcC, artC, cMary, cBirth, "subject")
	f.promote(cMary)
	cBirthH := f.promote(cBirth)

	srcB, artB := f.newSource("Obituary")
	bAnn := f.person(srcB, artB, "Ann Ames")
	bBirth := f.datedBirth(srcB, artB, 1817)
	f.participation(srcB, artB, bAnn, bBirth, "subject")

	got := f.propose(f.graph(srcB), []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: cBirthH.Entity.ID}})
	row := rowFor(got, bAnn.ID)
	if row.Target == graphalign.TargetHandle && bytes.Equal(row.HandleID, annH.Entity.ID) {
		t.Fatalf("person row %+v filed on the Ann whose birth is another event", row)
	}
}

// A location has no event type on either side: a place reached through a
// fixed birth files on the other record's place.
func TestProposePlaceThroughALocation(t *testing.T) {
	f := newFixture(t)
	aBirth := f.birth(f.source, f.artifact)
	aYork := f.place(f.source, f.artifact, "York")
	f.locate(f.source, f.artifact, aBirth, aYork)
	birthH := f.promote(aBirth)
	yorkH := f.promote(aYork)

	srcB, artB := f.newSource("Register")
	bBirth := f.birth(srcB, artB)
	bYork := f.place(srcB, artB, "York")
	f.locate(srcB, artB, bBirth, bYork)

	got := f.propose(f.graph(srcB), []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: birthH.Entity.ID}})
	row := rowFor(got, bYork.ID)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, yorkH.Entity.ID) || row.Reason != graphalign.ReasonVia {
		t.Fatalf("place row %+v, want York's handle via the birth", row)
	}
}

// "John child of Mary" reads as "Mary parent of John": the walk from Mary
// reaches John's handle through the canonical term.
func TestProposeChildOfReadsAsParentOf(t *testing.T) {
	f := newFixture(t)
	aMary := f.person(f.source, f.artifact, "Mary Smith")
	aJohn := f.person(f.source, f.artifact, "John Smith")
	f.relationship(f.source, f.artifact, aMary, aJohn, "parent")
	maryH := f.promote(aMary).Entity.ID
	johnH := f.promote(aJohn).Entity.ID

	srcB, artB := f.newSource("Baptism")
	bJohn := f.person(srcB, artB, "John Smith")
	bMary := f.person(srcB, artB, "Mary Smith")
	f.relationship(srcB, artB, bJohn, bMary, "child")

	got := f.propose(f.graph(srcB), []graphalign.Fixed{{SubjectID: bMary.ID, HandleID: maryH}})
	row := rowFor(got, bJohn.ID)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, johnH) || row.Reason != graphalign.ReasonVia {
		t.Fatalf("John row %+v, want his handle via Mary", row)
	}
	if row.Via == nil || row.Via.Key.Term != "child" {
		t.Fatalf("via %+v, want term child", row.Via)
	}
}

// The shortlist index is built on the first proposal and dropped by a handle
// commit, so a handle promoted afterwards is found.
func TestProposeSeesAHandlePromotedAfterTheIndexWasBuilt(t *testing.T) {
	f := newFixture(t)
	a := f.person(f.source, f.artifact, "Ann Ames")
	srcB, artB := f.newSource("Obituary")
	b := f.person(srcB, artB, "Ann Ames")

	first := rowFor(f.propose(f.graph(srcB), nil), b.ID)
	if first.Target != graphalign.TargetNew {
		t.Fatalf("with nothing promoted the row is %+v, want a new handle", first)
	}
	known := f.promote(a)
	second := rowFor(f.propose(f.graph(srcB), nil), b.ID)
	if second.Target != graphalign.TargetHandle || !bytes.Equal(second.HandleID, known.Entity.ID) {
		t.Fatalf("after promoting, the row is %+v, want %s", second, known.Entity.Ref)
	}
}

// A warm proposal takes the same number of queries however many people the
// layer holds.
func TestProposeQueryCountStaysFlatWhenWarm(t *testing.T) {
	queries := func(people int) int {
		f := newFixture(t)
		srcB, artB := f.newSource("Obituary")
		for i := 0; i < people; i++ {
			name := fmt.Sprintf("Person %c", 'A'+i)
			p := f.person(f.source, f.artifact, name)
			b := f.birth(f.source, f.artifact)
			f.participation(f.source, f.artifact, p, b, "subject")
			f.promote(p)
			f.promote(b)
			pb := f.person(srcB, artB, name)
			bb := f.birth(srcB, artB)
			f.participation(srcB, artB, pb, bb, "subject")
		}
		_, _, err := promotealign.Propose(f.c, f.graph(srcB), nil)
		must(t, err)
		n := 0
		f.c.Graph().SetQueryCounterForTest(func() { n++ })
		_, _, err = promotealign.Propose(f.c, f.graph(srcB), nil)
		must(t, err)
		return n
	}
	small, large := queries(2), queries(6)
	t.Logf("%d queries per warm proposal", small)
	if small != large {
		t.Fatalf("2 people took %d warm queries, 6 took %d; a read grew with the layer", small, large)
	}
}

// A cold proposal profiles only the handles the shortlist offers, and the ones
// the walk reaches from them: not the life of every person in the catalog.
func TestProposeColdIndexBuildsNoUnshortlistedLife(t *testing.T) {
	f := newFixture(t)
	ann := f.person(f.source, f.artifact, "Ann Ames")
	annBirth := f.datedBirth(f.source, f.artifact, 1850)
	f.participation(f.source, f.artifact, ann, annBirth, "subject")
	annH, annBirthH := f.promote(ann), f.promote(annBirth)
	for i := 1; i <= 6; i++ {
		p := f.person(f.source, f.artifact, fmt.Sprintf("Unrelated %d", i))
		b := f.datedBirth(f.source, f.artifact, 1900+i)
		f.participation(f.source, f.artifact, p, b, "subject")
		f.promote(p)
		f.promote(b)
	}
	srcB, artB := f.newSource("Obituary")
	f.person(srcB, artB, "Ann Ames")

	f.c.Graph().Drop()
	got, err := promotealign.ProfiledRefsForTest(f.c, f.graph(srcB))
	must(t, err)
	want := []string{annH.Entity.Ref, annBirthH.Entity.Ref}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("profiled %v, want only %v", got, want)
	}
}

// A same-name, same-birth record whose sex conflicts still files strongly on
// the links, and is marked contested for the conflict.
func TestProposeMarksAConflictingRowContested(t *testing.T) {
	f := newFixture(t)
	sexed := func(src sources.Source, art artifacts.Artifact, sex string) (subjects.Subject, subjects.Subject) {
		p := f.person(src, art, "Ann Ames")
		f.cite(art, p, observations.Input{
			PropertyID: f.prop("sex_at_birth").ID, ValueTermID: f.term("sex_at_birth", sex).ID,
		})
		b := f.datedBirth(src, art, 1817)
		f.participation(src, art, p, b, "subject")
		return p, b
	}
	pF, bF := sexed(f.source, f.artifact, "female")
	annH := f.promote(pF).Entity.ID
	f.promote(bF)

	src, art := f.newSource("Register")
	pM, bM := sexed(src, art, "male")
	got := f.propose(f.graph(src), nil)
	person, birth := rowFor(got, pM.ID), rowFor(got, bM.ID)
	if person.Target != graphalign.TargetHandle || !bytes.Equal(person.HandleID, annH) ||
		person.Assessment != graphalign.AssessStrong || !person.Flags.Contested {
		t.Fatalf("person row %+v, want the female Ann Ames, strong and contested", person)
	}
	if birth.Target != graphalign.TargetHandle || birth.Assessment != graphalign.AssessStrong || birth.Flags.Contested {
		t.Fatalf("birth row %+v, want strong and not contested", birth)
	}
}

// A name one letter off is not shortlisted, so it is never scored against the
// filed person: the row is New.
func TestProposeDoesNotCompareAnUnshortlistedHandle(t *testing.T) {
	f := newFixture(t)
	f.promote(f.person(f.source, f.artifact, "Robins"))

	src, art := f.newSource("Register")
	robnis := f.person(src, art, "Robnis")
	row := rowFor(f.propose(f.graph(src), nil), robnis.ID)
	if row.Target != graphalign.TargetNew || row.Reason != graphalign.ReasonNoMatch || len(row.Alternatives) != 0 {
		t.Fatalf("Robnis row %+v, want New with no_match and no alternatives", row)
	}
}

// A handle no shortlist key reaches is still loaded, scored and marked
// contested when the researcher fixes the pair, and is not offered unfixed.
func TestProposeHoldsAnUnshortlistedHandle(t *testing.T) {
	tests := []struct {
		name          string
		filedSex      string // "" cites none
		incomingSex   string
		wantOutcome   propertycompare.Outcome
		wantContested bool
	}{
		{name: "no sex", wantOutcome: propertycompare.OutcomeUnknown},
		{name: "same sex", filedSex: "female", incomingSex: "female", wantOutcome: propertycompare.OutcomeAgree},
		{name: "conflicting sex", filedSex: "female", incomingSex: "male", wantOutcome: propertycompare.OutcomeConflict, wantContested: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			sexed := func(src sources.Source, art artifacts.Artifact, name, sex string) subjects.Subject {
				p := f.person(src, art, name)
				if sex != "" {
					f.cite(art, p, observations.Input{
						PropertyID: f.prop("sex_at_birth").ID, ValueTermID: f.term("sex_at_birth", sex).ID,
					})
				}
				return p
			}
			handle := f.promote(sexed(f.source, f.artifact, "Robins", tt.filedSex)).Entity.ID

			src, art := f.newSource("Register")
			robnis := sexed(src, art, "Robnis", tt.incomingSex)
			graph := f.graph(src)

			unfixed := rowFor(f.propose(graph, nil), robnis.ID)
			if unfixed.Target == graphalign.TargetHandle && bytes.Equal(unfixed.HandleID, handle) {
				t.Fatalf("unfixed row %+v targets the unshortlisted handle", unfixed)
			}
			for _, alt := range unfixed.Alternatives {
				if bytes.Equal(alt.HandleID, handle) {
					t.Fatalf("unfixed row %+v lists the unshortlisted handle", unfixed)
				}
			}

			row := rowFor(f.propose(graph, []graphalign.Fixed{{SubjectID: robnis.ID, HandleID: handle}}), robnis.ID)
			if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, handle) ||
				row.Reason != graphalign.ReasonDecided || row.Flags.ConflictWithFixed {
				t.Fatalf("fixed row %+v, want the handle, decided, no conflict with fixed", row)
			}
			comparisonOutcome(t, row, "name")
			if got := comparisonOutcome(t, row, "sex_at_birth"); got != tt.wantOutcome {
				t.Fatalf("sex_at_birth outcome %q, want %q", got, tt.wantOutcome)
			}
			if row.Flags.Contested != tt.wantContested {
				t.Fatalf("contested %v, want %v", row.Flags.Contested, tt.wantContested)
			}
		})
	}
}

// The link set merges a recorded link and an implied one of the same ends and
// keys into one, and merges the links of two marriages of one couple.
func TestLinkSet(t *testing.T) {
	terms := segmentcomparetest.Terms()
	rules := matchregistry.Default().Kinship
	person := func(n byte) []byte { return []byte{0x10 + n} }
	event := func(n byte) []byte { return []byte{0x20 + n} }
	bridgeID := func(n byte) []byte { return []byte{0x30 + n} }
	part := func(p, e, b byte) promotealign.SideLink {
		ka, kb := segmentcompare.Keys(segmentcomparetest.Bridge(t, vocabkey.BridgeParticipation), "subject", "marriage", terms)
		return promotealign.SideLink{Edge: kinship.Edge{A: person(p), B: event(e), KeyA: ka, KeyB: kb, Associations: [][]byte{bridgeID(b)}}}
	}
	spouse := func(p, q, b byte) promotealign.SideLink {
		ka, kb := segmentcompare.Keys(segmentcomparetest.Bridge(t, vocabkey.BridgeRelationship), "spouse", "", terms)
		return promotealign.SideLink{Edge: kinship.Edge{A: person(p), B: person(q), KeyA: ka, KeyB: kb, Associations: [][]byte{bridgeID(b)}}}
	}
	flip := func(l promotealign.SideLink) promotealign.SideLink {
		l.A, l.B, l.KeyA, l.KeyB = l.B, l.A, l.KeyB, l.KeyA
		return l
	}
	t.Run("the same recorded link from each end is one", func(t *testing.T) {
		got := promotealign.LinkSetForTest([]promotealign.SideLink{spouse(1, 2, 1), flip(spouse(1, 2, 2))}, terms, rules)
		if len(got) != 1 || len(got[0].Associations) != 2 || got[0].OnlyImplied || len(got[0].Shapes) != 0 {
			t.Fatalf("links %+v, want one recorded link with 2 associations", got)
		}
	})
	t.Run("a recorded spouse link and a marriage are one link", func(t *testing.T) {
		got := promotealign.LinkSetForTest([]promotealign.SideLink{spouse(1, 2, 1), part(1, 1, 2), part(2, 1, 3)}, terms, rules)
		var spouses []promotealign.SideLink
		for _, l := range got {
			if l.KeyA.Type == vocabkey.BridgeRelationship {
				spouses = append(spouses, l)
			}
		}
		if len(spouses) != 1 || len(spouses[0].Associations) != 3 || len(spouses[0].Shapes) != 1 || spouses[0].OnlyImplied {
			t.Fatalf("spouse links %+v, want one with 3 associations, one shape, recorded", spouses)
		}
	})
	t.Run("two marriages of one couple are one implied link", func(t *testing.T) {
		got := promotealign.LinkSetForTest([]promotealign.SideLink{part(1, 1, 1), part(2, 1, 2), part(1, 2, 3), part(2, 2, 4)}, terms, rules)
		var spouses []promotealign.SideLink
		for _, l := range got {
			if l.KeyA.Type == vocabkey.BridgeRelationship {
				spouses = append(spouses, l)
			}
		}
		if len(spouses) != 1 || len(spouses[0].Associations) != 4 || len(spouses[0].Shapes) != 2 || !spouses[0].OnlyImplied {
			t.Fatalf("spouse links %+v, want one implied link with 4 associations and 2 shapes", spouses)
		}
	})
	t.Run("the same implied link added twice is one shape", func(t *testing.T) {
		imp := spouse(1, 2, 1)
		imp.Recording = kinship.Recording{Shapes: []kinship.Shape{{Pattern: kinship.PatternMarriageSpouses, Pivot: event(1)}}, OnlyImplied: true}
		got := promotealign.LinkSetForTest([]promotealign.SideLink{imp, imp}, terms, rules)
		if len(got) != 1 || len(got[0].Shapes) != 1 || !got[0].OnlyImplied {
			t.Fatalf("links %+v, want one implied link with one shape", got)
		}
	})
	t.Run("an implied link does not alias the input", func(t *testing.T) {
		in := []promotealign.SideLink{part(1, 1, 1), part(2, 1, 2)}
		got := promotealign.LinkSetForTest(in, terms, rules)
		var imp promotealign.SideLink
		for _, l := range got {
			if l.KeyA.Type == vocabkey.BridgeRelationship {
				imp = l
			}
		}
		if imp.A == nil {
			t.Fatalf("no implied link in %+v", got)
		}
		want := bytes.Clone(imp.A)
		in[0].A[0], in[1].A[0] = 0xff, 0xff
		if !bytes.Equal(imp.A, want) {
			t.Fatalf("implied A %x changed with the input, want %x", imp.A, want)
		}
	})
}

// A Source that records a spouse relationship and a marriage has one spouse
// bridge, with every association behind it; a marriage alone has an implied one.
func TestLayerBridgesMergeKinshipShapes(t *testing.T) {
	spouses := func(bs []graphalign.Bridge) []graphalign.Bridge {
		var out []graphalign.Bridge
		for _, b := range bs {
			if b.KeyA.Type == vocabkey.BridgeRelationship && b.KeyA.Term == "spouse" {
				out = append(out, b)
			}
		}
		return out
	}
	build := func(withRelationship bool) []graphalign.Bridge {
		f := newFixture(t)
		ann := f.person(f.source, f.artifact, "Ann Ames")
		bob := f.person(f.source, f.artifact, "Bob Ames")
		marriage := f.eventOn(f.source, f.artifact, "marriage", nil)
		f.participation(f.source, f.artifact, ann, marriage, "subject")
		f.participation(f.source, f.artifact, bob, marriage, "subject")
		if withRelationship {
			f.relationship(f.source, f.artifact, ann, bob, "spouse")
		}
		bridges, err := promotealign.LayerBridges(f.c, f.graph(f.source))
		must(t, err)
		return bridges
	}
	both := spouses(build(true))
	if len(both) != 1 || len(both[0].Associations) != 3 || len(both[0].Shapes) != 1 || both[0].OnlyImplied {
		t.Fatalf("recorded and implied: %+v, want one recorded bridge with 3 associations and one shape", both)
	}
	only := spouses(build(false))
	if len(only) != 1 || len(only[0].Associations) != 2 || !only[0].OnlyImplied {
		t.Fatalf("marriage only: %+v, want one implied bridge with 2 associations", only)
	}
}
