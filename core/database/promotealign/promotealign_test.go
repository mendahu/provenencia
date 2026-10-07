package promotealign_test

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
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
	p, err := promotealign.Propose(db, sourceID, fixed)
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

func TestProposeInvalidSource(t *testing.T) {
	f := newFixture(t)
	db, err := f.c.DB()
	must(t, err)
	_, err = promotealign.Propose(db, make([]byte, 16), nil)
	if err != promote.ErrInvalid {
		t.Fatalf("err %v, want promote.ErrInvalid", err)
	}
}

func TestStatsCacheInvalidatesOnWrite(t *testing.T) {
	f := newFixture(t)
	db, err := f.c.DB()
	must(t, err)

	// Prime cache.
	_, err = promotealign.Propose(db, f.source.ID, nil)
	must(t, err)

	// A write bumps audit revision; next Propose must recompute (no panic / stale).
	_ = f.person(f.source, f.artifact, "Ada")
	got := f.propose(f.source.ID, nil)
	if len(got.Rows) != 1 {
		t.Fatalf("rows %d", len(got.Rows))
	}
}
