package promotealign_test

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"reflect"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/canonlinks"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/observations/observationstest"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/valuecodec"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

var (
	userID = evidencelayertest.UserID
	must   = evidencelayertest.Must
)

const locator = evidencelayertest.Locator

// fixture is the shared Source-building records, with the fields the tests read
// by name.
type fixture struct {
	*evidencelayertest.Records
	t        testing.TB
	c        *database.Catalog
	source   sources.Source
	artifact artifacts.Artifact
}

func newFixture(t testing.TB) *fixture {
	t.Helper()
	r := evidencelayertest.New(t)
	return &fixture{Records: r, t: t, c: r.C, source: r.Source, artifact: r.Artifact}
}

func (f *fixture) promote(s subjects.Subject) promote.Result {
	f.t.Helper()
	return promotetest.Save(f.t, f.c, userID, promote.Input{SubjectID: s.ID})
}

func (f *fixture) propose(graphID []byte, fixed []graphalign.Fixed) graphalign.Proposal {
	f.t.Helper()
	p, _, err := promotealign.Propose(f.c, graphID, fixed, nil)
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
	aPerson := f.Person(f.source, f.artifact, "Gracie")
	per := f.promote(aPerson)
	srcB, artB := f.NewSource("Obituary")
	bPerson := f.Person(srcB, artB, "Gracie")

	f.propose(f.Graph(srcB), []graphalign.Fixed{{SubjectID: bPerson.ID, HandleID: per.Entity.ID}})
	if _, ok, _ := promotealign.MatchIndex.Get(f.c.Graph()); ok {
		t.Error("Propose built the match index with no Subject to seed")
	}
	f.propose(f.Graph(srcB), nil)
	if _, ok, _ := promotealign.MatchIndex.Get(f.c.Graph()); !ok {
		t.Error("Propose with an unfixed Subject did not build the match index")
	}
}

// Seed from an already-promoted birth: the unpromoted person on a new Source
// should accept the matching Gracie handle through the participation edge.
func TestProposeSeedFromPromotedNeighbor(t *testing.T) {
	f := newFixture(t)

	// Canon: Gracie ↔ birth filed on Source A.
	aPerson := f.Person(f.source, f.artifact, "Gracie")
	aBirth := f.Birth(f.source, f.artifact)
	f.Participation(f.source, f.artifact, aPerson, aBirth, "subject")
	per := f.promote(aPerson)
	evt := f.promote(aBirth)

	// Layer B: matching person + birth; fix the birth to the known handle.
	srcB, artB := f.NewSource("Obituary")
	bPerson := f.Person(srcB, artB, "Gracie")
	bBirth := f.Birth(srcB, artB)
	f.Participation(srcB, artB, bPerson, bBirth, "subject")

	got := f.propose(f.Graph(srcB), []graphalign.Fixed{{
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

	york := f.Place(f.source, f.artifact, "York")
	ontario := f.Place(f.source, f.artifact, "Ontario")
	f.PlaceRel(f.source, f.artifact, york, ontario, "part_of")
	yorkH := f.promote(york)
	ontarioH := f.promote(ontario)

	srcB, artB := f.NewSource("Gazetteer")
	york2 := f.Place(srcB, artB, "York")
	ontario2 := f.Place(srcB, artB, "Ontario")
	f.PlaceRel(srcB, artB, york2, ontario2, "part_of")

	got := f.propose(f.Graph(srcB), []graphalign.Fixed{{
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

	aPerson := f.Person(f.source, f.artifact, "Doug")
	aEvent := f.Bare(f.source, "event")
	f.Cite(f.artifact, aEvent,
		observations.Input{PropertyID: f.Prop("event_type").ID, ValueTermID: f.Term("event_type", "death").ID},
		observations.Input{PropertyID: f.Prop("start_date").ID, Date: earlier},
	)
	f.Participation(f.source, f.artifact, aPerson, aEvent, "subject")
	per := f.promote(aPerson)
	evt := f.promote(aEvent)

	srcB, artB := f.NewSource("Letter")
	bPerson := f.Person(srcB, artB, "Doug")
	bEvent := f.Bare(srcB, "event")
	f.Cite(artB, bEvent,
		observations.Input{PropertyID: f.Prop("event_type").ID, ValueTermID: f.Term("event_type", "death").ID},
		observations.Input{PropertyID: f.Prop("start_date").ID, Date: later},
	)
	f.Participation(srcB, artB, bPerson, bEvent, "subject")

	got := f.propose(f.Graph(srcB), []graphalign.Fixed{{SubjectID: bPerson.ID, HandleID: per.Entity.ID}})
	row := rowFor(got, bEvent.ID)
	if !bytes.Equal(row.HandleID, evt.Entity.ID) {
		t.Fatalf("event %+v, want %s", row, evt.Entity.Ref)
	}
}

func TestProposeInvalidGraph(t *testing.T) {
	f := newFixture(t)
	_, _, err := promotealign.Propose(f.c, make([]byte, 16), nil, nil)
	if err != promote.ErrInvalid {
		t.Fatalf("err %v, want promote.ErrInvalid", err)
	}
}

// A Source with two graphs proposes one graph's Subjects at a time.
func TestProposeReadsOneGraph(t *testing.T) {
	f := newFixture(t)
	census, censusArt := f.NewSource("Census")
	def, jones, page2 := artifactstest.Split(t, f.c, userID, census.ID, "Jones")
	smith := subjectstest.On(t, f.c, userID, def.ID, "person", "")
	f.Cite(censusArt, smith, observations.Input{PropertyID: f.Prop("name").ID, Name: namevaluetest.Western("John Smith")})
	mary := subjectstest.On(t, f.c, userID, jones.ID, "person", "")
	f.Cite(page2, mary, observations.Input{PropertyID: f.Prop("name").ID, Name: namevaluetest.Western("Mary Jones")})

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
			if _, _, err := promotealign.Propose(f.c, id, nil, nil); !errors.Is(err, promote.ErrInvalid) {
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
			per := f.promote(f.Person(f.source, f.artifact, "Gracie"))
			srcB, artB := f.NewSource("Obituary")
			again := f.Person(srcB, artB, "Gracie")
			fixed := tt.fixed(per.Entity.ID)
			fixed.SubjectID = again.ID
			p, _, err := promotealign.Propose(f.c, f.Graph(srcB), []graphalign.Fixed{fixed}, nil)
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
		srcB, artB := f.NewSource("Obituary")
		for i := 0; i < people; i++ {
			name := fmt.Sprintf("Person %c", 'A'+i)
			p := f.Person(f.source, f.artifact, name)
			b := f.Birth(f.source, f.artifact)
			f.Participation(f.source, f.artifact, p, b, "subject")
			f.promote(p)
			f.promote(b)
			pb := f.Person(srcB, artB, name)
			bb := f.Birth(srcB, artB)
			f.Participation(srcB, artB, pb, bb, "subject")
		}
		n := 0
		f.c.Graph().SetQueryCounterForTest(func() { n++ })
		_, _, err := promotealign.Propose(f.c, f.Graph(srcB), nil, nil)
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
	city := f.Place(f.source, f.artifact, "Gumptiontown")
	state := f.Place(f.source, f.artifact, "Provenance")
	country := f.Place(f.source, f.artifact, "Arcadia")
	f.PlaceRel(f.source, f.artifact, city, state, "part_of")
	f.PlaceRel(f.source, f.artifact, state, country, "part_of")
	cityH := f.promote(city)
	stateH := f.promote(state)
	countryH := f.promote(country)

	srcB, artB := f.NewSource("Gazetteer")
	city2 := f.Place(srcB, artB, "Gumptiontown")
	state2 := f.Place(srcB, artB, "Provenance")
	country2 := f.Place(srcB, artB, "Arcadia")
	f.PlaceRel(srcB, artB, city2, state2, "part_of")
	f.PlaceRel(srcB, artB, state2, country2, "part_of")

	prop := f.propose(f.Graph(srcB), nil)
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
	york := f.promote(f.Place(f.source, f.artifact, "York"))
	f.promote(f.Place(f.source, f.artifact, "Leeds"))

	srcB, artB := f.NewSource("Gazetteer")
	again := f.Place(srcB, artB, "York")
	row := rowFor(f.propose(f.Graph(srcB), nil), again.ID)
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
			known := f.promote(f.Person(f.source, f.artifact, tc.full))
			srcB, artB := f.NewSource("Census")
			again := f.Person(srcB, artB, tc.short)
			row := rowFor(f.propose(f.Graph(srcB), nil), again.ID)
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
	f.promote(f.Place(f.source, f.artifact, "Provenanced"))
	f.promote(f.Place(f.source, f.artifact, "Leeds"))

	srcB, artB := f.NewSource("Gazetteer")
	again := f.Place(srcB, artB, "Provenance")
	row := rowFor(f.propose(f.Graph(srcB), nil), again.ID)
	if row.Target != "" || row.Reason != graphalign.ReasonUnplaced {
		t.Fatalf("provenance row %+v, want no target and unplaced", row)
	}
}

// A directed term is matched from the same end: the fixed place's part (not
// its whole, though both are named York) is what the layer's part matches.
func TestProposePartOfMatchesTheSameEnd(t *testing.T) {
	f := newFixture(t)
	part := f.Place(f.source, f.artifact, "York")
	middle := f.Place(f.source, f.artifact, "Ontario")
	whole := f.Place(f.source, f.artifact, "York")
	f.PlaceRel(f.source, f.artifact, part, middle, "part_of")
	f.PlaceRel(f.source, f.artifact, middle, whole, "part_of")
	partH := f.promote(part)
	middleH := f.promote(middle)
	f.promote(whole)
	// Other places, so "York" is not a common name and the edge decides.
	for _, name := range []string{"Bath", "Leeds", "Hull", "Derby", "Ely", "Wells"} {
		f.promote(f.Place(f.source, f.artifact, name))
	}

	srcB, artB := f.NewSource("Gazetteer")
	york := f.Place(srcB, artB, "York")
	ontario := f.Place(srcB, artB, "Ontario")
	f.PlaceRel(srcB, artB, york, ontario, "part_of")

	row := rowFor(f.propose(f.Graph(srcB), []graphalign.Fixed{{SubjectID: ontario.ID, HandleID: middleH.Entity.ID}}), york.ID)
	if row.Reason != graphalign.ReasonVia || !bytes.Equal(row.HandleID, partH.Entity.ID) {
		t.Fatalf("york row %+v, want the part %s via Ontario", row, partH.Entity.Ref)
	}
}

// "Mary parent of John" on one Source and "John child of Mary" on another are
// one relationship: the walk from Mary reaches John's handle either way.
func TestProposeInverseKinshipTermsCorrespond(t *testing.T) {
	f := newFixture(t)
	mary := f.Person(f.source, f.artifact, "Mary Smith")
	john := f.Person(f.source, f.artifact, "John Smith")
	f.Relationship(f.source, f.artifact, mary, john, "parent")
	hMary := f.promote(mary).Entity.ID
	hJohn := f.promote(john).Entity.ID

	src2, art2 := f.NewSource("Baptism")
	son := f.Person(src2, art2, "John Smith")
	mother := f.Person(src2, art2, "Mary Smith")
	f.Relationship(src2, art2, son, mother, "child")

	p := f.propose(f.Graph(src2), []graphalign.Fixed{{SubjectID: mother.ID, HandleID: hMary}})
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
	mary := f.Person(f.source, f.artifact, "Mary Smith")
	john := f.Person(f.source, f.artifact, "John Smith")
	f.Relationship(f.source, f.artifact, mary, john, "parent")
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
			src, art := f.NewSource(tt.name)
			var parent, child subjects.Subject
			var fixed graphalign.Fixed
			var want []byte
			if tt.anchorIsParent {
				parent = f.Person(src, art, "Mary Smith")
				child = f.SexOnly(src, art, "male")
				if tt.namedOther {
					child = f.Person(src, art, "John Smith")
				}
				fixed = graphalign.Fixed{SubjectID: parent.ID, HandleID: hMary}
				want = hJohn
			} else {
				parent = f.SexOnly(src, art, "female")
				child = f.Person(src, art, "John Smith")
				fixed = graphalign.Fixed{SubjectID: child.ID, HandleID: hJohn}
				want = hMary
			}
			f.Relationship(src, art, parent, child, "parent")
			p := f.propose(f.Graph(src), []graphalign.Fixed{fixed})
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
			if r.Target != "" || r.Reason != graphalign.ReasonUnplaced {
				t.Fatalf("target=%s reason=%s, want no target (under the accept bar)", r.Target, r.Reason)
			}
			if len(r.Alternatives) == 0 || !bytes.Equal(r.Alternatives[0].HandleID, want) {
				t.Fatalf("alternatives %+v, want the other end's handle first, via the anchor", r.Alternatives)
			}
		})
	}
}

// A place with several filed events still proposes one row per layer
// subject, and the place lands on its handle.
func TestPlaceWithManyEventsProposalStaysOnTheLayer(t *testing.T) {
	f := newFixture(t)
	city := f.Place(f.source, f.artifact, "Gumptiontown")
	cityH := f.promote(city)
	for i := 0; i < 3; i++ {
		person := f.Person(f.source, f.artifact, fmt.Sprintf("Neighbor %d", i))
		birth := f.Birth(f.source, f.artifact)
		f.Participation(f.source, f.artifact, person, birth, "subject")
		f.Locate(f.source, f.artifact, birth, city)
		f.promote(person)
		f.promote(birth)
	}

	srcB, artB := f.NewSource("Gazetteer")
	city2 := f.Place(srcB, artB, "Gumptiontown")
	birth2 := f.Birth(srcB, artB)
	f.Locate(srcB, artB, birth2, city2)

	got := f.propose(f.Graph(srcB), nil)
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
		s := f.Person(src, art, "John Smith")
		f.Cite(art, s, observations.Input{
			PropertyID: f.Prop("sex_at_birth").ID, ValueTermID: f.Term("sex_at_birth", sex).ID,
		})
		return s
	}
	handle := f.promote(named(f.source, f.artifact, "male")).Entity.ID

	src, art := f.NewSource("Register")
	unknown := named(src, art, "unknown")
	r := rowFor(f.propose(f.Graph(src), nil), unknown.ID)
	if r.Target != graphalign.TargetHandle || !bytes.Equal(r.HandleID, handle) {
		t.Fatalf("unknown sex: target %s score %.2f, want the male John Smith", r.Target, r.Score)
	}

	src2, art2 := f.NewSource("Census")
	female := named(src2, art2, "female")
	if r := rowFor(f.propose(f.Graph(src2), nil), female.ID); r.Target == graphalign.TargetHandle && bytes.Equal(r.HandleID, handle) {
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
		s := f.Person(src, art, name)
		f.Cite(art, s, observations.Input{
			PropertyID: f.Prop("gender").ID, ValueTermID: f.Term("gender", "man").ID,
		})
		return s
	}
	james := withGender(f.source, f.artifact, "James Kenneth Robins")
	handle := f.promote(james).Entity.ID
	birthSrc, birthArt := f.NewSource("Birth Announcement")
	f.join(withGender(birthSrc, birthArt, "Jake Robins"), handle)

	noteSrc, noteArt := f.NewSource("Note")
	probe := withGender(noteSrc, noteArt, "Jake Robins")
	row := rowFor(f.propose(f.Graph(noteSrc), nil), probe.ID)
	cfg := matchregistry.Default().AlignConfig()
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, handle) ||
		row.Assessment != graphalign.AssessWeak || row.Score < cfg.Alignment.Bands.Weak || row.Score >= cfg.Alignment.Bands.Medium ||
		math.Abs(row.Score-(2.5*0.75+0.6)) > 1e-9 {
		t.Fatalf("Jake Robins: target %s assessment %s score %.4f", row.Target, row.Assessment, row.Score)
	}
}

// A given name that starts another's is offered the record and is a weak match
// with the same surname, whichever side is filed. No nickname set holds ros, so
// the shortlist's prefix keys and the comparer's prefix rule carry it (S11-45).
func TestProposeAPrefixNameIsAWeakMatch(t *testing.T) {
	for _, tc := range []struct{ filed, other, probe string }{
		{"Rosalind Long", "Mary Long", "Ros Long"},
		{"Ros Long", "Mary Long", "Rosalind Long"},
	} {
		t.Run(tc.filed+" and "+tc.probe, func(t *testing.T) {
			f := newFixture(t)
			known := f.promote(f.Person(f.source, f.artifact, tc.filed))
			f.promote(f.Person(f.source, f.artifact, tc.other))
			srcB, artB := f.NewSource("Census")
			again := f.Person(srcB, artB, tc.probe)
			row := rowFor(f.propose(f.Graph(srcB), nil), again.ID)
			bands := matchregistry.Default().AlignConfig().Alignment.Bands
			if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, known.Entity.ID) ||
				row.Assessment != graphalign.AssessWeak || row.Score < bands.Weak || row.Score >= bands.Medium {
				t.Fatalf("%s: target %s assessment %s score %.4f", tc.probe, row.Target, row.Assessment, row.Score)
			}
		})
	}
}

// namedMembers files n people called name, each on a Source of its own titled
// "Census 1851", "Census 1861", ...: the first on a new handle, the rest joined
// to it. It returns the handle and the members in filing order.
func (f *fixture) namedMembers(n int, name string) (handle []byte, members []subjects.Subject) {
	f.t.Helper()
	for i := 0; i < n; i++ {
		src, art := f.NewSource(fmt.Sprintf("Census %d", 1851+10*i))
		p := f.Person(src, art, name)
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

func TestProposeJudgesAgainstTheMajority(t *testing.T) {
	t.Run("a stray name", func(t *testing.T) {
		f := newFixture(t)
		handle := f.promote(f.Person(f.source, f.artifact, "Mary Smith")).Entity.ID
		blob, err := valuecodec.MarshalName(*namevaluetest.Western("James Robins"))
		must(t, err)
		db, err := f.c.DB()
		must(t, err)
		_, err = db.Exec(`UPDATE auto_reconciler_values SET support = 5 WHERE entity_id = ? AND property_id = ? AND rank = 1`, handle, f.Prop("name").ID)
		must(t, err)
		_, err = db.Exec(`INSERT INTO auto_reconciler_values (entity_id, property_id, rank, value_name, support, against, reason)
			VALUES (?, ?, 2, ?, 1, 0, 'outvoted')`, handle, f.Prop("name").ID, blob)
		must(t, err)
		f.c.Graph().Drop()

		src, art := f.NewSource("Register")
		incoming := f.Person(src, art, "James Robins")
		row := rowFor(f.propose(f.Graph(src), []graphalign.Fixed{{SubjectID: incoming.ID, HandleID: handle}}), incoming.ID)
		if row.Score >= 0 {
			t.Fatalf("score %.2f, want the name conflict to count against the pair", row.Score)
		}
	})

	t.Run("a negative Observation", func(t *testing.T) {
		f := newFixture(t)
		sex := func(src sources.Source, art artifacts.Artifact, key string) subjects.Subject {
			s := f.Person(src, art, "Mary Smith")
			f.Cite(art, s, observations.Input{PropertyID: f.Prop("sex_at_birth").ID, ValueTermID: f.Term("sex_at_birth", key).ID})
			return s
		}
		handle := f.promote(sex(f.source, f.artifact, "male")).Entity.ID
		db, err := f.c.DB()
		must(t, err)
		_, err = db.Exec(`UPDATE auto_reconciler_values SET support = 3 WHERE entity_id = ? AND property_id = ?`, handle, f.Prop("sex_at_birth").ID)
		must(t, err)
		f.c.Graph().Drop()

		src, art := f.NewSource("Census")
		incoming := f.Person(src, art, "Mary Smith")
		f.Cite(art, incoming, observations.Input{
			PropertyID: f.Prop("sex_at_birth").ID, ValueTermID: f.Term("sex_at_birth", "male").ID,
			Polarity: observations.PolarityNegative,
		})
		row := rowFor(f.propose(f.Graph(src), []graphalign.Fixed{{SubjectID: incoming.ID, HandleID: handle}}), incoming.ID)
		if row.Score >= 3 {
			t.Fatalf("score %.2f, want the sex conflict to count against the name's agreement", row.Score)
		}
	})
}

// Likeness has no weight: a person proposed with likenesses on the Source and
// on the handle gets the band and score it gets without them.
func TestProposeIgnoresLikenesses(t *testing.T) {
	score := func(t *testing.T, withLikeness bool) graphalign.Row {
		f := newFixture(t)
		a := f.Person(f.source, f.artifact, "John Smith")
		if withLikeness {
			observationstest.SaveLikeness(t, f.c, userID, f.source.ID, a.ID, 40)
		}
		f.promote(a)
		src, art := f.NewSource("Obituary")
		incoming := f.Person(src, art, "John Smith")
		if withLikeness {
			observationstest.SaveLikeness(t, f.c, userID, src.ID, incoming.ID, 200)
		}
		_ = art
		return rowFor(f.propose(f.Graph(src), nil), incoming.ID)
	}
	with, without := score(t, true), score(t, false)
	if with.Target != without.Target || with.Assessment != without.Assessment || with.Score != without.Score || with.Reason != without.Reason {
		t.Fatalf("with likenesses %+v\nwithout %+v", with, without)
	}
}

// A Place handle keeps every toponym it was read as; the minority one still
// agrees, because a multiple-cardinality Property weighs no value down.
func TestProposeMultipleToponyms(t *testing.T) {
	f := newFixture(t)
	handle := f.promote(f.Place(f.source, f.artifact, "York")).Entity.ID
	db, err := f.c.DB()
	must(t, err)
	_, err = db.Exec(`UPDATE auto_reconciler_values SET support = 5 WHERE entity_id = ? AND property_id = ? AND rank = 1`, handle, f.Prop("toponym").ID)
	must(t, err)
	_, err = db.Exec(`INSERT INTO auto_reconciler_values (entity_id, property_id, rank, value_text, support, against, reason)
		VALUES (?, ?, 2, 'Eboracum', 1, 0, 'kept')`, handle, f.Prop("toponym").ID)
	must(t, err)
	f.c.Graph().Drop()

	src, art := f.NewSource("Roman itinerary")
	incoming := f.Place(src, art, "Eboracum")
	row := rowFor(f.propose(f.Graph(src), []graphalign.Fixed{{SubjectID: incoming.ID, HandleID: handle}}), incoming.ID)
	if row.Score <= 0 {
		t.Fatalf("score %.2f, want the minority toponym to agree", row.Score)
	}
}

func datedEvent(f *fixture, src sources.Source, art artifacts.Artifact, eventType string, year int) subjects.Subject {
	return f.DatedOn(src, art, eventType, year, 0, 0)
}

// lifeFixture is a person "James Robins" on the first Source, the subject of a
// dated death and promoted with it, and the same-named person on a second
// Source, the subject of a birth in 1850.
func lifeFixture(t *testing.T, deathYear int) (f *fixture, handle []byte, src sources.Source, person subjects.Subject) {
	t.Helper()
	f = newFixture(t)
	aPerson := f.Person(f.source, f.artifact, "James Robins")
	aDeath := datedEvent(f, f.source, f.artifact, "death", deathYear)
	f.Participation(f.source, f.artifact, aPerson, aDeath, "subject")
	handle = f.promote(aPerson).Entity.ID
	f.promote(aDeath)

	src, art := f.NewSource("Letter")
	person = f.Person(src, art, "James Robins")
	bBirth := datedEvent(f, src, art, "birth", 1850)
	f.Participation(src, art, person, bBirth, "subject")
	return f, handle, src, person
}

// The slice check: a person who died before the Subject was born is never
// proposed for them, and one who could be them is.
func TestProposeLifespanFromLifeEvents(t *testing.T) {
	t.Run("AC-3.1", func(t *testing.T) {
		f, _, src, person := lifeFixture(t, 1830)
		row := rowFor(f.propose(f.Graph(src), nil), person.ID)
		if row.Target != "" || row.Reason != graphalign.ReasonUnplaced {
			t.Fatalf("a man who died in 1830 for a man born in 1850: %+v", row)
		}
	})
	t.Run("AC-3.2", func(t *testing.T) {
		f, handle, src, person := lifeFixture(t, 1920)
		row := rowFor(f.propose(f.Graph(src), nil), person.ID)
		if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, handle) ||
			row.Assessment != graphalign.AssessWeak || math.Abs(row.Score-2.6) > 1e-9 {
			t.Fatalf("a man who died in 1920 for a man born in 1850: %+v, want the handle at weak 2.6", row)
		}
	})
}

// The lifespan counts in the row's score: a man who died before the handle
// was born takes its conflict from the score.
func TestProposeComparesTheLifespan(t *testing.T) {
	f, handle, src, person := lifeFixture(t, 1830)
	row := rowFor(f.propose(f.Graph(src), []graphalign.Fixed{{SubjectID: person.ID, HandleID: handle}}), person.ID)
	if row.Score > -10 {
		t.Fatalf("score %.2f, want the lifespan's -20 conflict counted", row.Score)
	}
}

// A birth placed on another event than the handle's birth subtracts from the
// person: a person has one birth.
func TestProposeBirthOnAnotherPersonSubtracts(t *testing.T) {
	f := newFixture(t)
	aAnn := f.Person(f.source, f.artifact, "Ann Ames")
	aBirth := f.DatedBirth(f.source, f.artifact, 1817)
	f.Participation(f.source, f.artifact, aAnn, aBirth, "subject")
	annH := f.promote(aAnn)
	f.promote(aBirth)

	srcC, artC := f.NewSource("Register")
	cMary := f.Person(srcC, artC, "Mary Ames")
	cBirth := f.DatedBirth(srcC, artC, 1817)
	f.Participation(srcC, artC, cMary, cBirth, "subject")
	f.promote(cMary)
	cBirthH := f.promote(cBirth)

	srcB, artB := f.NewSource("Obituary")
	bAnn := f.Person(srcB, artB, "Ann Ames")
	bBirth := f.DatedBirth(srcB, artB, 1817)
	f.Participation(srcB, artB, bAnn, bBirth, "subject")

	got := f.propose(f.Graph(srcB), []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: cBirthH.Entity.ID}})
	row := rowFor(got, bAnn.ID)
	if row.Target == graphalign.TargetHandle && bytes.Equal(row.HandleID, annH.Entity.ID) {
		t.Fatalf("person row %+v filed on the Ann whose birth is another event", row)
	}
}

// A location has no event type on either side: a place reached through a
// fixed birth files on the other record's place.
func TestProposePlaceThroughALocation(t *testing.T) {
	f := newFixture(t)
	aBirth := f.Birth(f.source, f.artifact)
	aYork := f.Place(f.source, f.artifact, "York")
	f.Locate(f.source, f.artifact, aBirth, aYork)
	birthH := f.promote(aBirth)
	yorkH := f.promote(aYork)

	srcB, artB := f.NewSource("Register")
	bBirth := f.Birth(srcB, artB)
	bYork := f.Place(srcB, artB, "York")
	f.Locate(srcB, artB, bBirth, bYork)

	got := f.propose(f.Graph(srcB), []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: birthH.Entity.ID}})
	row := rowFor(got, bYork.ID)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, yorkH.Entity.ID) || row.Reason != graphalign.ReasonVia {
		t.Fatalf("place row %+v, want York's handle via the birth", row)
	}
}

// "John child of Mary" reads as "Mary parent of John": the walk from Mary
// reaches John's handle through the canonical term.
func TestProposeChildOfReadsAsParentOf(t *testing.T) {
	f := newFixture(t)
	aMary := f.Person(f.source, f.artifact, "Mary Smith")
	aJohn := f.Person(f.source, f.artifact, "John Smith")
	f.Relationship(f.source, f.artifact, aMary, aJohn, "parent")
	maryH := f.promote(aMary).Entity.ID
	johnH := f.promote(aJohn).Entity.ID

	srcB, artB := f.NewSource("Baptism")
	bJohn := f.Person(srcB, artB, "John Smith")
	bMary := f.Person(srcB, artB, "Mary Smith")
	f.Relationship(srcB, artB, bJohn, bMary, "child")

	got := f.propose(f.Graph(srcB), []graphalign.Fixed{{SubjectID: bMary.ID, HandleID: maryH}})
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
	a := f.Person(f.source, f.artifact, "Ann Ames")
	srcB, artB := f.NewSource("Obituary")
	b := f.Person(srcB, artB, "Ann Ames")

	first := rowFor(f.propose(f.Graph(srcB), nil), b.ID)
	if first.Target != "" {
		t.Fatalf("with nothing promoted the row is %+v, want no target", first)
	}
	known := f.promote(a)
	second := rowFor(f.propose(f.Graph(srcB), nil), b.ID)
	if second.Target != graphalign.TargetHandle || !bytes.Equal(second.HandleID, known.Entity.ID) {
		t.Fatalf("after promoting, the row is %+v, want %s", second, known.Entity.Ref)
	}
}

// A warm proposal takes the same number of queries however many people the
// layer holds.
func TestProposeQueryCountStaysFlatWhenWarm(t *testing.T) {
	queries := func(people int) int {
		f := newFixture(t)
		srcB, artB := f.NewSource("Obituary")
		for i := 0; i < people; i++ {
			name := fmt.Sprintf("Person %c", 'A'+i)
			p := f.Person(f.source, f.artifact, name)
			b := f.Birth(f.source, f.artifact)
			f.Participation(f.source, f.artifact, p, b, "subject")
			f.promote(p)
			f.promote(b)
			pb := f.Person(srcB, artB, name)
			bb := f.Birth(srcB, artB)
			f.Participation(srcB, artB, pb, bb, "subject")
		}
		_, _, err := promotealign.Propose(f.c, f.Graph(srcB), nil, nil)
		must(t, err)
		n := 0
		f.c.Graph().SetQueryCounterForTest(func() { n++ })
		_, _, err = promotealign.Propose(f.c, f.Graph(srcB), nil, nil)
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
	ann := f.Person(f.source, f.artifact, "Ann Ames")
	annBirth := f.DatedBirth(f.source, f.artifact, 1850)
	f.Participation(f.source, f.artifact, ann, annBirth, "subject")
	annH, annBirthH := f.promote(ann), f.promote(annBirth)
	for i := 1; i <= 6; i++ {
		p := f.Person(f.source, f.artifact, fmt.Sprintf("Unrelated %d", i))
		b := f.DatedBirth(f.source, f.artifact, 1900+i)
		f.Participation(f.source, f.artifact, p, b, "subject")
		f.promote(p)
		f.promote(b)
	}
	srcB, artB := f.NewSource("Obituary")
	f.Person(srcB, artB, "Ann Ames")

	f.c.Graph().Drop()
	got, err := promotealign.ProfiledRefsForTest(f.c, f.Graph(srcB))
	must(t, err)
	want := []string{annH.Entity.Ref, annBirthH.Entity.Ref}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("profiled %v, want only %v", got, want)
	}
}

// A same-name, same-birth record whose sex conflicts still files strongly on
// the links.
func TestProposeFilesAConflictingRowOnTheLinks(t *testing.T) {
	f := newFixture(t)
	sexed := func(src sources.Source, art artifacts.Artifact, sex string) (subjects.Subject, subjects.Subject) {
		p := f.Person(src, art, "Ann Ames")
		f.Cite(art, p, observations.Input{
			PropertyID: f.Prop("sex_at_birth").ID, ValueTermID: f.Term("sex_at_birth", sex).ID,
		})
		b := f.DatedBirth(src, art, 1817)
		f.Participation(src, art, p, b, "subject")
		return p, b
	}
	pF, bF := sexed(f.source, f.artifact, "female")
	annH := f.promote(pF).Entity.ID
	f.promote(bF)

	src, art := f.NewSource("Register")
	pM, bM := sexed(src, art, "male")
	got := f.propose(f.Graph(src), nil)
	person, birth := rowFor(got, pM.ID), rowFor(got, bM.ID)
	if person.Target != graphalign.TargetHandle || !bytes.Equal(person.HandleID, annH) ||
		person.Assessment != graphalign.AssessStrong {
		t.Fatalf("person row %+v, want the female Ann Ames, strong", person)
	}
	if birth.Target != graphalign.TargetHandle || birth.Assessment != graphalign.AssessStrong {
		t.Fatalf("birth row %+v, want strong", birth)
	}
}

// A name one letter off is not shortlisted, so it is never scored against the
// filed person: the row is New.
func TestProposeDoesNotCompareAnUnshortlistedHandle(t *testing.T) {
	f := newFixture(t)
	f.promote(f.Person(f.source, f.artifact, "Robins"))

	src, art := f.NewSource("Register")
	robnis := f.Person(src, art, "Robnis")
	row := rowFor(f.propose(f.Graph(src), nil), robnis.ID)
	if row.Target != "" || row.Reason != graphalign.ReasonUnplaced || len(row.Alternatives) != 0 {
		t.Fatalf("Robnis row %+v, want unplaced with no alternatives", row)
	}
}

// A handle no shortlist key reaches is still loaded, scored and marked
// held when the researcher fixes the pair, and is not offered unfixed.
func TestProposeHoldsAnUnshortlistedHandle(t *testing.T) {
	tests := []struct {
		name        string
		filedSex    string // "" cites none
		incomingSex string
	}{
		{name: "no sex"},
		{name: "same sex", filedSex: "female", incomingSex: "female"},
		{name: "conflicting sex", filedSex: "female", incomingSex: "male"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			sexed := func(src sources.Source, art artifacts.Artifact, name, sex string) subjects.Subject {
				p := f.Person(src, art, name)
				if sex != "" {
					f.Cite(art, p, observations.Input{
						PropertyID: f.Prop("sex_at_birth").ID, ValueTermID: f.Term("sex_at_birth", sex).ID,
					})
				}
				return p
			}
			handle := f.promote(sexed(f.source, f.artifact, "Robins", tt.filedSex)).Entity.ID

			src, art := f.NewSource("Register")
			robnis := sexed(src, art, "Robnis", tt.incomingSex)
			graph := f.Graph(src)

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
		})
	}
}

// kinShape records one way of saying a kinship fact between a and b on src
// (m is the person between them, for a parent chain), and returns the Subjects
// it adds besides the people: events the tree promotes.
type kinShape func(f *fixture, src sources.Source, art artifacts.Artifact, a, m, b subjects.Subject) []subjects.Subject

// kinSide is one side of a proposal: the shape it records, whether it names the
// person between, and, for a Source, whether the target gives only a sex and
// whether Ann is fixed on her handle.
type kinSide struct {
	shape   kinShape
	mid     bool
	sexOnly bool
	fixAnn  bool
}

type kinRow struct {
	row         graphalign.Row
	ann, target subjects.Subject
}

func kinRelationship(term string) kinShape {
	return func(f *fixture, src sources.Source, art artifacts.Artifact, a, _, b subjects.Subject) []subjects.Subject {
		f.Relationship(src, art, a, b, term)
		return nil
	}
}

func kinMarriage(role string) kinShape {
	return func(f *fixture, src sources.Source, art artifacts.Artifact, a, _, b subjects.Subject) []subjects.Subject {
		return []subjects.Subject{f.Marriage(src, art, a, b, role)}
	}
}

func kinBirth(parentRole string) kinShape {
	return func(f *fixture, src sources.Source, art artifacts.Artifact, a, _, b subjects.Subject) []subjects.Subject {
		birth := f.EventOn(src, art, "birth", nil)
		f.Participation(src, art, a, birth, parentRole)
		f.Participation(src, art, b, birth, "subject")
		return []subjects.Subject{birth}
	}
}

func kinChain(f *fixture, src sources.Source, art artifacts.Artifact, a, m, b subjects.Subject) []subjects.Subject {
	f.Relationship(src, art, a, m, "parent")
	f.Relationship(src, art, m, b, "parent")
	return nil
}

func kinBoth(shapes ...kinShape) kinShape {
	return func(f *fixture, src sources.Source, art artifacts.Artifact, a, m, b subjects.Subject) []subjects.Subject {
		var extra []subjects.Subject
		for _, s := range shapes {
			extra = append(extra, s(f, src, art, a, m, b)...)
		}
		return extra
	}
}

// kinTree files a tree on f.source that records tree between Ann Ames and the
// target, promotes everything on it, then proposes each Source side as its own
// Source and returns the target's row of each.
func kinTree(t *testing.T, target string, tree kinSide, sides ...kinSide) (targetHandle []byte, out []kinRow) {
	t.Helper()
	f := newFixture(t)
	people := func(src sources.Source, art artifacts.Artifact, s kinSide) (a, m, b subjects.Subject) {
		a = f.Person(src, art, "Ann Ames")
		if s.mid {
			m = f.Person(src, art, "Bea Ames")
		}
		if s.sexOnly {
			b = f.SexOnly(src, art, "male")
		} else {
			b = f.Person(src, art, target)
		}
		return a, m, b
	}
	a, m, b := people(f.source, f.artifact, tree)
	extra := tree.shape(f, f.source, f.artifact, a, m, b)
	annHandle := f.promote(a).Entity.ID
	if tree.mid {
		f.promote(m)
	}
	targetHandle = f.promote(b).Entity.ID
	for _, s := range extra {
		f.promote(s)
	}
	for i, side := range sides {
		src, art := f.NewSource(fmt.Sprintf("Source %d", i))
		sa, sm, sb := people(src, art, side)
		side.shape(f, src, art, sa, sm, sb)
		var fixed []graphalign.Fixed
		if side.fixAnn {
			fixed = []graphalign.Fixed{{SubjectID: sa.ID, HandleID: annHandle}}
		}
		out = append(out, kinRow{row: rowFor(f.propose(f.Graph(src), fixed), sb.ID), ann: sa, target: sb})
	}
	return targetHandle, out
}

// onHandle is whether the row is proposed on the handle, at the band.
func onHandle(row graphalign.Row, handle []byte, band graphalign.Assessment) bool {
	return row.Target == graphalign.TargetHandle && bytes.Equal(row.HandleID, handle) && row.Assessment == band
}

// firstAlternative is the row's first alternative, and whether it is the handle.
func firstAlternative(row graphalign.Row, handle []byte) (graphalign.Alternative, bool) {
	if len(row.Alternatives) == 0 {
		return graphalign.Alternative{}, false
	}
	alt := row.Alternatives[0]
	return alt, bytes.Equal(alt.HandleID, handle)
}

func sameScore(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// A kinship fact recorded in another shape than the tree's helps the match as
// much as when both record it the same way, and counts once when both record
// it through the same paired shape.
func TestProposeKinshipShapes(t *testing.T) {
	spouse, parent, grandparent := kinRelationship("spouse"), kinRelationship("parent"), kinRelationship("grandparent")
	marriage, witnessed := kinMarriage("subject"), kinMarriage("witness")
	atBirth := kinBirth("father")
	// direct is the row of a Source that records the fact the tree does, in the
	// shape the tree does.
	direct := func(t *testing.T, target string, shape kinShape, mid, sexOnly, fixAnn bool) kinRow {
		_, rows := kinTree(t, target, kinSide{shape: shape, mid: mid}, kinSide{shape: shape, mid: mid, sexOnly: sexOnly, fixAnn: fixAnn})
		return rows[0]
	}

	t.Run("AC-1.1", func(t *testing.T) {
		bob, rows := kinTree(t, "Bob Ames", kinSide{shape: spouse}, kinSide{shape: spouse}, kinSide{shape: marriage}, kinSide{shape: witnessed})
		rel, married, witness := rows[0].row, rows[1].row, rows[2].row
		if !onHandle(married, bob, graphalign.AssessMedium) || !sameScore(married.Score, rel.Score) {
			t.Fatalf("marriage row %+v, want Bob's handle at medium, score %.4f", married, rel.Score)
		}
		if married.Score <= witness.Score || witness.Assessment != graphalign.AssessWeak {
			t.Fatalf("marriage %.4f, witnessed %+v: want the marriage above a weak witnessed one", married.Score, witness)
		}
	})
	t.Run("AC-1.2", func(t *testing.T) {
		want := direct(t, "Bob Ames", spouse, false, false, false)
		bob, rows := kinTree(t, "Bob Ames", kinSide{shape: marriage}, kinSide{shape: spouse})
		if !onHandle(rows[0].row, bob, graphalign.AssessMedium) || !sameScore(rows[0].row.Score, want.row.Score) {
			t.Fatalf("row %+v, want Bob's handle at medium, score %.4f", rows[0].row, want.row.Score)
		}
	})
	t.Run("AC-1.3", func(t *testing.T) {
		bob, rows := kinTree(t, "Bob Ames", kinSide{shape: parent},
			kinSide{shape: parent, sexOnly: true, fixAnn: true}, kinSide{shape: atBirth, sexOnly: true, fixAnn: true})
		want, got := rows[0].row, rows[1].row
		alt, ok := firstAlternative(got, bob)
		if !ok || len(want.Alternatives) == 0 || !sameScore(alt.Score, want.Alternatives[0].Score) {
			t.Fatalf("row %+v, want Bob's handle first, via, score %+v", got, want.Alternatives)
		}
	})
	t.Run("AC-1.4", func(t *testing.T) {
		want := direct(t, "Bob Ames", parent, false, true, true)
		bob, rows := kinTree(t, "Bob Ames", kinSide{shape: atBirth}, kinSide{shape: parent, sexOnly: true, fixAnn: true})
		alt, ok := firstAlternative(rows[0].row, bob)
		if !ok || len(want.row.Alternatives) == 0 || !sameScore(alt.Score, want.row.Alternatives[0].Score) {
			t.Fatalf("row %+v, want Bob's handle first, via, score %+v", rows[0].row, want.row.Alternatives)
		}
	})
	t.Run("AC-1.5", func(t *testing.T) {
		cal, rows := kinTree(t, "Cal Ames", kinSide{shape: grandparent},
			kinSide{shape: grandparent, fixAnn: true}, kinSide{shape: kinChain, mid: true, fixAnn: true})
		want, got := rows[0].row, rows[1].row
		if !onHandle(got, cal, graphalign.AssessMedium) || got.Reason != graphalign.ReasonVia || !sameScore(got.Score, want.Score) {
			t.Fatalf("row %+v, want Cal's handle at medium via, score %.4f", got, want.Score)
		}
		wantKey := vocabkey.BridgeKey{Type: vocabkey.BridgeRelationship, Term: "grandchild", End: vocabkey.PropertyRelatedTo}
		if got.Via == nil || !bytes.Equal(got.Via.NeighborSubjectID, rows[1].ann.ID) || got.Via.Key != wantKey {
			t.Fatalf("via %+v, want Ann's Source Subject through %s", got.Via, wantKey)
		}
	})
	t.Run("AC-1.6", func(t *testing.T) {
		want := direct(t, "Cal Ames", grandparent, false, false, true)
		cal, rows := kinTree(t, "Cal Ames", kinSide{shape: kinChain, mid: true}, kinSide{shape: grandparent, fixAnn: true})
		got := rows[0].row
		if !onHandle(got, cal, graphalign.AssessMedium) || got.Reason != graphalign.ReasonVia || !sameScore(got.Score, want.row.Score) {
			t.Fatalf("row %+v, want Cal's handle at medium via, score %.4f", got, want.row.Score)
		}
	})
	t.Run("AC-2.1", func(t *testing.T) {
		bob, rows := kinTree(t, "Bob Ames", kinSide{shape: spouse}, kinSide{shape: spouse}, kinSide{shape: kinBoth(spouse, marriage)})
		if !onHandle(rows[1].row, bob, graphalign.AssessMedium) || !sameScore(rows[1].row.Score, rows[0].row.Score) {
			t.Fatalf("row %+v, want Bob's handle at medium, score %.4f", rows[1].row, rows[0].row.Score)
		}
	})
	t.Run("AC-2.2", func(t *testing.T) {
		_, witness := kinTree(t, "Bob Ames", kinSide{shape: spouse}, kinSide{shape: witnessed})
		want := direct(t, "Bob Ames", spouse, false, false, false)
		bob, rows := kinTree(t, "Bob Ames", kinSide{shape: marriage}, kinSide{shape: marriage})
		got := rows[0].row
		if !onHandle(got, bob, graphalign.AssessMedium) || got.Score >= want.row.Score || got.Score <= witness[0].row.Score {
			t.Fatalf("row %+v, want Bob's handle at medium, between %.4f and %.4f", got, witness[0].row.Score, want.row.Score)
		}
	})
	t.Run("AC-2.3", func(t *testing.T) {
		_, wantRows := kinTree(t, "Bob Ames", kinSide{shape: marriage}, kinSide{shape: spouse, sexOnly: true, fixAnn: true})
		bob, rows := kinTree(t, "Bob Ames", kinSide{shape: marriage}, kinSide{shape: marriage, sexOnly: true, fixAnn: true})
		alt, ok := firstAlternative(rows[0].row, bob)
		if !ok || len(wantRows[0].row.Alternatives) == 0 || !sameScore(alt.Score, wantRows[0].row.Alternatives[0].Score) {
			t.Fatalf("row %+v, want Bob's handle first, via, score %+v", rows[0].row, wantRows[0].row.Alternatives)
		}
	})
	t.Run("AC-2.4", func(t *testing.T) {
		_, wantRows := kinTree(t, "Cal Ames", kinSide{shape: grandparent}, kinSide{shape: grandparent, fixAnn: true})
		cal, rows := kinTree(t, "Cal Ames", kinSide{shape: kinChain, mid: true}, kinSide{shape: kinChain, mid: true, fixAnn: true})
		got := rows[0].row
		if !onHandle(got, cal, graphalign.AssessMedium) || !sameScore(got.Score, wantRows[0].row.Score) {
			t.Fatalf("row %+v, want Cal's handle at medium, score %.4f", got, wantRows[0].row.Score)
		}
	})
	t.Run("AC-2.5", func(t *testing.T) {
		_, rows := kinTree(t, "Bob Ames", kinSide{shape: marriage}, kinSide{shape: marriage}, kinSide{shape: kinBoth(spouse, marriage)})
		if !sameScore(rows[1].row.Score, rows[0].row.Score) {
			t.Fatalf("both shapes %.4f, marriage alone %.4f, want equal", rows[1].row.Score, rows[0].row.Score)
		}
	})
}

// filedBirth files a person and his dated birth on src, joined by a subject
// participation (year 0: no date), joins them to the given entities or
// promotes them when an entity is nil, and returns them with their entities.
func (f *fixture) filedBirth(src sources.Source, art artifacts.Artifact, name string, year int, personHandle, birthHandle []byte) (p, b subjects.Subject, ph, bh []byte) {
	f.t.Helper()
	p = f.Person(src, art, name)
	if year == 0 {
		b = f.Birth(src, art)
	} else {
		b = f.DatedBirth(src, art, year)
	}
	f.Participation(src, art, p, b, "subject")
	ph, bh = personHandle, birthHandle
	if ph == nil {
		ph = f.promote(p).Entity.ID
	} else {
		f.join(p, ph)
	}
	if bh == nil {
		bh = f.promote(b).Entity.ID
	} else {
		f.join(b, bh)
	}
	return p, b, ph, bh
}

// A shared parent is not sibling evidence for Promote (R-18a): the siblings the
// page reads leave the proposal as it was. Siblings are no longer stored, so
// the read composes them and the proposal still cannot see them.
func TestProposeIgnoresPageSiblings(t *testing.T) {
	f := newFixture(t)
	ann := f.Person(f.source, f.artifact, "Ann Ames")
	bea := f.Person(f.source, f.artifact, "Bea Ames")
	cal := f.Person(f.source, f.artifact, "Cal Ames")
	f.Relationship(f.source, f.artifact, ann, bea, "parent")
	f.Relationship(f.source, f.artifact, ann, cal, "parent")
	f.promote(ann)
	beaID := f.promote(bea).Entity.ID
	calID := f.promote(cal).Entity.ID

	src, art := f.NewSource("Source 1")
	sbea := f.Person(src, art, "Bea Ames")
	scal := f.Person(src, art, "Cal Ames")
	f.Relationship(src, art, sbea, scal, "sibling")
	before := f.propose(f.Graph(src), nil)

	reader, err := canonlinks.NewReader(f.c.Graph())
	must(t, err)
	all, err := reader.Relationships([][]byte{beaID, calID}, canonlinks.Options{Siblings: true})
	must(t, err)
	if !slices.ContainsFunc(all[string(beaID)], func(l kinship.Link) bool {
		return slices.ContainsFunc(l.Shapes, func(sh kinship.Shape) bool { return sh.Pattern == kinship.PatternSharedParent })
	}) {
		t.Fatal("setup: Bea has no sibling through Ann")
	}
	after := f.propose(f.Graph(src), nil)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("proposal changed after the page read siblings:\n%+v\n%+v", before, after)
	}
}

// placeRelationships is how many association handles of the place relationship
// type the catalog holds: what Done has minted.
func (f *fixture) placeRelationships() int {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	var n int
	must(f.t, db.QueryRow(`SELECT COUNT(*) FROM canonical_entities ce
		JOIN subject_types st ON st.id = ce.subject_type_id
			AND st.key = 'place_relationship' AND st.origin = 'provenencia'`).Scan(&n))
	return n
}

// done runs Done on the graph with every given Subject on its handle, at the
// revision the proposal was read at.
func (f *fixture) done(graphID []byte, rev int64, onto map[*subjects.Subject][]byte) {
	f.t.Helper()
	var rows []promote.BatchRow
	for s, h := range onto {
		rows = append(rows, promote.BatchRow{SubjectID: s.ID, Target: promote.TargetHandle, EntityID: h})
	}
	_, err := writes.Call(f.c, writes.Op{Action: "promote_batch", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		return promote.SaveBatch(tx, userID, promote.Batch{GraphID: graphID, SeenRevision: rev, Rows: rows})
	})
	must(f.t, err)
}

// proposeConnections is the proposal's connections, with the revision.
func (f *fixture) proposeConnections(graphID []byte, fixed []graphalign.Fixed, only []byte) ([]graphalign.Connection, int64) {
	f.t.Helper()
	p, rev, err := promotealign.Propose(f.c, graphID, fixed, only)
	must(f.t, err)
	return p.Connections, rev
}

// The proposal says what Done files: a connection reads as filed exactly when
// Done mints its association, on a loop, a different term, a scoped page, a
// bridge already filed, a type changed after filing, and two type terms.
func TestProposeReportsEachConnection(t *testing.T) {
	// tree: York part_of Ontario, promoted on Source A. Source B then records
	// Ontario `kind` York and puts both ends on the tree's records.
	type tree struct {
		f           *fixture
		yorkH, ontH []byte
		srcB        sources.Source
		artB        artifacts.Artifact
		york2, ont2 subjects.Subject
		bridge      subjects.Subject
		treeBridge  subjects.Subject
		fixed       []graphalign.Fixed
		bridgeGraph []byte
	}
	build := func(t *testing.T, kind string) tree {
		f := newFixture(t)
		york := f.Place(f.source, f.artifact, "York")
		ontario := f.Place(f.source, f.artifact, "Ontario")
		treeBridge := f.PlaceRel(f.source, f.artifact, york, ontario, "part_of")
		yorkH := f.promote(york).Entity.ID
		ontH := f.promote(ontario).Entity.ID
		if n := f.placeRelationships(); n != 1 {
			t.Fatalf("tree filed %d place relationships, want 1", n)
		}
		srcB, artB := f.NewSource("Gazetteer")
		york2 := f.Place(srcB, artB, "York")
		ont2 := f.Place(srcB, artB, "Ontario")
		bridge := f.PlaceRel(srcB, artB, ont2, york2, kind)
		return tree{
			f: f, yorkH: yorkH, ontH: ontH, srcB: srcB, artB: artB, york2: york2, ont2: ont2, bridge: bridge, treeBridge: treeBridge,
			fixed: []graphalign.Fixed{
				{SubjectID: ont2.ID, HandleID: ontH},
				{SubjectID: york2.ID, HandleID: yorkH},
			},
			bridgeGraph: f.Graph(srcB),
		}
	}
	// agree fails unless the one connection reads as Done files it.
	agree := func(t *testing.T, tr tree, want graphalign.ConnectionOutcome) {
		t.Helper()
		conns, rev := tr.f.proposeConnections(tr.bridgeGraph, tr.fixed, nil)
		if len(conns) != 1 || conns[0].Outcome != want || !bytes.Equal(conns[0].BridgeID, tr.bridge.ID) {
			t.Fatalf("connections %+v, want one %q for the bridge", conns, want)
		}
		before := tr.f.placeRelationships()
		tr.f.done(tr.bridgeGraph, rev, map[*subjects.Subject][]byte{&tr.ont2: tr.ontH, &tr.york2: tr.yorkH})
		minted := tr.f.placeRelationships() - before
		files := 0
		if want == graphalign.ConnectionFiles {
			files = 1
		}
		if minted != files {
			t.Fatalf("Done minted %d place relationships, the proposal counted %d", minted, files)
		}
	}
	typeObs := func(t *testing.T, f *fixture, bridge subjects.Subject) observations.Listed {
		t.Helper()
		list, err := observations.ListBySubject(f.c, bridge.ID)
		must(t, err)
		for _, o := range list {
			if o.PropertyKey == "place_relationship_type" {
				return o
			}
		}
		t.Fatal("no type observation")
		return observations.Listed{}
	}

	t.Run("a loop with a filed link is refused", func(t *testing.T) {
		agree(t, build(t, "part_of"), graphalign.ConnectionOutcome("cycle"))
	})
	t.Run("another term over the same places files", func(t *testing.T) {
		agree(t, build(t, "succeeded_by"), graphalign.ConnectionFiles)
	})
	t.Run("the scope puts the other end off the page", func(t *testing.T) {
		tr := build(t, "part_of")
		conns, _ := tr.f.proposeConnections(tr.bridgeGraph, []graphalign.Fixed{{SubjectID: tr.york2.ID, HandleID: tr.yorkH}}, tr.york2.ID)
		if len(conns) != 1 || conns[0].Outcome != graphalign.ConnectionNotOnPage || !bytes.Equal(conns[0].Waiting, tr.ont2.ID) {
			t.Fatalf("connections %+v, want not_on_page waiting for Ontario", conns)
		}
	})
	t.Run("a filed bridge is not reported", func(t *testing.T) {
		f := newFixture(t)
		york := f.Place(f.source, f.artifact, "York")
		ontario := f.Place(f.source, f.artifact, "Ontario")
		f.PlaceRel(f.source, f.artifact, york, ontario, "part_of")
		f.promote(york)
		f.promote(ontario)
		if conns, _ := f.proposeConnections(f.Graph(f.source), nil, nil); len(conns) != 0 {
			t.Fatalf("connections %+v, want none", conns)
		}
	})
	t.Run("a type changed after filing", func(t *testing.T) {
		tr := build(t, "part_of")
		// The tree's own bridge now says succeeded_by, though it was filed as part_of.
		old := typeObs(t, tr.f, tr.treeBridge)
		must(t, func() error {
			_, err := observationstest.UpdateObservation(tr.f.c, userID, observations.Input{
				ID: old.ID, SubjectID: tr.treeBridge.ID, PropertyID: old.PropertyID,
				ValueTermID: tr.f.Term("place_relationship_type", "succeeded_by").ID,
			})
			return err
		}())
		conns, rev := tr.f.proposeConnections(tr.bridgeGraph, tr.fixed, nil)
		if len(conns) != 1 {
			t.Fatalf("connections %+v", conns)
		}
		before := tr.f.placeRelationships()
		tr.f.done(tr.bridgeGraph, rev, map[*subjects.Subject][]byte{&tr.ont2: tr.ontH, &tr.york2: tr.yorkH})
		minted := tr.f.placeRelationships() - before
		if (conns[0].Outcome == graphalign.ConnectionFiles) != (minted == 1) {
			t.Fatalf("proposal %q, Done minted %d", conns[0].Outcome, minted)
		}
	})
	t.Run("two type terms", func(t *testing.T) {
		tr := build(t, "succeeded_by")
		tr.f.Cite(tr.artB, tr.bridge, observations.Input{
			PropertyID:  tr.f.Prop("place_relationship_type").ID,
			ValueTermID: tr.f.Term("place_relationship_type", "part_of").ID,
		})
		conns, rev := tr.f.proposeConnections(tr.bridgeGraph, tr.fixed, nil)
		if len(conns) != 1 {
			t.Fatalf("connections %+v", conns)
		}
		before := tr.f.placeRelationships()
		tr.f.done(tr.bridgeGraph, rev, map[*subjects.Subject][]byte{&tr.ont2: tr.ontH, &tr.york2: tr.yorkH})
		minted := tr.f.placeRelationships() - before
		if (conns[0].Outcome == graphalign.ConnectionFiles) != (minted == 1) {
			t.Fatalf("proposal %q, Done minted %d", conns[0].Outcome, minted)
		}
	})
}
