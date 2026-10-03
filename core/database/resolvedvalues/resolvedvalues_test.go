package resolvedvalues_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/resolvedvalues"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
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
	source   sources.Source
	artifact artifacts.Artifact
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
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	must(t, err)
	src, err := sources.Create(c, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	must(t, err)
	art, err := artifacts.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	must(t, err)

	f := &fixture{t: t, c: c, source: src, artifact: art,
		types: map[string]subjecttypes.Type{}, props: map[string]properties.Property{}, terms: map[string]propertyterms.Term{}}
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
	_, err = db.Exec(`INSERT INTO subject_type_fields (subject_type_id, property_id, sort_order) VALUES (?, ?, 99)`, f.types["person"].ID, ageID)
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

func (f *fixture) subject(kind string) subjects.Subject {
	f.t.Helper()
	s, err := subjects.Create(f.c, userID, subjects.CreateInput{SourceID: f.source.ID, SubjectTypeID: f.types[kind].ID}, nil)
	must(f.t, err)
	return s
}

func (f *fixture) cite(in ...observations.Input) []observations.Observation {
	f.t.Helper()
	res, err := citations.CreateWithObservations(f.c, userID, citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: locator}, in)
	must(f.t, err)
	return res.Observations
}

func (f *fixture) promote(s subjects.Subject) []byte {
	f.t.Helper()
	res, err := promote.Save(f.c, userID, promote.Input{SubjectID: s.ID})
	must(f.t, err)
	return res.Entity.ID
}

func nameIn(s subjects.Subject, p properties.Property, form string) observations.Input {
	return observations.Input{SubjectID: s.ID, PropertyID: p.ID, Name: &namevalues.Value{Form: form}}
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
	SortKey              *string
	Support              int
}

func snapshot(t *testing.T, q resolvedvalues.Querier) []row {
	t.Helper()
	rows, err := q.Query(`SELECT entity_id, property_id, rank, value_text, value_integer, value_term_id,
		value_date, value_name, sort_key, support FROM conclusion_resolved_values
		ORDER BY entity_id, property_id, rank`)
	must(t, err)
	defer rows.Close()
	var out []row
	for rows.Next() {
		var r row
		must(t, rows.Scan(&r.EntityID, &r.PropertyID, &r.Rank, &r.Text, &r.Integer, &r.TermID, &r.Date, &r.Name, &r.SortKey, &r.Support))
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
func (f *fixture) rebuiltSnapshot() []row {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	tx, err := db.Begin()
	must(f.t, err)
	defer func() { _ = tx.Rollback() }()
	must(f.t, resolvedvalues.Rebuild(tx))
	return snapshot(f.t, tx)
}

func TestResolvedValues(t *testing.T) {
	f := newFixture(t)
	person := f.subject("person")
	f.cite(nameIn(person, f.props["name"], "James Robins"), termIn(person, f.props["sex_at_birth"], f.terms["male"]), intIn(person, f.props["age"], 34))
	f.cite(nameIn(person, f.props["name"], "james  robins."), intIn(person, f.props["age"], 34))
	jim := f.cite(nameIn(person, f.props["name"], "Jim Robins"))[0]

	if got := f.rows(); len(got) != 0 {
		t.Fatalf("unpromoted subject cached %d rows", len(got))
	}
	per := f.promote(person)
	all := f.rows()

	names := rowsFor(all, per, f.props["name"].ID)
	if len(names) != 2 || names[0].Support != 2 || names[1].Support != 1 {
		t.Fatalf("name clusters %+v", names)
	}
	top, err := valuecodec.UnmarshalName(names[0].Name)
	must(t, err)
	if top.Form != "James Robins" || *names[0].SortKey != "james robins" || *names[1].SortKey != "jim robins" {
		t.Fatalf("rank 1 %q sort %q / %q", top.Form, *names[0].SortKey, *names[1].SortKey)
	}
	sex := rowsFor(all, per, f.props["sex_at_birth"].ID)
	if len(sex) != 1 || !bytes.Equal(sex[0].TermID, f.terms["male"].ID) || sex[0].SortKey != nil {
		t.Fatalf("sex %+v", sex)
	}
	age := rowsFor(all, per, f.props["age"].ID)
	if len(age) != 1 || *age[0].Integer != 34 || age[0].Support != 2 {
		t.Fatalf("age %+v", age)
	}

	t.Run("observation update", func(t *testing.T) {
		_, err := observations.Update(f.c, userID, observations.Input{
			ID: jim.ID, SubjectID: person.ID, PropertyID: f.props["name"].ID, Name: &namevalues.Value{Form: "James Robins"},
		})
		must(t, err)
		names := rowsFor(f.rows(), per, f.props["name"].ID)
		if len(names) != 1 || names[0].Support != 3 {
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
		if len(dates) != 2 || dates[0].SortKey != nil {
			t.Fatalf("dates %+v", dates)
		}
		d, err := valuecodec.UnmarshalDate(dates[0].Date)
		must(t, err)
		if *d.StartYear != 1985 || *d.StartMonth != 5 || d.StartDay != nil {
			t.Fatalf("rank 1 date %+v", d)
		}
		tops := rowsFor(all, plc, f.props["toponym"].ID)
		if len(tops) != 2 || *tops[0].Text != "Upper Canada" || *tops[0].SortKey != "upper canada" {
			t.Fatalf("toponyms %+v", tops)
		}
	})

	t.Run("subject delete leaves no rows", func(t *testing.T) {
		lone := f.subject("person")
		obs := f.cite(nameIn(lone, f.props["name"], "Mary Smith"))
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

	if got, want := f.rows(), f.rebuiltSnapshot(); !reflect.DeepEqual(got, want) {
		t.Fatalf("upkeep != rebuild\n got %d rows\nwant %d rows", len(got), len(want))
	}
}

func TestEnsureCatalogRebuildsStaleVersion(t *testing.T) {
	f := newFixture(t)
	p := f.subject("person")
	f.cite(nameIn(p, f.props["name"], "James Robins"))
	f.promote(p)
	want := f.rows()

	db, err := f.c.DB()
	must(t, err)
	_, err = db.Exec(`DELETE FROM conclusion_resolved_values`)
	must(t, err)
	_, err = db.Exec(`UPDATE conclusion_resolved_meta SET cache_version = 0`)
	must(t, err)
	need, err := resolvedvalues.NeedsRebuild(db)
	must(t, err)
	if !need {
		t.Fatal("stale version not detected")
	}
	must(t, resolvedvalues.EnsureCatalog(f.c))
	if v, err := resolvedvalues.StoredVersion(db); err != nil || v != resolvedvalues.CacheVersion {
		t.Fatalf("version %d %v", v, err)
	}
	if got := f.rows(); !reflect.DeepEqual(got, want) {
		t.Fatalf("rebuilt %d rows, want %d", len(got), len(want))
	}
	need, err = resolvedvalues.NeedsRebuild(db)
	must(t, err)
	if need {
		t.Fatal("rebuild did not store the version")
	}
}

func TestLoaderQueryCountIsConstant(t *testing.T) {
	f := newFixture(t)
	var handles [][]byte
	for i := 0; i < 60; i++ {
		p := f.subject("person")
		f.cite(nameIn(p, f.props["name"], fmt.Sprintf("Person %d", i)), intIn(p, f.props["age"], int64(i)))
		handles = append(handles, f.promote(p))
	}
	db, err := f.c.DB()
	must(t, err)
	one, err := resolvedvalues.LoadQueryCount(db, handles[:1])
	must(t, err)
	many, err := resolvedvalues.LoadQueryCount(db, handles)
	must(t, err)
	if one != many {
		t.Fatalf("loader queries: %d for 1 handle, %d for %d", one, many, len(handles))
	}
}

// TestRebuildEqualsUpkeep applies random write sequences through the real
// write APIs and checks the incrementally maintained table against a full
// rebuild. A new write path or trigger adds its operation here.
//
// Subject delete is exercised but cannot yet change rows: Impact refuses a
// Subject that still has Observations, and pins don't feed resolution until
// S9-14 (claim confidence) / S9-28 (subject-valued ends). Its hook is in place
// so those PRs only extend this sequence.
func TestRebuildEqualsUpkeep(t *testing.T) {
	for _, seed := range []int64{1, 2, 3, 4} {
		t.Run(fmt.Sprint("seed ", seed), func(t *testing.T) {
			f := newFixture(t)
			rng := rand.New(rand.NewSource(seed))
			var subs []subjects.Subject
			var obs []observations.Observation
			kinds := []string{"person", "person", "event", "place"}
			names := []string{"James Robins", "james robins", "Jim Robins", "J. Robins", "Mary Smith"}
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
				switch op := rng.Intn(10); {
				case op < 2 || len(subs) == 0:
					subs = append(subs, f.subject(kinds[rng.Intn(len(kinds))]))
				case op < 5:
					s := subs[rng.Intn(len(subs))]
					obs = append(obs, f.cite(value(s), value(s))...)
				case op < 6:
					_, err := promote.Save(f.c, userID, promote.Input{SubjectID: subs[rng.Intn(len(subs))].ID})
					tolerate(err)
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
				default:
					i := rng.Intn(len(subs))
					err := subjects.Delete(f.c, userID, subs[i].ID)
					tolerate(err)
					if err == nil {
						subs = append(subs[:i], subs[i+1:]...)
					}
				}
				if step%10 == 0 {
					if got, want := f.rows(), f.rebuiltSnapshot(); !reflect.DeepEqual(got, want) {
						t.Fatalf("step %d: upkeep has %d rows, rebuild %d", step, len(got), len(want))
					}
				}
			}
			if len(f.rows()) == 0 {
				t.Fatal("sequence never cached anything; the test is not exercising upkeep")
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
