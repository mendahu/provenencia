package conclusiondetails_test

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/claimconfidencegrades"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/sourcecredibility"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
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
	t      *testing.T
	c      *database.Catalog
	typeID []byte
	arts   map[string]artifacts.Artifact
	person subjecttypes.Type
	name   properties.Property
	sex    properties.Property
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
	must(t, sourcecredibilitygrades.Install(c))
	must(t, claimconfidencegrades.Install(c))
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	must(t, err)
	f := &fixture{t: t, c: c, typeID: typeID, arts: map[string]artifacts.Artifact{}}
	f.person, err = subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
	must(t, err)
	f.name, err = properties.Lookup(c, "name", properties.OriginProvenencia)
	must(t, err)
	f.sex, err = properties.Lookup(c, "sex_at_birth", properties.OriginProvenencia)
	must(t, err)
	return f
}

func (f *fixture) source(title string) sources.Source {
	f.t.Helper()
	src, err := sources.Create(f.c, userID, sources.CreateInput{SourceTypeID: f.typeID, Title: title})
	must(f.t, err)
	art, err := artifacts.Create(f.c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	must(f.t, err)
	f.arts[string(src.ID)] = art
	return src
}

func (f *fixture) personOn(src sources.Source) subjects.Subject {
	f.t.Helper()
	s, err := subjects.Create(f.c, userID, subjects.CreateInput{SourceID: src.ID, SubjectTypeID: f.person.ID}, nil)
	must(f.t, err)
	return s
}

// name files a name Observation from "type=value|…" parts (or "form:…").
func (f *fixture) cite(s subjects.Subject, in observations.Input) observations.Observation {
	f.t.Helper()
	res, err := citations.CreateWithObservations(f.c, userID, citations.CreateInput{ArtifactID: f.arts[string(s.SourceID)].ID, LocatorJSON: locator}, []observations.Input{in})
	must(f.t, err)
	return res.Observations[0]
}

func (f *fixture) nameIn(s subjects.Subject, spec string) observations.Input {
	v := &namevalues.Value{}
	if form, ok := strings.CutPrefix(spec, "form:"); ok {
		v.Form = form
	} else {
		var words []string
		for i, item := range strings.Split(spec, "|") {
			typ, val, _ := strings.Cut(item, "=")
			v.Parts = append(v.Parts, namevalues.Part{Idx: i, Type: typ, Value: val})
			words = append(words, val)
		}
		v.Form = strings.Join(words, " ")
	}
	return observations.Input{SubjectID: s.ID, PropertyID: f.name.ID, Name: v}
}

func (f *fixture) promote(s subjects.Subject, onto []byte) []byte {
	f.t.Helper()
	res, err := promote.Save(f.c, userID, promote.Input{SubjectID: s.ID, EntityID: onto})
	must(f.t, err)
	return res.Entity.ID
}

func (f *fixture) detail(h []byte) conclusiondetails.Detail {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	d, err := conclusiondetails.ForEntity(db, h)
	must(f.t, err)
	return d
}

func field(d conclusiondetails.Detail, key string) conclusiondetails.Field {
	for _, f := range d.Fields {
		if f.PropertyKey == key {
			return f
		}
	}
	panic("no field " + key)
}

func outcomeOf(f conclusiondetails.Field, obsID []byte) conclusiondetails.Outcome {
	for _, o := range f.Outcomes {
		if bytes.Equal(o.ObservationID, obsID) {
			return o
		}
	}
	panic("no outcome")
}

func TestForEntity(t *testing.T) {
	f := newFixture(t)
	register, census, bible := f.source("Register"), f.source("Census"), f.source("Bible")
	a, b, c := f.personOn(register), f.personOn(census), f.personOn(bible)
	initial := f.cite(a, f.nameIn(a, "given=J.|surname=Robins"))
	full := f.cite(b, f.nameIn(b, "given=James|surname=Robins"))
	misspelt := f.cite(c, f.nameIn(c, "given=James|surname=Robbins"))
	bare := f.cite(a, f.nameIn(a, "form:James Robins"))
	h := f.promote(a, nil)
	f.promote(b, h)
	f.promote(c, h)

	d := f.detail(h)
	if !bytes.Equal(d.Entity.ID, h) || d.MemberCount != 3 {
		t.Fatalf("entity %+v, %d members", d.Entity, d.MemberCount)
	}
	name := field(d, "name")
	if name.State != autoreconcile.StateMerged || name.ValueType != properties.ValueTypeName || name.Label == "" {
		t.Fatalf("name field %+v", name)
	}
	if len(name.Values) != 2 || !name.Values[0].Displayed() || name.Values[0].Support != 2 ||
		name.Values[0].Value.Name.Form != "James Robins" || name.Values[1].Reason != "outvoted" {
		t.Fatalf("name values %+v", name.Values)
	}

	if o := outcomeOf(name, initial.ID); o.Reason != "folded" || o.ValueRank != 1 || o.SourceTitle != "Register" ||
		o.Recorded.Name == nil || o.Recorded.Name.Form != "J. Robins" || o.ObservationRef == "" || o.SubjectRef == "" {
		t.Fatalf("folded outcome %+v", o)
	}
	if o := outcomeOf(name, full.ID); o.Reason != "kept" || o.ValueRank != 1 || o.SourceTitle != "Census" ||
		len(o.ArtifactID) != 16 || o.Vote != (autoreconcile.Vote{}) {
		t.Fatalf("kept outcome %+v", o)
	}
	// Robins from the Register and the Census beat the Bible's Robbins.
	if o := outcomeOf(name, misspelt.ID); o.Reason != "outvoted" || o.ValueRank != 2 || o.Vote != (autoreconcile.Vote{Support: 2, Of: 3}) {
		t.Fatalf("outvoted outcome %+v", o)
	}
	if o := outcomeOf(name, bare.ID); o.Reason != "no_evidence" || o.ValueRank != 0 || o.Recorded.Name == nil || o.Recorded.Name.Form != "James Robins" {
		t.Fatalf("no-evidence outcome %+v", o)
	}
	// Outcomes are ordered by the value they went into; no value last.
	if last := name.Outcomes[len(name.Outcomes)-1]; last.ValueRank != 0 {
		t.Fatalf("outcome order %+v", name.Outcomes)
	}

	if sex := field(d, "sex_at_birth"); sex.State != autoreconcile.StateEmpty || len(sex.Values) != 0 || len(sex.Outcomes) != 0 {
		t.Fatalf("empty field %+v", sex)
	}
	if d.Fields[0].PropertyKey != "name" {
		t.Fatalf("binding order: first field %q", d.Fields[0].PropertyKey)
	}
}

func TestForEntityEvidence(t *testing.T) {
	f := newFixture(t)
	register, census, gazette := f.source("Register"), f.source("Census"), f.source("Gazette")
	g, err := sourcecredibilitygrades.Lookup(f.c, "low_trust", sourcecredibilitygrades.OriginProvenencia)
	must(t, err)
	_, err = sourcecredibility.Upsert(f.c, userID, sourcecredibility.UpsertInput{SourceID: register.ID, CredibilityGradeID: g.ID})
	must(t, err)
	hg, err := sourcecredibilitygrades.Lookup(f.c, "high_trust", sourcecredibilitygrades.OriginProvenencia)
	must(t, err)
	_, err = sourcecredibility.Upsert(f.c, userID, sourcecredibility.UpsertInput{SourceID: gazette.ID, CredibilityGradeID: hg.ID})
	must(t, err)

	a, b, c := f.personOn(register), f.personOn(census), f.personOn(gazette)
	weak := f.cite(a, f.nameIn(a, "given=Jake|surname=Robins"))
	f.cite(b, f.nameIn(b, "given=James|surname=Robins"))
	denied := f.cite(b, f.nameIn(b, "given=James|surname=Robbins"))
	neg := f.nameIn(c, "given=James|surname=Robbins")
	neg.Polarity = observations.PolarityNegative
	negative := f.cite(c, neg)
	h := f.promote(a, nil)
	f.promote(b, h)
	f.promote(c, h)

	name := field(f.detail(h), "name")
	// Provenance is the auto-reconciler's own view: weak from the Source alone.
	if o := outcomeOf(name, weak.ID); o.Reason != "weak" || o.CredibilityKey != "low_trust" ||
		!o.Provenance.Weak() || o.Provenance.Credibility >= 0 || o.Provenance.Uncertain || o.Provenance.ClaimConfidence != 0 {
		t.Fatalf("weak outcome %+v", o)
	}
	if o := outcomeOf(name, denied.ID); o.Reason != "denied" || !bytes.Equal(o.DeniedBy, negative.ID) {
		t.Fatalf("denied outcome %+v", o)
	}
	if o := outcomeOf(name, negative.ID); o.Reason != "against" || o.CredibilityKey != "high_trust" || o.Provenance.Credibility <= 0 {
		t.Fatalf("negative outcome %+v", o)
	}
	var against int
	for _, v := range name.Values {
		against += v.Against
	}
	if against != 1 {
		t.Fatalf("against %+v", name.Values)
	}
}

func TestForEntityNotFound(t *testing.T) {
	f := newFixture(t)
	db, err := f.c.DB()
	must(t, err)
	for _, id := range [][]byte{nil, []byte("short"), bytes.Repeat([]byte{9}, 16)} {
		if _, err := conclusiondetails.ForEntity(db, id); !errors.Is(err, conclusiondetails.ErrNotFound) {
			t.Fatalf("%x: %v", id, err)
		}
	}
	a := f.personOn(f.source("Register"))
	h := f.promote(a, nil)
	other := f.promote(f.personOn(f.source("Census")), nil)
	_, err = db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, other, h)
	must(t, err)
	if _, err := conclusiondetails.ForEntity(db, h); !errors.Is(err, conclusiondetails.ErrNotFound) {
		t.Fatalf("merged: %v", err)
	}
}

// countingQuerier counts the statements the composer sends.
type countingQuerier struct {
	db *sql.DB
	n  int
}

func (c *countingQuerier) Query(query string, args ...any) (*sql.Rows, error) {
	c.n++
	return c.db.Query(query, args...)
}

func TestForEntityQueryCountIsConstant(t *testing.T) {
	f := newFixture(t)
	small := f.personOn(f.source("Register"))
	f.cite(small, f.nameIn(small, "given=Ann|surname=Lee"))
	hs := f.promote(small, nil)

	big := f.personOn(f.source("Census"))
	for i := 0; i < 12; i++ {
		f.cite(big, f.nameIn(big, fmt.Sprintf("given=Name%d|surname=Lee", i)))
	}
	f.cite(big, observations.Input{SubjectID: big.ID, PropertyID: f.sex.ID, ValueTermID: termID(t, f, "male")})
	hb := f.promote(big, nil)

	db, err := f.c.DB()
	must(t, err)
	one := &countingQuerier{db: db}
	_, err = conclusiondetails.ForEntity(one, hs)
	must(t, err)
	many := &countingQuerier{db: db}
	_, err = conclusiondetails.ForEntity(many, hb)
	must(t, err)
	if one.n != many.n {
		t.Fatalf("queries: %d for a small Person, %d for a big one", one.n, many.n)
	}
}

func termID(t *testing.T, f *fixture, key string) []byte {
	t.Helper()
	db, err := f.c.DB()
	must(t, err)
	var id []byte
	must(t, db.QueryRow(`SELECT id FROM property_terms WHERE property_id = ? AND key = ?`, f.sex.ID, key).Scan(&id))
	return id
}
