package promotealign_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	_ "github.com/mendahu/provenencia/core/database/sourcefill"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/valuecodec"
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
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	must(t, err)
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	must(t, err)
	art, err := runArtifactCreate(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	must(t, err)
	return &fixture{t: t, c: c, source: src, artifact: art}
}

func (f *fixture) newSource(title string) (sources.Source, artifacts.Artifact) {
	f.t.Helper()
	typeID, err := sourcetypes.Upsert(f.c, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	must(f.t, err)
	src, _, err := writes.Run(f.c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: title})
	})
	must(f.t, err)
	art, err := runArtifactCreate(f.c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
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
	s, err := writes.Call(f.c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{SourceID: src.ID, SubjectTypeID: st.ID}, nil)
	})
	must(f.t, err)
	_, err = writes.Call(f.c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
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
			SourceID: src.ID, FromSubjectID: person.ID, ToSubjectID: event.ID, BridgeTypeKey: "participation",
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
			SourceID: src.ID, FromSubjectID: event.ID, ToSubjectID: place.ID, BridgeTypeKey: "location",
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
			SourceID: src.ID, FromSubjectID: from.ID, ToSubjectID: to.ID, BridgeTypeKey: "place_relationship",
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
			SourceID: src.ID, FromSubjectID: person.ID, ToSubjectID: related.ID, BridgeTypeKey: "relationship",
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
	res, err := writes.Call(f.c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, promote.Input{SubjectID: s.ID})
	})
	must(f.t, err)
	return res
}

func (f *fixture) propose(sourceID []byte, fixed []graphalign.Fixed) graphalign.Proposal {
	f.t.Helper()
	p, _, err := promotealign.Propose(f.c, sourceID, fixed)
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

func TestProposeOverlappingBoundShowsTheDateAndResembles(t *testing.T) {
	f := newFixture(t)
	bef := func(year, month, day int) *datevalue.Value {
		y, m, d := year, month, day
		return &datevalue.Value{
			Kind: datevalue.KindPoint, Qualifier: datevalue.QualifierBEF,
			StartYear: &y, StartMonth: &m, StartDay: &d,
		}
	}
	earlier, later := bef(1985, 2, 1), bef(2001, 3, 31)

	aPerson := f.person(f.source, f.artifact, "Doug")
	aEvent := f.bare(f.source, "event")
	f.cite(f.artifact, aEvent,
		observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID},
		observations.Input{PropertyID: f.prop("start_date").ID, Date: earlier},
	)
	f.participation(f.source, f.artifact, aPerson, aEvent, "subject")
	per := f.promote(aPerson)
	evt := f.promote(aEvent)

	srcB, artB := f.newSource("Letter")
	bPerson := f.person(srcB, artB, "Doug")
	bEvent := f.bare(srcB, "event")
	f.cite(artB, bEvent,
		observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID},
		observations.Input{PropertyID: f.prop("start_date").ID, Date: later},
	)
	f.participation(srcB, artB, bPerson, bEvent, "subject")

	got := f.propose(srcB.ID, []graphalign.Fixed{{SubjectID: bPerson.ID, HandleID: per.Entity.ID}})
	row := rowFor(got, bEvent.ID)
	if !bytes.Equal(row.HandleID, evt.Entity.ID) {
		t.Fatalf("event %+v, want %s", row, evt.Entity.Ref)
	}
	var saw bool
	for _, ex := range row.Exhibits {
		if ex.Property.Key != "start_date" {
			continue
		}
		saw = true
		if ex.Outcome != "partial" || ex.Pinned || ex.Weight <= 0 ||
			ex.IncomingDisplay != "Before 2001-03-31" || ex.MemberDisplay != "Before 1985-02-01" {
			t.Fatalf("start date %+v", ex)
		}
	}
	if !saw {
		t.Fatalf("missing start date in %+v", row.Exhibits)
	}
}

func TestProposeOneHopDateExhibit(t *testing.T) {
	f := newFixture(t)
	year := 1901
	date := &datevalue.Value{Kind: datevalue.KindPoint, Calendar: "gregorian", StartYear: &year}

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
	_, _, err := promotealign.Propose(f.c, make([]byte, 16), nil)
	if err != promote.ErrInvalid {
		t.Fatalf("err %v, want promote.ErrInvalid", err)
	}
}

func TestStatsCacheInvalidatesOnWrite(t *testing.T) {
	f := newFixture(t)

	// Prime cache.
	_, _, err := promotealign.Propose(f.c, f.source.ID, nil)
	must(t, err)

	// A write bumps audit revision; next Propose must recompute (no panic / stale).
	_ = f.person(f.source, f.artifact, "Ada")
	got := f.propose(f.source.ID, nil)
	if len(got.Rows) != 1 {
		t.Fatalf("rows %d", len(got.Rows))
	}

	// Name counts are read once between handle commits, then recounted.
	f.promote(f.person(f.source, f.artifact, "Ada Smith"))
	read := func() graphalign.Stats {
		st, err := promotealign.LoadStatsForTest(f.c)
		must(t, err)
		return st
	}
	first, again := read(), read()
	if reflect.ValueOf(first.Words.Words).Pointer() != reflect.ValueOf(again.Words.Words).Pointer() {
		t.Fatal("two reads with no commit between them counted twice")
	}
	f.promote(f.person(f.source, f.artifact, "Bob Smith"))
	if after := read(); after.Words.Words["smith"] != first.Words.Words["smith"]+1 {
		t.Fatalf("smith %d after a handle commit, was %d", after.Words.Words["smith"], first.Words.Words["smith"])
	}
}

// sixty files 60 people named John <Surname>: 30 Smith, 1 Quirk, 29 others.
func (f *fixture) sixty() {
	f.t.Helper()
	for i := 0; i < 30; i++ {
		f.promote(f.person(f.source, f.artifact, "John Smith"))
	}
	f.promote(f.person(f.source, f.artifact, "John Quirk"))
	for i := 1; i <= 29; i++ {
		f.promote(f.person(f.source, f.artifact, fmt.Sprintf("John Surname%02d", i)))
	}
}

func TestStatsCountsNameWords(t *testing.T) {
	f := newFixture(t)
	db, err := f.c.DB()
	must(t, err)
	both := f.promote(f.person(f.source, f.artifact, "John Smith")).Entity.ID
	f.promote(f.person(f.source, f.artifact, "José Smith"))
	merged := f.promote(f.person(f.source, f.artifact, "Jane Smith")).Entity.ID
	// A second reconciled value on the first handle that also spells Smith.
	blob, err := valuecodec.MarshalName(*namevaluetest.Western("J. Smith"))
	must(t, err)
	_, err = db.Exec(`INSERT INTO auto_reconciler_values (entity_id, property_id, rank, value_name, support, against, reason)
		VALUES (?, ?, 2, ?, 1, 0, 'kept')`, both, f.prop("name").ID, blob)
	must(t, err)
	_, err = db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, both, merged)
	must(t, err)
	f.c.Graph().Drop()

	st, err := promotealign.LoadStatsForTest(f.c)
	must(t, err)
	if st.Words.People != 2 || st.Words.Words["smith"] != 2 || st.Words.Words["john"] != 1 || st.Words.Words["jose"] != 1 {
		t.Fatalf("counts %+v", st.Words)
	}
	if _, ok := st.Words.Words["jane"]; ok {
		t.Fatalf("a merged handle was counted: %+v", st.Words)
	}
	if _, ok := st.Words.Words["j"]; ok {
		t.Fatalf("an initial was counted: %+v", st.Words)
	}
}

// Two names that agree and pin, one common in the project and one rare: the
// common one earns fewer points. Similarity and pinning do not move.
func TestProposeWeighsACommonNameLess(t *testing.T) {
	f := newFixture(t)
	f.sixty()
	src, art := f.newSource("Register")
	smith := f.person(src, art, "John Smith")
	quirk := f.person(src, art, "John Quirk")
	p := f.propose(src.ID, nil)
	nameWeight := func(s subjects.Subject) float64 {
		for _, c := range rowFor(p, s.ID).Comparisons {
			if c.Property.Key == "name" {
				if c.Outcome != match.OutcomeAgree {
					t.Fatalf("%+v, want agree", c)
				}
				return c.Weight
			}
		}
		t.Fatal("no name comparison")
		return 0
	}
	q, sm := nameWeight(quirk), nameWeight(smith)
	t.Logf("name weight: quirk %.3f, smith %.3f", q, sm)
	if q <= sm {
		t.Fatalf("quirk %.3f, smith %.3f", q, sm)
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
		st, err := promotealign.LoadStatsForTest(f.c)
		must(f.t, err)
		return st
	}
	// Both catalogs hold a person, their birth, and a lone place. Filing
	// person and birth files the participation; filing person and place doesn't.
	build := func(t testing.TB, fileBirth bool) *fixture {
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
		f := build(t, true)
		if got, ok := stats(f).FanOut[subjectAtBirth.Key()]; !ok || got != 1 {
			t.Fatalf("fan-out %v (present %v), want 1 under %q", got, ok, subjectAtBirth.Key())
		}
	})
	t.Run("another project at the same revision recomputes", func(t *testing.T) {
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
	yearOf := func(y int) *datevalue.Value {
		return &datevalue.Value{Kind: datevalue.KindPoint, Calendar: "gregorian", StartYear: &y}
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
			p, _, err := promotealign.Propose(f.c, srcB.ID, []graphalign.Fixed{fixed})
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
		_, _, err := promotealign.Propose(f.c, srcB.ID, nil)
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

	prop := f.propose(srcB.ID, nil)
	cfg := graphalign.DefaultConfig()
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
			row.Assessment != graphalign.AssessStrong || row.Score < cfg.StrongScore {
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
	row := rowFor(f.propose(srcB.ID, nil), again.ID)
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
			row := rowFor(f.propose(srcB.ID, nil), again.ID)
			cfg := graphalign.DefaultConfig()
			if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, known.Entity.ID) ||
				row.Assessment != graphalign.AssessWeak ||
				row.Score < cfg.WeakScore || row.Score >= cfg.MediumScore {
				t.Fatalf("%s row %+v, want %s as a weak match", tc.short, row, known.Entity.Ref)
			}
			var saw bool
			for _, ex := range row.Exhibits {
				if ex.Property.Key != "name" {
					continue
				}
				saw = true
				if ex.Outcome == "conflict" || ex.Outcome == "unknown" {
					t.Fatalf("name line %+v", ex)
				}
			}
			if !saw {
				t.Fatal("missing name exhibit")
			}
		})
	}
}

func TestSpellingVariantToponymIsAWeakMatch(t *testing.T) {
	f := newFixture(t)
	named := f.promote(f.place(f.source, f.artifact, "Provenanced"))
	f.promote(f.place(f.source, f.artifact, "Leeds"))

	srcB, artB := f.newSource("Gazetteer")
	again := f.place(srcB, artB, "Provenance")
	row := rowFor(f.propose(srcB.ID, nil), again.ID)
	cfg := graphalign.DefaultConfig()
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, named.Entity.ID) ||
		row.Assessment != graphalign.AssessWeak ||
		row.Score < cfg.WeakScore || row.Score >= cfg.MediumScore {
		t.Fatalf("provenance row %+v, want %s as a weak match", row, named.Entity.Ref)
	}
	var saw bool
	for _, ex := range row.Exhibits {
		if ex.Property.Key != "toponym" {
			continue
		}
		saw = true
		if ex.Outcome != "partial" || ex.Pinned {
			t.Fatalf("toponym exhibit %+v", ex)
		}
	}
	if !saw {
		t.Fatal("missing toponym exhibit")
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
	son := f.sexOnly(src2, art2, "male")
	mother := f.person(src2, art2, "Mary Smith")
	f.relationship(src2, art2, son, mother, "child")

	p := f.propose(src2.ID, []graphalign.Fixed{{SubjectID: mother.ID, HandleID: hMary}})
	r := rowFor(p, son.ID)
	if r.Target != graphalign.TargetHandle || !bytes.Equal(r.HandleID, hJohn) || r.Reason != graphalign.ReasonVia {
		t.Fatalf("son: target=%s handle-is-John=%v reason=%s, want John's handle via his mother",
			r.Target, bytes.Equal(r.HandleID, hJohn), r.Reason)
	}
}

// A relationship is walked from either end: anchoring the child reaches the
// parent as anchoring the parent reaches the child.
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
	}{
		{name: "child anchored reaches the parent"},
		{name: "parent anchored reaches the child", anchorIsParent: true},
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
				fixed = graphalign.Fixed{SubjectID: parent.ID, HandleID: hMary}
				want = hJohn
			} else {
				parent = f.sexOnly(src, art, "female")
				child = f.person(src, art, "John Smith")
				fixed = graphalign.Fixed{SubjectID: child.ID, HandleID: hJohn}
				want = hMary
			}
			f.relationship(src, art, parent, child, "parent")
			p := f.propose(src.ID, []graphalign.Fixed{fixed})
			other := child
			if !tt.anchorIsParent {
				other = parent
			}
			r := rowFor(p, other.ID)
			if r.Target != graphalign.TargetHandle || !bytes.Equal(r.HandleID, want) || r.Reason != graphalign.ReasonVia {
				t.Fatalf("target=%s right-handle=%v reason=%s, want the other end's handle via the anchor",
					r.Target, bytes.Equal(r.HandleID, want), r.Reason)
			}
		})
	}
}

// A place with several events is a hub. The proposal still has one row per
// layer subject, and the place lands on that hub.
func TestHubPlaceProposalStaysOnTheLayer(t *testing.T) {
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

	got := f.propose(srcB.ID, nil)
	if len(got.Rows) != 2 {
		t.Fatalf("rows %d, want the layer's place and birth", len(got.Rows))
	}
	row := rowFor(got, city2.ID)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, cityH.Entity.ID) {
		t.Fatalf("place row %+v, want the hub", row.Target)
	}
	if err := f.c.Graph().Verify(); err != nil {
		t.Fatal(err)
	}
}

func runArtifactCreate(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error) {
	a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, in)
		})
	return a, err
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
	r := rowFor(f.propose(src.ID, nil), unknown.ID)
	if r.Target != graphalign.TargetHandle || !bytes.Equal(r.HandleID, handle) {
		t.Fatalf("unknown sex: target %s score %.2f, want the male John Smith", r.Target, r.Score)
	}
	for _, c := range r.Comparisons {
		if c.Property.Key == "sex_at_birth" && c.Outcome != match.OutcomeUnknown {
			t.Fatalf("unknown sex compared as %s", c.Outcome)
		}
	}

	src2, art2 := f.newSource("Census")
	female := named(src2, art2, "female")
	if r := rowFor(f.propose(src2.ID, nil), female.ID); r.Target == graphalign.TargetHandle && bytes.Equal(r.HandleID, handle) {
		t.Fatalf("female John Smith landed on the male one (score %.2f)", r.Score)
	}
}

// Jake Robins agrees with James Kenneth Jake Robins: Jake and Robins match,
// and James and Kenneth exist only on the conclusion, so they are neutral.
// Gender agrees as well. Together that is a medium match.
func TestProposeExtraGivenNamesAreNeutral(t *testing.T) {
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
	row := rowFor(f.propose(noteSrc.ID, nil), probe.ID)
	cfg := graphalign.DefaultConfig()
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, handle) ||
		row.Assessment != graphalign.AssessMedium || row.Score < cfg.MediumScore || row.Score >= cfg.StrongScore {
		t.Fatalf("Jake Robins: target %s assessment %s score %.2f", row.Target, row.Assessment, row.Score)
	}
}

func (f *fixture) join(s subjects.Subject, entityID []byte) {
	f.t.Helper()
	_, err := writes.Call(f.c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, promote.Input{SubjectID: s.ID, EntityID: entityID})
	})
	must(f.t, err)
}

func comparisonOutcome(t *testing.T, row graphalign.Row, key string) match.Outcome {
	t.Helper()
	for _, c := range row.Comparisons {
		if c.Property.Key == key {
			return c.Outcome
		}
	}
	t.Fatalf("no %s comparison in %+v", key, row.Comparisons)
	return ""
}

// conflictExhibit is the row's own-group conflict line for a Property.
func conflictExhibit(t *testing.T, row graphalign.Row, key string) graphalign.Exhibit {
	t.Helper()
	for _, e := range row.Exhibits {
		if e.Property.Key == key && e.Outcome == match.OutcomeConflict && len(e.GroupSubjectID) == 0 {
			return e
		}
	}
	t.Fatalf("no %s conflict exhibit in %+v", key, row.Exhibits)
	return graphalign.Exhibit{}
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
			VALUES (?, ?, 2, ?, 1, 0, 'kept')`, handle, f.prop("name").ID, blob)
		must(t, err)
		f.c.Graph().Drop()

		src, art := f.newSource("Register")
		incoming := f.person(src, art, "James Robins")
		row := rowFor(f.propose(src.ID, []graphalign.Fixed{{SubjectID: incoming.ID, HandleID: handle}}), incoming.ID)
		if got := comparisonOutcome(t, row, "name"); got != match.OutcomeConflict {
			t.Fatalf("name comparison %s, want conflict", got)
		}
		ex := conflictExhibit(t, row, "name")
		if ex.Pinned || ex.IncomingNegative || ex.IncomingDisplay != "James Robins" {
			t.Fatalf("name conflict exhibit %+v", ex)
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
		row := rowFor(f.propose(src.ID, []graphalign.Fixed{{SubjectID: incoming.ID, HandleID: handle}}), incoming.ID)
		if got := comparisonOutcome(t, row, "sex_at_birth"); got != match.OutcomeConflict {
			t.Fatalf("sex comparison %s, want conflict", got)
		}
		ex := conflictExhibit(t, row, "sex_at_birth")
		if ex.Pinned || len(ex.IncomingObservationID) != 0 || !ex.IncomingNegative || ex.IncomingDisplay != "Male" || ex.MemberDisplay != "Male" {
			t.Fatalf("sex conflict exhibit %+v", ex)
		}
		hasName := false
		for _, e := range row.Exhibits {
			hasName = hasName || e.Property.Key == "name"
		}
		if !hasName {
			t.Fatalf("expected the agreeing name exhibit beside the conflict: %+v", row.Exhibits)
		}
	})
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
	row := rowFor(f.propose(src.ID, []graphalign.Fixed{{SubjectID: incoming.ID, HandleID: handle}}), incoming.ID)
	if got := comparisonOutcome(t, row, "toponym"); got != match.OutcomeAgree {
		t.Fatalf("toponym comparison %s, want agree", got)
	}
}

// A name Observation with parts and no Form still shows on the sheet. The
// writers require a Form, so the fixture blanks it after the names are filed.
func TestExhibitNameWithoutForm(t *testing.T) {
	f := newFixture(t)
	handle := f.promote(f.person(f.source, f.artifact, "Mary Smith")).Entity.ID
	src, art := f.newSource("Register")
	incoming := f.person(src, art, "Mary Smith")
	db, err := f.c.DB()
	must(t, err)
	_, err = db.Exec(`UPDATE name_values SET form = ''`)
	must(t, err)

	row := rowFor(f.propose(src.ID, []graphalign.Fixed{{SubjectID: incoming.ID, HandleID: handle}}), incoming.ID)
	var found bool
	for _, ex := range row.Exhibits {
		if ex.Property.Key == "name" && len(ex.IncomingObservationID) > 0 {
			found = true
			if ex.IncomingDisplay != "Mary Smith" || ex.MemberDisplay != "Mary Smith" {
				t.Fatalf("name exhibit %+v", ex)
			}
		}
	}
	if !found {
		t.Fatalf("no name exhibit in %+v", row.Exhibits)
	}
}

// provisionalOnly leaves a handle with no accepted member claim, so Propose
// builds no exhibits for a row fixed to it.
func (f *fixture) provisionalOnly(entityID []byte) {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	_, err = db.Exec(`UPDATE identity_claims SET status = 'provisional' WHERE entity_id = ?`, entityID)
	must(f.t, err)
	f.c.Graph().Drop()
}

// A row with no exhibit still shows the two values behind each outcome.
func TestProposeRowWithoutExhibitsShowsValues(t *testing.T) {
	f := newFixture(t)
	sex := func(src sources.Source, art artifacts.Artifact) subjects.Subject {
		s := f.person(src, art, "Mary Smith")
		f.cite(art, s, observations.Input{PropertyID: f.prop("sex_at_birth").ID, ValueTermID: f.term("sex_at_birth", "female").ID})
		return s
	}
	handle := f.promote(sex(f.source, f.artifact)).Entity.ID
	f.provisionalOnly(handle)

	src, art := f.newSource("Register")
	incoming := sex(src, art)
	row := rowFor(f.propose(src.ID, []graphalign.Fixed{{SubjectID: incoming.ID, HandleID: handle}}), incoming.ID)

	var want int
	for _, c := range row.Comparisons {
		if c.Outcome != match.OutcomeUnknown {
			want++
		}
	}
	if want != 2 || len(row.Exhibits) != want {
		t.Fatalf("%d exhibits for %d spoken comparisons: %+v", len(row.Exhibits), want, row.Exhibits)
	}
	shown := map[string][2]string{}
	for _, ex := range row.Exhibits {
		if ex.Pinned || len(ex.IncomingObservationID) != 0 || len(ex.MemberObservationID) != 0 || len(ex.GroupSubjectID) != 0 {
			t.Fatalf("a ref line is never pinned or tied to an Observation: %+v", ex)
		}
		if got := comparisonOutcome(t, row, ex.Property.Key); got != ex.Outcome {
			t.Fatalf("%s line outcome %s, comparison %s", ex.Property.Key, ex.Outcome, got)
		}
		shown[ex.Property.Key] = [2]string{ex.IncomingDisplay, ex.MemberDisplay}
	}
	if shown["name"] != [2]string{"Mary Smith", "Mary Smith"} || shown["sex_at_birth"] != [2]string{"Female", "Female"} {
		t.Fatalf("values %v", shown)
	}
	if _, ok := shown["gender"]; ok {
		t.Fatalf("an unknown comparison has no line: %v", shown)
	}
}

func birthEvent(f *fixture, src sources.Source, art artifacts.Artifact, year int, polarity string) subjects.Subject {
	b := f.bare(src, "event")
	f.cite(art, b,
		observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID},
		observations.Input{PropertyID: f.prop("date").ID, Polarity: polarity,
			Date: &datevalue.Value{Kind: datevalue.KindPoint, Calendar: "gregorian", StartYear: &year}},
	)
	return b
}

func dateLine(t *testing.T, row graphalign.Row) graphalign.Exhibit {
	t.Helper()
	for _, ex := range row.Exhibits {
		if ex.Property.Key == "date" && len(ex.GroupSubjectID) == 0 {
			return ex
		}
	}
	t.Fatalf("no date line in %+v", row.Exhibits)
	return graphalign.Exhibit{}
}

func TestProposeEventWithoutExhibitsShowsDates(t *testing.T) {
	f := newFixture(t)
	handle := f.promote(birthEvent(f, f.source, f.artifact, 1890, "")).Entity.ID
	f.provisionalOnly(handle)

	src, art := f.newSource("Obituary")
	incoming := birthEvent(f, src, art, 1890, "")
	row := rowFor(f.propose(src.ID, []graphalign.Fixed{{SubjectID: incoming.ID, HandleID: handle}}), incoming.ID)
	ex := dateLine(t, row)
	if ex.IncomingDate == nil || ex.MemberDate == nil || ex.Pinned {
		t.Fatalf("date line %+v", ex)
	}
	for _, e := range row.Exhibits {
		if e.Property.Key == "event_type" && (e.IncomingDisplay != "Birth" || e.MemberDisplay != "Birth") {
			t.Fatalf("event_type line %+v", e)
		}
	}
}

func TestProposeNegativeDateConflictLine(t *testing.T) {
	f := newFixture(t)
	handle := f.promote(birthEvent(f, f.source, f.artifact, 1890, "")).Entity.ID

	src, art := f.newSource("Obituary")
	incoming := birthEvent(f, src, art, 1890, observations.PolarityNegative)
	row := rowFor(f.propose(src.ID, []graphalign.Fixed{{SubjectID: incoming.ID, HandleID: handle}}), incoming.ID)
	ex := dateLine(t, row)
	if ex.Outcome != match.OutcomeConflict || !ex.IncomingNegative || ex.IncomingDate == nil || ex.MemberDate == nil {
		t.Fatalf("date line %+v", ex)
	}
}
