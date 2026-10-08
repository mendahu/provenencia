package promotealign_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/ref"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

const locator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

type fixture struct {
	t        *testing.T
	c        *database.Catalog
	source   sources.Source
	artifact artifacts.Artifact
	y        int64
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	promotealign.ResetStatsCacheForTest()
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
	return &fixture{t: t, c: c, source: src, artifact: art}
}

func (f *fixture) newSource(title string) (sources.Source, artifacts.Artifact) {
	f.t.Helper()
	typeID, err := sourcetypes.Upsert(f.c, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	must(f.t, err)
	src, err := sources.Create(f.c, userID, sources.CreateInput{SourceTypeID: typeID, Title: title})
	must(f.t, err)
	art, err := artifacts.Create(f.c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	must(f.t, err)
	return src, art
}

func (f *fixture) prop(key string) properties.Property {
	f.t.Helper()
	p, err := properties.Lookup(f.c, key, properties.OriginProvenencia)
	must(f.t, err)
	return p
}

func (f *fixture) term(propKey, termKey string) propertyterms.Term {
	f.t.Helper()
	t, err := propertyterms.Lookup(f.c, f.prop(propKey).ID, termKey, propertyterms.OriginProvenencia)
	must(f.t, err)
	return t
}

func (f *fixture) bare(src sources.Source, kind string) subjects.Subject {
	f.t.Helper()
	st, err := subjecttypes.Lookup(f.c, kind, subjecttypes.OriginProvenencia)
	must(f.t, err)
	s, err := subjects.Create(f.c, userID, subjects.CreateInput{SourceID: src.ID, SubjectTypeID: st.ID}, nil)
	must(f.t, err)
	_, err = subjectpositions.Set(f.c, s.ID, 0, f.y)
	must(f.t, err)
	f.y += 2
	return s
}

func (f *fixture) cite(art artifacts.Artifact, s subjects.Subject, in ...observations.Input) {
	f.t.Helper()
	for i := range in {
		in[i].SubjectID = s.ID
	}
	_, err := citations.CreateWithObservations(f.c, userID, citations.CreateInput{
		ArtifactID: art.ID, LocatorJSON: locator,
	}, in)
	must(f.t, err)
}

func (f *fixture) person(src sources.Source, art artifacts.Artifact, name string) subjects.Subject {
	f.t.Helper()
	s := f.bare(src, "person")
	f.cite(art, s, observations.Input{
		PropertyID: f.prop("name").ID, Name: namevaluestest.Western(name),
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
	_, err := connect.CreateCitedBridge(f.c, userID, connect.CreateInput{
		SourceID: src.ID, FromSubjectID: person.ID, ToSubjectID: event.ID, BridgeTypeKey: "participation",
		Citation: citations.CreateInput{ArtifactID: art.ID, LocatorJSON: locator},
		Observations: []observations.Input{
			{PropertyID: f.prop("person").ID, ValueSubjectID: person.ID},
			{PropertyID: f.prop("event").ID, ValueSubjectID: event.ID},
			{PropertyID: f.prop("role").ID, ValueTermID: f.term("role", role).ID},
		},
	})
	must(f.t, err)
}

func (f *fixture) placeRel(src sources.Source, art artifacts.Artifact, from, to subjects.Subject, kind string) {
	f.t.Helper()
	_, err := connect.CreateCitedBridge(f.c, userID, connect.CreateInput{
		SourceID: src.ID, FromSubjectID: from.ID, ToSubjectID: to.ID, BridgeTypeKey: "place_relationship",
		Citation: citations.CreateInput{ArtifactID: art.ID, LocatorJSON: locator},
		Observations: []observations.Input{
			{PropertyID: f.prop("from").ID, ValueSubjectID: from.ID},
			{PropertyID: f.prop("to").ID, ValueSubjectID: to.ID},
			{PropertyID: f.prop("place_relationship_type").ID, ValueTermID: f.term("place_relationship_type", kind).ID},
		},
	})
	must(f.t, err)
}

func (f *fixture) promote(s subjects.Subject) promote.Result {
	f.t.Helper()
	res, err := promote.Save(f.c, userID, promote.Input{SubjectID: s.ID})
	must(f.t, err)
	return res
}

func (f *fixture) propose(sourceID []byte, fixed []graphalign.Fixed) graphalign.Proposal {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	p, _, err := promotealign.Propose(db, sourceID, fixed)
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

	got := f.propose(srcB.ID, []graphalign.Fixed{{
		SubjectID: bBirth.ID, HandleID: evt.Entity.ID,
	}})
	row := rowFor(got, bPerson.ID)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, per.Entity.ID) {
		t.Fatalf("person row %+v, want handle %x", row, per.Entity.ID)
	}
	if row.Via == nil || !bytes.Equal(row.Via.NeighborSubjectID, bBirth.ID) || row.Via.Signature.BridgeType != "participation" {
		t.Fatalf("via %+v, want birth participation", row.Via)
	}
	var namePinned bool
	for _, ex := range row.Exhibits {
		if ex.Property.Key == "name" && ex.Outcome == "agree" && ex.Pinned &&
			len(ex.IncomingObservationID) == 16 && len(ex.MemberObservationID) == 16 &&
			ex.IncomingDisplay == "Gracie" && ex.IncomingSource == "Obituary" {
			namePinned = true
		}
	}
	if !namePinned {
		t.Fatalf("name exhibit missing in %+v", row.Exhibits)
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

	got := f.propose(srcB.ID, []graphalign.Fixed{{
		SubjectID: ontario2.ID, HandleID: ontarioH.Entity.ID,
	}})
	row := rowFor(got, york2.ID)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, yorkH.Entity.ID) {
		t.Fatalf("york row %+v, want handle %x", row, yorkH.Entity.ID)
	}
}

func TestProposeOneHopDateExhibit(t *testing.T) {
	f := newFixture(t)
	year := 1901
	date := &datevalues.Value{Kind: datevalues.KindPoint, Calendar: "gregorian", StartYear: &year}

	aPerson := f.person(f.source, f.artifact, "Gracie")
	aBirth := f.bare(f.source, "event")
	f.cite(f.artifact, aBirth,
		observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID},
		observations.Input{PropertyID: f.prop("date").ID, Date: date},
	)
	f.participation(f.source, f.artifact, aPerson, aBirth, "subject")
	per := f.promote(aPerson)
	evt := f.promote(aBirth)

	srcB, artB := f.newSource("Obituary")
	bPerson := f.person(srcB, artB, "Gracie")
	bBirth := f.bare(srcB, "event")
	f.cite(artB, bBirth,
		observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID},
		observations.Input{PropertyID: f.prop("date").ID, Date: date},
	)
	f.participation(srcB, artB, bPerson, bBirth, "subject")

	got := f.propose(srcB.ID, []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: evt.Entity.ID}})
	row := rowFor(got, bPerson.ID)
	if !bytes.Equal(row.HandleID, per.Entity.ID) {
		t.Fatalf("person %+v", row.Target)
	}
	var hop bool
	for _, ex := range row.Exhibits {
		if ex.Property.Key == "date" && ex.Outcome == "agree" && ex.Pinned &&
			bytes.Equal(ex.GroupSubjectID, bBirth.ID) && ex.IncomingDisplay == "1901" && ex.MemberDisplay == "1901" {
			hop = true
		}
	}
	if !hop {
		t.Fatalf("one-hop date exhibit missing in %+v", row.Exhibits)
	}
	for _, ex := range row.Exhibits {
		if ex.Property.Key == "event_type" && (ex.IncomingDisplay != "Birth" || ex.MemberDisplay != "Birth") {
			t.Fatalf("a term shows its label, not its key: %+v", ex)
		}
	}
}

func TestProposeInvalidSource(t *testing.T) {
	f := newFixture(t)
	db, err := f.c.DB()
	must(t, err)
	_, _, err = promotealign.Propose(db, make([]byte, 16), nil)
	if err != promote.ErrInvalid {
		t.Fatalf("err %v, want promote.ErrInvalid", err)
	}
}

func TestStatsCacheInvalidatesOnWrite(t *testing.T) {
	f := newFixture(t)
	db, err := f.c.DB()
	must(t, err)

	// Prime cache.
	_, _, err = promotealign.Propose(db, f.source.ID, nil)
	must(t, err)

	// A write bumps audit revision; next Propose must recompute (no panic / stale).
	_ = f.person(f.source, f.artifact, "Ada")
	got := f.propose(f.source.ID, nil)
	if len(got.Rows) != 1 {
		t.Fatalf("rows %d", len(got.Rows))
	}
}

func TestStatsFanOut(t *testing.T) {
	subjectAtBirth := graphalign.EdgeSignature{
		BridgeType: "participation", RoleOrType: "subject", NeighborKind: "event", NeighborTypeTerm: "birth",
	}
	revision := func(f *fixture) int64 {
		db, err := f.c.DB()
		must(f.t, err)
		var rev int64
		must(f.t, db.QueryRow(`SELECT MAX(revision) FROM audit_transactions`).Scan(&rev))
		return rev
	}
	stats := func(f *fixture) graphalign.Stats {
		db, err := f.c.DB()
		must(f.t, err)
		st, err := promotealign.LoadStatsForTest(db)
		must(f.t, err)
		return st
	}
	// Both catalogs hold a person, their birth, and a lone place. Filing
	// person and birth files the participation; filing person and place doesn't.
	build := func(t *testing.T, fileBirth bool) *fixture {
		f := &fixture{t: t}
		c, err := database.Create(t.TempDir(), "t.provenencia")
		must(t, err)
		t.Cleanup(func() { _ = c.Close() })
		r, err := ref.Mint(ref.PrefixUser)
		must(t, err)
		must(t, users.Upsert(c, userID, "Tester", r))
		must(t, subjectvocab.Install(c))
		f.c = c
		f.source, f.artifact = f.newSource("Register")
		person := f.person(f.source, f.artifact, "Gracie")
		birth := f.birth(f.source, f.artifact)
		place := f.place(f.source, f.artifact, "York")
		f.participation(f.source, f.artifact, person, birth, "subject")
		f.promote(person)
		if fileBirth {
			f.promote(birth)
		} else {
			f.promote(place)
		}
		return f
	}

	t.Run("keys match the walk's edge signature", func(t *testing.T) {
		promotealign.ResetStatsCacheForTest()
		f := build(t, true)
		if got, ok := stats(f).FanOut[subjectAtBirth.Key()]; !ok || got != 1 {
			t.Fatalf("fan-out %v (present %v), want 1 under %q", got, ok, subjectAtBirth.Key())
		}
	})
	t.Run("another project at the same revision recomputes", func(t *testing.T) {
		promotealign.ResetStatsCacheForTest()
		filed, other := build(t, true), build(t, false)
		if revision(filed) != revision(other) {
			t.Fatalf("revisions %d and %d differ; the test needs them equal", revision(filed), revision(other))
		}
		_ = stats(filed)
		if _, leaked := stats(other).FanOut[subjectAtBirth.Key()]; leaked {
			t.Fatal("stats from the first catalog were served for the second")
		}
	})
}

// One-hop exhibits compare a neighbor only with the handle Align mapped it to.
// The handle's person also has a death in 1901; a layer event dated 1901 must
// not "agree" with it unless that event is mapped to the death.
func TestProposeOneHopExhibitsPairOnlyTheMappedNeighbor(t *testing.T) {
	yearOf := func(y int) *datevalues.Value {
		return &datevalues.Value{Kind: datevalues.KindPoint, Calendar: "gregorian", StartYear: &y}
	}
	tests := []struct {
		name        string
		layerType   string
		fixLayerEvt bool // fix the layer event to the canon birth
		want        string
	}{
		{name: "a birth mapped to the birth compares with the birth", layerType: "birth", fixLayerEvt: true, want: "conflict"},
		{name: "an event mapped to nothing adds no one-hop lines", layerType: "marriage", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			event := func(src sources.Source, art artifacts.Artifact, kind string, year int) subjects.Subject {
				s := f.bare(src, "event")
				f.cite(art, s,
					observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", kind).ID},
					observations.Input{PropertyID: f.prop("date").ID, Date: yearOf(year)},
				)
				return s
			}
			aPerson := f.person(f.source, f.artifact, "Gracie")
			aBirth := event(f.source, f.artifact, "birth", 1880)
			aDeath := event(f.source, f.artifact, "death", 1901)
			f.participation(f.source, f.artifact, aPerson, aBirth, "subject")
			f.participation(f.source, f.artifact, aPerson, aDeath, "subject")
			per := f.promote(aPerson)
			birthH := f.promote(aBirth)
			f.promote(aDeath)

			srcB, artB := f.newSource("Obituary")
			bPerson := f.person(srcB, artB, "Gracie")
			bEvent := event(srcB, artB, tt.layerType, 1901)
			f.participation(srcB, artB, bPerson, bEvent, "subject")

			fixed := []graphalign.Fixed{{SubjectID: bPerson.ID, HandleID: per.Entity.ID}}
			if tt.fixLayerEvt {
				fixed = append(fixed, graphalign.Fixed{SubjectID: bEvent.ID, HandleID: birthH.Entity.ID})
			}
			row := rowFor(f.propose(srcB.ID, fixed), bPerson.ID)
			got := ""
			for _, ex := range row.Exhibits {
				if ex.Property.Key == "date" {
					if got != "" {
						t.Fatalf("more than one date line: %+v", row.Exhibits)
					}
					got = string(ex.Outcome)
					if ex.Pinned != (ex.Outcome == "agree") {
						t.Fatalf("pinned %v on a %s line", ex.Pinned, ex.Outcome)
					}
				}
			}
			if got != tt.want {
				t.Fatalf("date line %q, want %q", got, tt.want)
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
			db, err := f.c.DB()
			must(t, err)
			p, _, err := promotealign.Propose(db, srcB.ID, []graphalign.Fixed{fixed})
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

type countingQuerier struct {
	db *sql.DB
	n  int
}

func (c *countingQuerier) Query(query string, args ...any) (*sql.Rows, error) {
	c.n++
	return c.db.Query(query, args...)
}

func (c *countingQuerier) QueryRow(query string, args ...any) *sql.Row {
	c.n++
	return c.db.QueryRow(query, args...)
}

// Every read in Propose is batched, so a bigger layer and catalog cost no
// more queries than a small one of the same shape.
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
		db, err := f.c.DB()
		must(t, err)
		promotealign.ResetStatsCacheForTest()
		counter := &countingQuerier{db: db}
		_, _, err = promotealign.Propose(counter, srcB.ID, nil)
		must(t, err)
		return counter.n
	}
	small, large := queries(2), queries(6)
	t.Logf("%d queries per proposal", small)
	if small != large {
		t.Fatalf("2 people took %d queries, 6 took %d; a read grew with the layer", small, large)
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

	row := rowFor(f.propose(srcB.ID, []graphalign.Fixed{{SubjectID: ontario.ID, HandleID: middleH.Entity.ID}}), york.ID)
	if row.Reason != graphalign.ReasonVia || !bytes.Equal(row.HandleID, partH.Entity.ID) {
		t.Fatalf("york row %+v, want the part %s via Ontario", row, partH.Entity.Ref)
	}
}
