package autoreconciler_test

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/claimconfidencegrades"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/database/sourcecredibility"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/valuecodec"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

const locator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

// fixture is a catalog with one Source and Artifact, the seeded vocabulary,
// and a user integer Property ("age") bound to person.
type fixture struct {
	t        *testing.T
	c        *database.Catalog
	source   sources.Source                // the first Source ("Register")
	artifact artifacts.Artifact            // its Artifact
	other    sources.Source                // a second Source ("Census"), for independent evidence
	typeID   []byte                        // the book Source type, for newSource
	arts     map[string]artifacts.Artifact // Source id → its Artifact
	types    map[string]subjecttypes.Type
	props    map[string]properties.Property
	terms    map[string]propertyterms.Term
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	r, err := ref.Mint(ref.PrefixUser)
	must(t, err)
	must(t, users.Upsert(c, userID, "Tester", r))
	must(t, subjectvocab.Install(c))
	must(t, sourcecredibilitygrades.Install(c))
	must(t, claimconfidencegrades.Install(c))
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	must(t, err)
	f := &fixture{t: t, c: c, typeID: typeID, arts: map[string]artifacts.Artifact{},
		types: map[string]subjecttypes.Type{}, props: map[string]properties.Property{}, terms: map[string]propertyterms.Term{}}
	f.source = f.newSource("Register")
	f.artifact = f.arts[string(f.source.ID)]
	f.other = f.newSource("Census")
	for _, k := range []string{"person", "event", "place"} {
		st, err := subjecttypes.Lookup(c, k, subjecttypes.OriginProvenencia)
		must(t, err)
		f.types[k] = st
	}
	for _, k := range []string{"name", "sex_at_birth", "event_type", "date", "toponym"} {
		p, err := properties.Lookup(c, k, properties.OriginProvenencia)
		must(t, err)
		f.props[k] = p
	}
	ageID, err := properties.Upsert(c, properties.Property{Key: "age", Origin: properties.OriginUser, Label: "Age", ValueType: properties.ValueTypeInteger})
	must(t, err)
	f.props["age"] = properties.Property{ID: ageID, Key: "age", ValueType: properties.ValueTypeInteger}
	db, err := c.DB()
	must(t, err)
	_, err = db.Exec(`INSERT INTO subject_type_properties (subject_type_id, property_id, sort_order) VALUES (?, ?, 99)`, f.types["person"].ID, ageID)
	must(t, err)
	for _, k := range []string{"female", "male"} {
		term, err := propertyterms.Lookup(c, f.props["sex_at_birth"].ID, k, propertyterms.OriginProvenencia)
		must(t, err)
		f.terms[k] = term
	}
	for _, k := range []string{"birth", "death"} {
		term, err := propertyterms.Lookup(c, f.props["event_type"].ID, k, propertyterms.OriginProvenencia)
		must(t, err)
		f.terms[k] = term
	}
	return f
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// newSource is another Source with one Artifact: independent evidence, a
// separate vote for the auto-reconciler's majority.
func (f *fixture) newSource(title string) sources.Source {
	f.t.Helper()
	src, err := sources.Create(f.c, userID, sources.CreateInput{SourceTypeID: f.typeID, Title: title})
	must(f.t, err)
	art, err := artifacts.Create(f.c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	must(f.t, err)
	f.arts[string(src.ID)] = art
	return src
}

func (f *fixture) bind(kind string, propertyID []byte) {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	_, err = db.Exec(`INSERT INTO subject_type_properties (subject_type_id, property_id, sort_order) VALUES (?, ?, 98)`, f.types[kind].ID, propertyID)
	must(f.t, err)
}

func (f *fixture) subject(kind string) subjects.Subject { return f.subjectOn(f.source, kind) }

func (f *fixture) subjectOn(src sources.Source, kind string) subjects.Subject {
	f.t.Helper()
	s, err := subjects.Create(f.c, userID, subjects.CreateInput{SourceID: src.ID, SubjectTypeID: f.types[kind].ID}, nil)
	must(f.t, err)
	return s
}

// cite files the Observations on one new Citation, on the Artifact of the
// first one's Subject's Source.
func (f *fixture) cite(in ...observations.Input) []observations.Observation {
	f.t.Helper()
	art := f.artifact
	if len(in) > 0 {
		s, err := subjects.Get(f.c, in[0].SubjectID)
		must(f.t, err)
		art = f.arts[string(s.SourceID)]
	}
	res, err := citations.CreateWithObservations(f.c, userID, citations.CreateInput{ArtifactID: art.ID, LocatorJSON: locator}, in)
	must(f.t, err)
	return res.Observations
}

// credibility assesses a Source with a seeded grade key.
func (f *fixture) credibility(src sources.Source, key string) error {
	g, err := sourcecredibilitygrades.Lookup(f.c, key, sourcecredibilitygrades.OriginProvenencia)
	must(f.t, err)
	_, err = sourcecredibility.Upsert(f.c, userID, sourcecredibility.UpsertInput{SourceID: src.ID, CredibilityGradeID: g.ID})
	return err
}

// certainty sets a Citation's transcription certainty.
func (f *fixture) certainty(citationID []byte, uncertain bool) error {
	_, err := citations.Update(f.c, userID, citationID, citations.CitationFieldsInput{LocatorJSON: locator, TranscriptionUncertain: uncertain})
	return err
}

// grade is a seeded claim confidence grade's id.
func (f *fixture) grade(key string) []byte {
	g, err := claimconfidencegrades.Lookup(f.c, key, claimconfidencegrades.OriginProvenencia)
	must(f.t, err)
	return g.ID
}

func negative(in observations.Input) observations.Input {
	in.Polarity = observations.PolarityNegative
	return in
}

func place(f *fixture, s subjects.Subject, x, y int64) error {
	_, err := subjectpositions.Set(f.c, s.ID, x, y)
	return err
}

func mustProp(f *fixture, key string) properties.Property {
	f.t.Helper()
	p, err := properties.Lookup(f.c, key, properties.OriginProvenencia)
	must(f.t, err)
	return p
}

func associationOf(f *fixture, typeKey string) []byte {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	var id []byte
	must(f.t, db.QueryRow(`SELECT e.id FROM canonical_entities e
		JOIN subject_types st ON st.id = e.subject_type_id
		WHERE st.key = ? AND st.origin = 'provenencia'`, typeKey).Scan(&id))
	return id
}

func (f *fixture) promote(s subjects.Subject) []byte {
	f.t.Helper()
	res, err := promote.Save(f.c, userID, promote.Input{SubjectID: s.ID})
	must(f.t, err)
	return res.Entity.ID
}

func (f *fixture) join(s subjects.Subject, entityID []byte) {
	f.t.Helper()
	_, err := promote.Save(f.c, userID, promote.Input{SubjectID: s.ID, EntityID: entityID})
	must(f.t, err)
}

// nameIn is a name Observation from "type=value" parts separated by "|"
// (form: the values in order), or "form:…" for a name with no parts.
func nameIn(s subjects.Subject, p properties.Property, spec string) observations.Input {
	return observations.Input{SubjectID: s.ID, PropertyID: p.ID, Name: nameValue(spec)}
}

func nameValue(spec string) *namevalues.Value {
	if form, ok := strings.CutPrefix(spec, "form:"); ok {
		return &namevalues.Value{Form: form}
	}
	v := &namevalues.Value{}
	var words []string
	for i, item := range strings.Split(spec, "|") {
		typ, val, _ := strings.Cut(item, "=")
		v.Parts = append(v.Parts, namevalues.Part{Idx: i, Type: typ, Value: val})
		words = append(words, val)
	}
	v.Form = strings.Join(words, " ")
	return v
}

func textIn(s subjects.Subject, p properties.Property, v string) observations.Input {
	return observations.Input{SubjectID: s.ID, PropertyID: p.ID, ValueText: v, HasText: true}
}

func termIn(s subjects.Subject, p properties.Property, term propertyterms.Term) observations.Input {
	return observations.Input{SubjectID: s.ID, PropertyID: p.ID, ValueTermID: term.ID}
}

func intIn(s subjects.Subject, p properties.Property, n int64) observations.Input {
	return observations.Input{SubjectID: s.ID, PropertyID: p.ID, ValueInteger: n, HasInteger: true}
}

func ip(n int) *int { return &n }

func dateIn(s subjects.Subject, p properties.Property, y int, m, d *int) observations.Input {
	return observations.Input{SubjectID: s.ID, PropertyID: p.ID, Date: &datevalues.Value{
		Kind: datevalues.KindPoint, Calendar: "gregorian", StartYear: ip(y), StartMonth: m, StartDay: d,
	}}
}

// row is one cache row as stored.
type row struct {
	EntityID, PropertyID []byte
	Rank                 int
	Text                 *string
	Integer              *int64
	TermID               []byte
	Date, Name           []byte
	DateLo, DateHi       *int
	SortKey              *string
	Support, Against     int
	Reason               string
	ValueEntityID        []byte
}

func snapshot(t *testing.T, q autoreconciler.Querier) []row {
	t.Helper()
	rows, err := q.Query(`SELECT entity_id, property_id, rank, value_text, value_integer, value_term_id,
		value_date, value_name, date_lo, date_hi, sort_key, support, against, reason, value_entity_id
		FROM auto_reconciler_values
		ORDER BY entity_id, property_id, rank`)
	must(t, err)
	defer rows.Close()
	var out []row
	for rows.Next() {
		var r row
		must(t, rows.Scan(&r.EntityID, &r.PropertyID, &r.Rank, &r.Text, &r.Integer, &r.TermID, &r.Date, &r.Name, &r.DateLo, &r.DateHi, &r.SortKey, &r.Support, &r.Against, &r.Reason, &r.ValueEntityID))
		out = append(out, r)
	}
	must(t, rows.Err())
	return out
}

func (f *fixture) rows() []row {
	db, err := f.c.DB()
	must(f.t, err)
	return snapshot(f.t, db)
}

func rowsFor(all []row, entityID, propertyID []byte) []row {
	var out []row
	for _, r := range all {
		if bytes.Equal(r.EntityID, entityID) && bytes.Equal(r.PropertyID, propertyID) {
			out = append(out, r)
		}
	}
	return out
}

// rebuiltSnapshot rebuilds inside a rolled-back transaction, so upkeep state
// is left as it was.
// outcome is one auto_reconciler_outcomes row as stored.
type outcome struct {
	EntityID, PropertyID, ObservationID []byte
	Reason                              string
	ValueRank                           *int
	DeniedBy                            []byte
	VoteSupport, VoteTotal              *int
}

func outcomesSnapshot(t *testing.T, q autoreconciler.Querier) []outcome {
	t.Helper()
	rows, err := q.Query(`SELECT entity_id, property_id, observation_id, reason, value_rank, denied_by,
			vote_support, vote_total
		FROM auto_reconciler_outcomes ORDER BY entity_id, property_id, observation_id`)
	must(t, err)
	defer rows.Close()
	var out []outcome
	for rows.Next() {
		var o outcome
		must(t, rows.Scan(&o.EntityID, &o.PropertyID, &o.ObservationID, &o.Reason, &o.ValueRank, &o.DeniedBy,
			&o.VoteSupport, &o.VoteTotal))
		out = append(out, o)
	}
	must(t, rows.Err())
	return out
}

func (f *fixture) outcomes() []outcome {
	db, err := f.c.DB()
	must(f.t, err)
	return outcomesSnapshot(f.t, db)
}

func (f *fixture) rebuiltSnapshot() []row {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	tx, err := db.Begin()
	must(f.t, err)
	defer func() { _ = tx.Rollback() }()
	must(f.t, autoreconciler.Rebuild(tx))
	return snapshot(f.t, tx)
}

func TestReconciledValues(t *testing.T) {
	f := newFixture(t)
	person := f.subject("person")
	f.cite(nameIn(person, f.props["name"], "given=James|surname=Robins"), termIn(person, f.props["sex_at_birth"], f.terms["male"]), intIn(person, f.props["age"], 34))
	f.cite(nameIn(person, f.props["name"], "given=james|surname=robins."), intIn(person, f.props["age"], 34))
	jim := f.cite(nameIn(person, f.props["name"], "given=Jim|surname=Robins"))[0]

	if got := f.rows(); len(got) != 0 {
		t.Fatalf("unpromoted subject cached %d rows", len(got))
	}
	per := f.promote(person)
	all := f.rows()

	names := rowsFor(all, per, f.props["name"].ID)
	// James, james and Jim reconcile into one name: Jim is a different given
	// name, not a misspelling, so it isn't outvoted (S9-13b). All three are
	// cited from the Register, so they are one Source's support (S9-14).
	if len(names) != 1 || names[0].Support != 1 {
		t.Fatalf("name values %+v", names)
	}
	top, err := valuecodec.UnmarshalName(names[0].Name)
	must(t, err)
	if top.Form != "James Jim Robins" || *names[0].SortKey != "james jim robins" {
		t.Fatalf("rank 1 %q sort %q", top.Form, *names[0].SortKey)
	}
	sex := rowsFor(all, per, f.props["sex_at_birth"].ID)
	if len(sex) != 1 || !bytes.Equal(sex[0].TermID, f.terms["male"].ID) || sex[0].SortKey != nil {
		t.Fatalf("sex %+v", sex)
	}
	age := rowsFor(all, per, f.props["age"].ID)
	if len(age) != 1 || *age[0].Integer != 34 || age[0].Support != 1 { // two Observations, one Source
		t.Fatalf("age %+v", age)
	}

	t.Run("observation update", func(t *testing.T) {
		_, err := observations.Update(f.c, userID, observations.Input{
			ID: jim.ID, SubjectID: person.ID, PropertyID: f.props["name"].ID, Name: nameValue("given=James|surname=Robins"),
		})
		must(t, err)
		names := rowsFor(f.rows(), per, f.props["name"].ID)
		if len(names) != 1 || names[0].Support != 1 {
			t.Fatalf("after update %+v", names)
		}
	})
	t.Run("observation add", func(t *testing.T) {
		obs, err := observations.AddToCitation(f.c, userID, jim.CitationID, []observations.Input{intIn(person, f.props["age"], 35)})
		must(t, err)
		age := rowsFor(f.rows(), per, f.props["age"].ID)
		if len(age) != 2 || *age[1].Integer != 35 {
			t.Fatalf("after add %+v", age)
		}
		must(t, observations.Delete(f.c, userID, obs[0].ID))
		if age := rowsFor(f.rows(), per, f.props["age"].ID); len(age) != 1 {
			t.Fatalf("after delete %+v", age)
		}
	})

	t.Run("event and place values", func(t *testing.T) {
		event := f.subject("event")
		f.cite(termIn(event, f.props["event_type"], f.terms["birth"]), dateIn(event, f.props["date"], 1985, ip(5), nil))
		f.cite(dateIn(event, f.props["date"], 1985, ip(5), ip(14)))
		place := f.subject("place")
		f.cite(textIn(place, f.props["toponym"], "Upper Canada"), textIn(place, f.props["toponym"], "U.C."))
		evt, plc := f.promote(event), f.promote(place)
		all := f.rows()
		dates := rowsFor(all, evt, f.props["date"].ID)
		if len(dates) != 1 || dates[0].SortKey == nil || dates[0].DateLo == nil || dates[0].DateHi == nil || *dates[0].DateLo != *dates[0].DateHi {
			t.Fatalf("dates %+v", dates)
		}
		d, err := valuecodec.UnmarshalDate(dates[0].Date)
		must(t, err)
		if *d.StartYear != 1985 || *d.StartMonth != 5 || d.StartDay == nil || *d.StartDay != 14 {
			t.Fatalf("rank 1 date %+v", d)
		}
		tops := rowsFor(all, plc, f.props["toponym"].ID)
		if len(tops) != 2 || *tops[0].Text != "Upper Canada" || *tops[0].SortKey != "upper canada" {
			t.Fatalf("toponyms %+v", tops)
		}
	})

	t.Run("subject delete leaves no rows", func(t *testing.T) {
		lone := f.subject("person")
		obs := f.cite(nameIn(lone, f.props["name"], "given=Mary|surname=Smith"))
		h := f.promote(lone)
		if len(rowsFor(f.rows(), h, f.props["name"].ID)) != 1 {
			t.Fatal("promoted name not cached")
		}
		must(t, observations.Delete(f.c, userID, obs[0].ID))
		must(t, subjects.Delete(f.c, userID, lone.ID))
		for _, r := range f.rows() {
			if bytes.Equal(r.EntityID, h) {
				t.Fatalf("emptied handle kept %+v", r)
			}
		}
	})

	f.assertUpkeepEqualsRebuild("TestAutoReconciledValues")
}

// Every value stays cached with its reason (S9-13). On a single-valued
// Property, a majority still outvotes a distinct spelling.
func TestReconciledValuesKeepEveryValue(t *testing.T) {
	f := newFixture(t)
	aliasID, err := properties.Upsert(f.c, properties.Property{
		Key: "alias", Origin: properties.OriginUser, Label: "Alias", ValueType: properties.ValueTypeText,
	})
	must(t, err)
	alias := properties.Property{ID: aliasID, Key: "alias", ValueType: properties.ValueTypeText, Cardinality: properties.CardinalitySingle}
	f.bind("place", aliasID)
	a, b, c := f.subject("place"), f.subjectOn(f.other, "place"), f.subjectOn(f.newSource("Gazette"), "place")
	f.cite(textIn(a, alias, "York"))
	f.cite(textIn(b, alias, "york"))
	f.cite(textIn(c, alias, "Muddy York"))
	h := f.promote(a)
	f.join(b, h)
	f.join(c, h)
	rows := rowsFor(f.rows(), h, alias.ID)
	if len(rows) != 2 || rows[0].Reason != "kept" || rows[0].Support != 2 || *rows[0].Text != "York" ||
		rows[1].Reason != "outvoted" || *rows[1].Text != "Muddy York" {
		t.Fatalf("rows %+v", rows)
	}
	f.assertUpkeepEqualsRebuild("TestAutoReconciledValuesKeepEveryValue")
}

// A multiple Property keeps every distinct surviving value. Case-only
// duplicates still merge, and a value carried only by weak evidence drops.
func TestMultipleCardinalityKeepsDistinctValues(t *testing.T) {
	f := newFixture(t)
	gazette := f.newSource("Gazette")
	paper := f.newSource("Paper")
	must(t, f.credibility(gazette, "low_trust"))
	a := f.subject("place")
	b := f.subjectOn(f.other, "place")
	c := f.subjectOn(paper, "place")
	d := f.subjectOn(gazette, "place")
	f.cite(textIn(a, f.props["toponym"], "York"))
	f.cite(textIn(b, f.props["toponym"], "york"))
	f.cite(textIn(c, f.props["toponym"], "Toronto"))
	f.cite(textIn(d, f.props["toponym"], "Muddy York"))
	h := f.promote(a)
	f.join(b, h)
	f.join(c, h)
	f.join(d, h)
	rows := rowsFor(f.rows(), h, f.props["toponym"].ID)
	if len(rows) != 3 || rows[0].Reason != "kept" || *rows[0].Text != "York" || rows[0].Support != 2 ||
		rows[1].Reason != "kept" || *rows[1].Text != "Toronto" || rows[1].Support != 1 ||
		rows[2].Reason != "weak" || *rows[2].Text != "Muddy York" {
		t.Fatalf("rows %+v", rows)
	}
	f.assertUpkeepEqualsRebuild("TestMultipleCardinalityKeepsDistinctValues")
}

// Changing a user Property to multiple recomputes the handles that carry it.
// A seeded Property's cardinality stays fixed.
func TestCardinalityChangeRecomputes(t *testing.T) {
	f := newFixture(t)
	aliasID, err := properties.Upsert(f.c, properties.Property{
		Key: "alias", Origin: properties.OriginUser, Label: "Alias", ValueType: properties.ValueTypeText,
	})
	must(t, err)
	alias := properties.Property{ID: aliasID, ValueType: properties.ValueTypeText}
	f.bind("place", aliasID)
	a, b, c := f.subject("place"), f.subjectOn(f.other, "place"), f.subjectOn(f.newSource("Gazette"), "place")
	f.cite(textIn(a, alias, "York"))
	f.cite(textIn(b, alias, "york"))
	f.cite(textIn(c, alias, "Toronto"))
	h := f.promote(a)
	f.join(b, h)
	f.join(c, h)
	rows := rowsFor(f.rows(), h, alias.ID)
	if len(rows) != 2 || rows[1].Reason != "outvoted" {
		t.Fatalf("single %+v", rows)
	}
	_, err = properties.Update(f.c, userID, alias.ID, "Alias", properties.ValueTypeText, "", properties.CardinalityMultiple,
		func(tx *sql.Tx) error { return autoreconciler.RecomputePropertyTx(tx, alias.ID) })
	must(t, err)
	rows = rowsFor(f.rows(), h, alias.ID)
	if len(rows) != 2 || rows[0].Reason != "kept" || *rows[0].Text != "York" ||
		rows[1].Reason != "kept" || *rows[1].Text != "Toronto" {
		t.Fatalf("multiple %+v", rows)
	}
	f.assertUpkeepEqualsRebuild("TestCardinalityChangeRecomputes")

	top := f.props["toponym"]
	_, err = properties.Update(f.c, userID, top.ID, top.Label, top.ValueType, top.Description, properties.CardinalitySingle, nil)
	if !errors.Is(err, properties.ErrLocked) {
		t.Fatalf("seeded cardinality: %v", err)
	}
}

// The cache stores the reconciled name (S9-13b): one row for names that
// reconcile, parts only, and nothing for a name with no parts.
func TestReconciledNames(t *testing.T) {
	f := newFixture(t)

	t.Run("an initial expands into the full name", func(t *testing.T) {
		a, b := f.subject("person"), f.subjectOn(f.other, "person")
		f.cite(nameIn(a, f.props["name"], "given=J.|surname=Robins"))
		f.cite(nameIn(b, f.props["name"], "given=James|surname=Robins|suffix=Jr."))
		h := f.promote(a)
		f.join(b, h)
		names := rowsFor(f.rows(), h, f.props["name"].ID)
		if len(names) != 1 || names[0].Support != 2 || names[0].Rank != 1 {
			t.Fatalf("names %+v", names)
		}
		n, err := valuecodec.UnmarshalName(names[0].Name)
		must(t, err)
		var parts []string
		for _, p := range n.Parts {
			parts = append(parts, p.Type+"="+p.Value)
		}
		if got := strings.Join(parts, "|"); got != "given=James|surname=Robins|suffix=Jr." {
			t.Fatalf("parts %s", got)
		}
		if n.Form != "James Robins Jr." || *names[0].SortKey != "james robins jr" {
			t.Fatalf("form %q sort %q", n.Form, *names[0].SortKey)
		}
	})

	t.Run("a name with no parts is not cached", func(t *testing.T) {
		p := f.subject("person")
		f.cite(nameIn(p, f.props["name"], "form:James Robins"), intIn(p, f.props["age"], 3))
		h := f.promote(p)
		all := f.rows()
		if names := rowsFor(all, h, f.props["name"].ID); len(names) != 0 {
			t.Fatalf("names %+v", names)
		}
		if age := rowsFor(all, h, f.props["age"].ID); len(age) != 1 {
			t.Fatalf("age %+v", age)
		}
	})

	f.assertUpkeepEqualsRebuild("TestAutoReconciledNames")
}

// A stale cache rebuilds on open: old versions, and the 3 and 4 stamped by
// the closed PRs' builds, which must never read as current.
// The auto-reconciler weighs real evidence (S9-14): each Observation's
// Source, credibility, certainty, claim confidence and status, and polarity.
func TestReconciledEvidence(t *testing.T) {
	texts := func(f *fixture, h []byte) string {
		var out []string
		for _, r := range rowsFor(f.rows(), h, f.props["toponym"].ID) {
			out = append(out, *r.Text+":"+r.Reason)
		}
		return strings.Join(out, " ")
	}
	reasonOf := func(f *fixture, obsID []byte) outcome {
		for _, o := range f.outcomes() {
			if bytes.Equal(o.ObservationID, obsID) {
				return o
			}
		}
		f.t.Fatalf("no outcome for %x", obsID)
		return outcome{}
	}

	t.Run("a low-trust Source's value is weak beside a standard one", func(t *testing.T) {
		f := newFixture(t)
		must(t, f.credibility(f.source, "low_trust"))
		a, b := f.subject("place"), f.subjectOn(f.other, "place")
		f.cite(textIn(a, f.props["toponym"], "U.C."))
		f.cite(textIn(b, f.props["toponym"], "Upper Canada"))
		h := f.promote(a)
		f.join(b, h)
		if got := texts(f, h); got != "Upper Canada:kept U.C.:weak" {
			t.Fatalf("toponyms %s", got)
		}
		f.assertUpkeepEqualsRebuild("low trust")
	})

	t.Run("an uncertain transcription is weak", func(t *testing.T) {
		f := newFixture(t)
		a, b := f.subject("place"), f.subjectOn(f.other, "place")
		res, err := citations.CreateWithObservations(f.c, userID, citations.CreateInput{
			ArtifactID: f.artifact.ID, LocatorJSON: locator, TranscriptionUncertain: true,
		}, []observations.Input{textIn(a, f.props["toponym"], "Yorke")})
		must(t, err)
		_ = res
		f.cite(textIn(b, f.props["toponym"], "York"))
		h := f.promote(a)
		f.join(b, h)
		if got := texts(f, h); got != "York:kept Yorke:weak" {
			t.Fatalf("toponyms %s", got)
		}
	})

	t.Run("two Observations from one Source are one vote", func(t *testing.T) {
		f := newFixture(t)
		a, b := f.subject("place"), f.subjectOn(f.other, "place")
		f.cite(textIn(a, f.props["toponym"], "York"))
		f.cite(textIn(a, f.props["toponym"], "York"))
		f.cite(textIn(b, f.props["toponym"], "Toronto"))
		h := f.promote(a)
		f.join(b, h)
		rows := rowsFor(f.rows(), h, f.props["toponym"].ID)
		if len(rows) != 2 || rows[0].Reason != "kept" || rows[1].Reason != "kept" || rows[0].Support != 1 || rows[1].Support != 1 {
			t.Fatalf("one Source's two records outvoted another Source: %+v", rows)
		}
	})

	t.Run("a stronger negative denies a value", func(t *testing.T) {
		f := newFixture(t)
		must(t, f.credibility(f.other, "high_trust"))
		a, b, c := f.subject("place"), f.subjectOn(f.other, "place"), f.subjectOn(f.newSource("Gazette"), "place")
		york := f.cite(textIn(a, f.props["toponym"], "York"))[0]
		denial := f.cite(negative(textIn(b, f.props["toponym"], "York")))[0]
		f.cite(textIn(c, f.props["toponym"], "Toronto"))
		h := f.promote(a)
		f.join(b, h)
		f.join(c, h)
		if got := texts(f, h); got != "Toronto:kept York:denied" {
			t.Fatalf("toponyms %s", got)
		}
		rows := rowsFor(f.rows(), h, f.props["toponym"].ID)
		if rows[1].Against != 1 {
			t.Fatalf("against %+v", rows[1])
		}
		if o := reasonOf(f, york.ID); o.Reason != "denied" || !bytes.Equal(o.DeniedBy, denial.ID) || o.ValueRank == nil || *o.ValueRank != 2 {
			t.Fatalf("York outcome %+v", o)
		}
		if o := reasonOf(f, denial.ID); o.Reason != "against" {
			t.Fatalf("negative outcome %+v", o)
		}
		f.assertUpkeepEqualsRebuild("denied")
	})

	t.Run("a low-confidence claim's different given name is weak", func(t *testing.T) {
		f := newFixture(t)
		a, b := f.subject("person"), f.subjectOn(f.other, "person")
		f.cite(nameIn(a, f.props["name"], "given=Jake|surname=Robins"))
		f.cite(nameIn(b, f.props["name"], "given=James|surname=Robins"))
		res, err := promote.Save(f.c, userID, promote.Input{SubjectID: a.ID, ConfidenceGradeID: f.grade("low_confidence")})
		must(t, err)
		f.join(b, res.Entity.ID)
		names := rowsFor(f.rows(), res.Entity.ID, f.props["name"].ID)
		if len(names) != 2 || names[0].Reason != "kept" || names[1].Reason != "weak" {
			t.Fatalf("names %+v", names)
		}
		if n, err := valuecodec.UnmarshalName(names[0].Name); err != nil || n.Form != "James Robins" {
			t.Fatalf("displayed %+v %v", n, err)
		}
	})

	t.Run("a misspelt surname from one Source in three is outvoted", func(t *testing.T) {
		f := newFixture(t)
		a, b, c := f.subject("person"), f.subjectOn(f.other, "person"), f.subjectOn(f.newSource("Bible"), "person")
		f.cite(nameIn(a, f.props["name"], "given=James|surname=Robins"))
		f.cite(nameIn(b, f.props["name"], "given=James|surname=Robins"))
		robbins := f.cite(nameIn(c, f.props["name"], "given=James|surname=Robbins"))[0]
		h := f.promote(a)
		f.join(b, h)
		f.join(c, h)
		names := rowsFor(f.rows(), h, f.props["name"].ID)
		if len(names) != 2 || names[0].Support != 2 || names[1].Reason != "outvoted" {
			t.Fatalf("names %+v", names)
		}
		// The outcome keeps the vote that beat it; the others keep none.
		for _, o := range f.outcomes() {
			outvoted := bytes.Equal(o.ObservationID, robbins.ID)
			if outvoted && (o.VoteSupport == nil || *o.VoteSupport != 2 || o.VoteTotal == nil || *o.VoteTotal != 3) {
				t.Fatalf("outvoted outcome %+v", o)
			}
			if !outvoted && (o.VoteSupport != nil || o.VoteTotal != nil) {
				t.Fatalf("vote on %+v", o)
			}
		}
	})

	// Creating a claim recomputes its handle in the same transaction, whatever
	// its status: no manual recompute, and upkeep still equals rebuild.
	claimTests := []struct {
		name        string
		status      string
		wantTexts   string
		wantOutcome string // "" = the claimed member's Observation has no outcome
	}{
		{name: "a provisional member counts only as reasoning", status: identityclaims.StatusProvisional,
			wantTexts: "York:kept Toronto:provisional", wantOutcome: "provisional"},
		{name: "a rejected member does not count at all", status: identityclaims.StatusRejected,
			wantTexts: "York:kept"},
	}
	for _, tt := range claimTests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			a, b := f.subject("place"), f.subjectOn(f.other, "place")
			f.cite(textIn(a, f.props["toponym"], "York"))
			claimed := f.cite(textIn(b, f.props["toponym"], "Toronto"))[0]
			h := f.promote(a)
			_, err := identityclaims.Create(f.c, userID, identityclaims.CreateInput{SubjectID: b.ID, EntityID: h, Status: tt.status})
			must(t, err)
			if got := texts(f, h); got != tt.wantTexts {
				t.Fatalf("toponyms %q want %q", got, tt.wantTexts)
			}
			var got *outcome
			for _, o := range f.outcomes() {
				if bytes.Equal(o.ObservationID, claimed.ID) {
					got = &o
				}
			}
			switch {
			case tt.wantOutcome == "" && got != nil:
				t.Fatalf("a %s member's Observation was considered: %+v", tt.status, *got)
			case tt.wantOutcome != "" && (got == nil || got.Reason != tt.wantOutcome):
				t.Fatalf("outcome %+v want reason %q", got, tt.wantOutcome)
			}
			f.assertUpkeepEqualsRebuild(tt.status)
		})
	}

	t.Run("every Observation has an outcome; no evidence has no value", func(t *testing.T) {
		f := newFixture(t)
		p := f.subject("person")
		named := f.cite(nameIn(p, f.props["name"], "given=James|surname=Robins"))[0]
		bare := f.cite(nameIn(p, f.props["name"], "form:James Robins"))[0]
		h := f.promote(p)
		if o := reasonOf(f, named.ID); o.Reason != "kept" || o.ValueRank == nil || *o.ValueRank != 1 {
			t.Fatalf("named outcome %+v", o)
		}
		if o := reasonOf(f, bare.ID); o.Reason != "no_evidence" || o.ValueRank != nil {
			t.Fatalf("parts-less outcome %+v", o)
		}
		_ = h
	})
}

func TestEnsureCatalogRebuildsStaleVersion(t *testing.T) {
	for _, stale := range []int{0, 2, 3, 4, 5, 6, 7} {
		t.Run(fmt.Sprint("version ", stale), func(t *testing.T) {
			f := newFixture(t)
			p := f.subject("person")
			f.cite(nameIn(p, f.props["name"], "given=James|surname=Robins"))
			f.promote(p)
			want := f.rows()

			db, err := f.c.DB()
			must(t, err)
			_, err = db.Exec(`DELETE FROM auto_reconciler_values`)
			must(t, err)
			_, err = db.Exec(`UPDATE auto_reconciler_meta SET cache_version = ?`, stale)
			must(t, err)
			need, err := autoreconciler.NeedsRebuild(db)
			must(t, err)
			if !need {
				t.Fatal("stale version not detected")
			}
			must(t, autoreconciler.EnsureCatalog(f.c))
			if v, err := autoreconciler.StoredVersion(db); err != nil || v != autoreconciler.CacheVersion {
				t.Fatalf("version %d %v", v, err)
			}
			if got := f.rows(); !reflect.DeepEqual(got, want) {
				t.Fatalf("rebuilt %d rows, want %d", len(got), len(want))
			}
			need, err = autoreconciler.NeedsRebuild(db)
			must(t, err)
			if need {
				t.Fatal("rebuild did not store the version")
			}
		})
	}
}

func TestLoaderQueryCountIsConstant(t *testing.T) {
	f := newFixture(t)
	var handles [][]byte
	for i := 0; i < 60; i++ {
		p := f.subject("person")
		f.cite(nameIn(p, f.props["name"], fmt.Sprintf("given=Person|surname=P%d", i)), intIn(p, f.props["age"], int64(i)))
		handles = append(handles, f.promote(p))
	}
	db, err := f.c.DB()
	must(t, err)
	one, err := autoreconciler.LoadQueryCount(db, handles[:1])
	must(t, err)
	many, err := autoreconciler.LoadQueryCount(db, handles)
	must(t, err)
	if one != many {
		t.Fatalf("loader queries: %d for 1 handle, %d for %d", one, many, len(handles))
	}
}

// assertUpkeepEqualsRebuild fails when the incrementally maintained table
// differs from a full rebuild (taken in a rolled-back transaction).
func (f *fixture) assertUpkeepEqualsRebuild(at string) {
	f.t.Helper()
	if got, want := f.rows(), f.rebuiltSnapshot(); !reflect.DeepEqual(got, want) {
		f.t.Fatalf("%s: upkeep has %d rows, rebuild %d", at, len(got), len(want))
	}
	// The outcomes (the auto-reconciler's reasoning) must match a rebuild too.
	{
		db, err := f.c.DB()
		must(f.t, err)
		got := outcomesSnapshot(f.t, db)
		tx, err := db.Begin()
		must(f.t, err)
		must(f.t, autoreconciler.Rebuild(tx))
		want := outcomesSnapshot(f.t, tx)
		_ = tx.Rollback()
		if !reflect.DeepEqual(got, want) {
			f.t.Fatalf("%s: upkeep has %d outcomes, rebuild %d", at, len(got), len(want))
		}
	}
	// The handle search documents ride on the same upkeep (RecomputeTx), so
	// they must equal a full search rebuild too.
	db, err := f.c.DB()
	must(f.t, err)
	got := handleDocs(f.t, db)
	tx, err := db.Begin()
	must(f.t, err)
	defer func() { _ = tx.Rollback() }()
	must(f.t, searchindex.RebuildAll(tx))
	if want := handleDocs(f.t, tx); !reflect.DeepEqual(got, want) {
		f.t.Fatalf("%s: handle search docs differ from a rebuild:\n upkeep  %v\n rebuild %v", at, got, want)
	}
}

// handleDocs is every person / event / place search document, as text.
func handleDocs(t *testing.T, q autoreconciler.Querier) []string {
	t.Helper()
	rows, err := q.Query(`SELECT kind, entity_id, display_ref, display_title, title, ref, secondary
		FROM catalog_search_docs WHERE kind IN ('person', 'event', 'place') ORDER BY kind, entity_id`)
	must(t, err)
	defer rows.Close()
	var out []string
	for rows.Next() {
		var kind, id, dref, dtitle, title, ref, secondary string
		must(t, rows.Scan(&kind, &id, &dref, &dtitle, &title, &ref, &secondary))
		out = append(out, strings.Join([]string{kind, id, dref, dtitle, title, ref, secondary}, "|"))
	}
	must(t, rows.Err())
	return out
}

// Rebuild-equals-upkeep is the guard against silent upkeep misses (Gotcha 5).
// Two tests share it:
//
//   - TestRebuildEqualsUpkeep_SeededSequences generates long mixed write
//     sequences from FIXED seeds. It is deterministic: the same seed applies
//     the same operations in the same order on every run and machine, and a
//     failure names its seed and step. Never seed from time or add seeds
//     that vary per run; to explore new seeds, try them locally and promote
//     any failure as below.
//   - TestRebuildEqualsUpkeep_Scenarios holds named, hand-written sequences.
//     When a seed finds a bug, shrink its operations to the shortest sequence
//     that still fails and add it here, so the regression reads on its own and
//     does not depend on the generator.
//
// A new write path or trigger adds its operation to the generator and, where
// it has a characteristic sequence, a scenario.
//
// Subject delete recomputes the handles a member left and the handles whose
// members cited it. A participation filed by promoting both ends is a named
// scenario below; the generator does not mint bridges.

func TestRebuildEqualsUpkeep_Scenarios(t *testing.T) {
	scenarios := []struct {
		name string
		run  func(f *fixture)
	}{
		{"edit moves the rank-1 value, then its last member is deleted", func(f *fixture) {
			p := f.subject("person")
			obs := f.cite(nameIn(p, f.props["name"], "given=James|surname=Robins"), nameIn(p, f.props["name"], "given=Jim|surname=Robins"))
			f.promote(p)
			_, err := observations.Update(f.c, userID, observations.Input{
				ID: obs[1].ID, SubjectID: p.ID, PropertyID: f.props["name"].ID, Name: nameValue("given=james|surname=robins"),
			})
			must(f.t, err)
			f.assertUpkeepEqualsRebuild("after update")
			must(f.t, observations.Delete(f.c, userID, obs[0].ID))
		}},
		{"observations added before and after promote on two members", func(f *fixture) {
			a, b := f.subject("place"), f.subject("place")
			f.cite(textIn(a, f.props["toponym"], "York"))
			h := f.promote(a)
			f.cite(textIn(a, f.props["toponym"], "york"), textIn(b, f.props["toponym"], "Toronto"))
			f.promote(b)
			f.assertUpkeepEqualsRebuild("two handles")
			// Both from the Register: one value, one Source's support.
			if rows := rowsFor(f.rows(), h, f.props["toponym"].ID); len(rows) != 1 || rows[0].Support != 1 {
				f.t.Fatalf("York / york should be one value: %+v", rows)
			}
		}},
		{"joining a second member merges its name, then adds a given name", func(f *fixture) {
			a, b := f.subject("person"), f.subjectOn(f.other, "person")
			f.cite(nameIn(a, f.props["name"], "given=James|surname=Robins"))
			f.cite(nameIn(b, f.props["name"], "given=james|surname=robins"))
			h := f.promote(a)
			f.join(b, h)
			f.assertUpkeepEqualsRebuild("after join")
			names := rowsFor(f.rows(), h, f.props["name"].ID)
			if len(names) != 1 || names[0].Support != 2 {
				f.t.Fatalf("join should merge the names into one value of 2: %+v", names)
			}
			f.cite(nameIn(b, f.props["name"], "given=Jim|surname=Robins"))
			names = rowsFor(f.rows(), h, f.props["name"].ID)
			if len(names) != 1 {
				f.t.Fatalf("a member's new given name should join the one name: %+v", names)
			}
			if n, err := valuecodec.UnmarshalName(names[0].Name); err != nil || n.Form != "James Jim Robins" {
				f.t.Fatalf("name %+v %v", n, err)
			}
		}},
		{"deleting the full given name drops rank 1 back to the initial", func(f *fixture) {
			a, b := f.subject("person"), f.subject("person")
			obs := f.cite(nameIn(a, f.props["name"], "given=J.|surname=Robins"), nameIn(b, f.props["name"], "given=James|surname=Robins"))
			h := f.promote(a)
			f.join(b, h)
			f.assertUpkeepEqualsRebuild("after join")
			must(f.t, observations.Delete(f.c, userID, obs[1].ID))
			names := rowsFor(f.rows(), h, f.props["name"].ID)
			if len(names) != 1 || names[0].Support != 1 {
				f.t.Fatalf("names %+v", names)
			}
			n, err := valuecodec.UnmarshalName(names[0].Name)
			must(f.t, err)
			if n.Form != "J. Robins" {
				f.t.Fatalf("rank 1 %q, want J. Robins", n.Form)
			}
		}},
		{"lowering a Source's credibility drops its spelling", func(f *fixture) {
			a, b := f.subject("place"), f.subjectOn(f.other, "place")
			f.cite(textIn(a, f.props["toponym"], "U.C."))
			f.cite(textIn(b, f.props["toponym"], "Upper Canada"))
			h := f.promote(a)
			f.join(b, h)
			f.assertUpkeepEqualsRebuild("before")
			must(f.t, f.credibility(f.source, "low_trust"))
			f.assertUpkeepEqualsRebuild("after the Register fell")
			if rows := rowsFor(f.rows(), h, f.props["toponym"].ID); len(rows) != 2 || *rows[0].Text != "Upper Canada" || rows[1].Reason != "weak" {
				f.t.Fatalf("toponyms %+v", rows)
			}
		}},
		{"a stronger negative removes a value, and deleting it brings it back", func(f *fixture) {
			a, b, c := f.subject("place"), f.subjectOn(f.other, "place"), f.subjectOn(f.newSource("Gazette"), "place")
			f.cite(textIn(a, f.props["toponym"], "York"))
			f.cite(textIn(c, f.props["toponym"], "Toronto"))
			must(f.t, f.credibility(f.other, "high_trust"))
			h := f.promote(a)
			f.join(b, h)
			f.join(c, h)
			denial := f.cite(negative(textIn(b, f.props["toponym"], "York")))
			f.assertUpkeepEqualsRebuild("denied")
			must(f.t, observations.Delete(f.c, userID, denial[0].ID))
			f.assertUpkeepEqualsRebuild("restored")
			for _, r := range rowsFor(f.rows(), h, f.props["toponym"].ID) {
				if r.Reason != "kept" || r.Against != 0 {
					f.t.Fatalf("after the negative went: %+v", r)
				}
			}
		}},
		{"certainty toggled on a cited value", func(f *fixture) {
			a, b := f.subject("place"), f.subjectOn(f.other, "place")
			obs := f.cite(textIn(a, f.props["toponym"], "Yorke"))
			f.cite(textIn(b, f.props["toponym"], "York"))
			h := f.promote(a)
			f.join(b, h)
			must(f.t, f.certainty(obs[0].CitationID, true))
			f.assertUpkeepEqualsRebuild("uncertain")
			if rows := rowsFor(f.rows(), h, f.props["toponym"].ID); len(rows) != 2 || rows[1].Reason != "weak" {
				f.t.Fatalf("toponyms %+v", rows)
			}
			must(f.t, f.certainty(obs[0].CitationID, false))
		}},
		{"promoting both ends of a participation files it, then the bridge is deleted", func(f *fixture) {
			person, event := f.subject("person"), f.subject("event")
			must(f.t, place(f, person, 0, 0))
			must(f.t, place(f, event, 2, 0))
			personProp := mustProp(f, "person")
			eventProp := mustProp(f, "event")
			roleProp := mustProp(f, "role")
			role, err := propertyterms.Lookup(f.c, roleProp.ID, "subject", propertyterms.OriginProvenencia)
			must(f.t, err)
			bridge, err := connect.CreateCitedBridge(f.c, userID, connect.CreateInput{
				SourceID: f.source.ID, FromSubjectID: person.ID, ToSubjectID: event.ID, BridgeTypeKey: "participation",
				Citation: citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: locator},
				Observations: []observations.Input{
					{PropertyID: personProp.ID, ValueSubjectID: person.ID},
					{PropertyID: eventProp.ID, ValueSubjectID: event.ID},
					{PropertyID: roleProp.ID, ValueTermID: role.ID},
				},
			})
			must(f.t, err)
			personHandle := f.promote(person)
			f.promote(event)
			f.assertUpkeepEqualsRebuild("filed")
			rows := rowsFor(f.rows(), associationOf(f, "participation"), personProp.ID)
			if len(rows) != 1 || !bytes.Equal(rows[0].ValueEntityID, personHandle) {
				f.t.Fatalf("person edge %+v", rows)
			}
			must(f.t, subjects.Delete(f.c, userID, bridge.Subject.ID))
		}},
		{"emptied member subject deleted", func(f *fixture) {
			p := f.subject("event")
			obs := f.cite(dateIn(p, f.props["date"], 1985, ip(5), nil))
			f.promote(p)
			must(f.t, observations.Delete(f.c, userID, obs[0].ID))
			must(f.t, subjects.Delete(f.c, userID, p.ID))
		}},
	}
	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			f := newFixture(t)
			sc.run(f)
			f.assertUpkeepEqualsRebuild("end")
		})
	}
}

func TestRebuildEqualsUpkeep_SeededSequences(t *testing.T) {
	for _, seed := range []int64{1, 2, 3, 4} {
		t.Run(fmt.Sprint("seed ", seed), func(t *testing.T) {
			f := newFixture(t)
			rng := rand.New(rand.NewSource(seed))
			var subs []subjects.Subject
			var obs []observations.Observation
			handles := map[string][][]byte{}
			joins := 0
			kinds := []string{"person", "person", "event", "place"}
			credibilities := []string{"low_trust", "standard", "high_trust"}
			grades := []string{"", "low_confidence", "moderate", "high_confidence"}
			names := []string{"given=James|surname=Robins", "given=james|surname=robins", "given=Jim|surname=Robins",
				"given=J.|surname=Robins", "given=James|given=K.|surname=Robins", "surname=Robbins",
				"given=Mary|surname=Smith", "form:James Robins"}
			toponyms := []string{"York", "york", "Upper Canada", "Toronto"}

			value := func(s subjects.Subject) observations.Input {
				switch kindOf(f, s) {
				case "person":
					switch rng.Intn(3) {
					case 0:
						return nameIn(s, f.props["name"], names[rng.Intn(len(names))])
					case 1:
						return termIn(s, f.props["sex_at_birth"], f.terms[[]string{"female", "male"}[rng.Intn(2)]])
					default:
						return intIn(s, f.props["age"], int64(rng.Intn(4)-1))
					}
				case "event":
					if rng.Intn(2) == 0 {
						return termIn(s, f.props["event_type"], f.terms[[]string{"birth", "death"}[rng.Intn(2)]])
					}
					var day *int
					if rng.Intn(2) == 0 {
						day = ip(14)
					}
					return dateIn(s, f.props["date"], 1985, ip(5), day)
				default:
					return textIn(s, f.props["toponym"], toponyms[rng.Intn(len(toponyms))])
				}
			}
			tolerate := func(err error) {
				t.Helper()
				var ae *apperr.Error
				if err != nil && !(errors.As(err, &ae) && ae.Kind() != apperr.KindInternal) {
					t.Fatal(err)
				}
			}

			for step := 1; step <= 200; step++ {
				switch op := rng.Intn(13); {
				case op < 2 || len(subs) == 0:
					src := f.source
					if rng.Intn(2) == 0 {
						src = f.other
					}
					subs = append(subs, f.subjectOn(src, kinds[rng.Intn(len(kinds))]))
				case op < 5:
					s := subs[rng.Intn(len(subs))]
					obs = append(obs, f.cite(value(s), value(s))...)
				case op < 6:
					// Mint, or join a handle of the Subject's kind already minted.
					s := subs[rng.Intn(len(subs))]
					in := promote.Input{SubjectID: s.ID}
					if g := grades[rng.Intn(len(grades))]; g != "" {
						in.ConfidenceGradeID = f.grade(g)
					}
					if same := handles[kindOf(f, s)]; len(same) > 0 && rng.Intn(2) == 0 {
						in.EntityID = same[rng.Intn(len(same))]
					}
					res, err := promote.Save(f.c, userID, in)
					tolerate(err)
					switch {
					case err == nil && in.EntityID == nil:
						handles[kindOf(f, s)] = append(handles[kindOf(f, s)], res.Entity.ID)
					case err == nil:
						joins++
					}
				case op < 8 && len(obs) > 0:
					i := rng.Intn(len(obs))
					s := subjectByID(subs, obs[i].SubjectID)
					in := value(s)
					in.ID = obs[i].ID
					_, err := observations.Update(f.c, userID, in)
					tolerate(err)
				case op < 9 && len(obs) > 0:
					i := rng.Intn(len(obs))
					err := observations.Delete(f.c, userID, obs[i].ID)
					tolerate(err)
					if err == nil {
						obs = append(obs[:i], obs[i+1:]...)
					}
				case op == 10:
					src := f.source
					if rng.Intn(2) == 0 {
						src = f.other
					}
					tolerate(f.credibility(src, credibilities[rng.Intn(len(credibilities))]))
					// Checked at once: a later write to the handle would heal a miss.
					f.assertUpkeepEqualsRebuild(fmt.Sprintf("seed %d step %d credibility", seed, step))
				case op == 11 && len(obs) > 0:
					// Prefer a name or toponym Citation: certainty moves those most.
					o := obs[rng.Intn(len(obs))]
					for _, c := range obs {
						if (bytes.Equal(c.PropertyID, f.props["name"].ID) || bytes.Equal(c.PropertyID, f.props["toponym"].ID)) && rng.Intn(2) == 0 {
							o = c
							break
						}
					}
					tolerate(f.certainty(o.CitationID, rng.Intn(2) == 0))
					f.assertUpkeepEqualsRebuild(fmt.Sprintf("seed %d step %d certainty", seed, step))
				case op == 12:
					s := subs[rng.Intn(len(subs))]
					obs = append(obs, f.cite(negative(value(s)))...)
				default:
					i := rng.Intn(len(subs))
					err := subjects.Delete(f.c, userID, subs[i].ID)
					tolerate(err)
					if err == nil {
						subs = append(subs[:i], subs[i+1:]...)
					}
				}
				if step%10 == 0 {
					f.assertUpkeepEqualsRebuild(fmt.Sprintf("seed %d step %d", seed, step))
				}
			}
			if len(f.rows()) == 0 {
				t.Fatal("sequence never cached anything; the test is not exercising upkeep")
			}
			if joins == 0 {
				t.Fatal("sequence never joined an existing handle")
			}
			db, err := f.c.DB()
			must(t, err)
			if len(handleDocs(t, db)) == 0 {
				t.Fatal("sequence never indexed a handle; the search comparison is not exercised")
			}
		})
	}
}

func kindOf(f *fixture, s subjects.Subject) string {
	for k, st := range f.types {
		if bytes.Equal(st.ID, s.SubjectTypeID) {
			return k
		}
	}
	return ""
}

func subjectByID(subs []subjects.Subject, id []byte) subjects.Subject {
	for _, s := range subs {
		if bytes.Equal(s.ID, id) {
			return s
		}
	}
	panic("observation on an unknown subject")
}
