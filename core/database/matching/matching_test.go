package matching_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/matching"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/ref"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

const locator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

type fixture struct {
	t        *testing.T
	c        *database.Catalog
	db       *sql.DB
	source   sources.Source
	artifact artifacts.Artifact
	types    map[string]subjecttypes.Type
	props    map[string]properties.Property
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func ip(n int) *int { return &n }

func newFixture(t *testing.T) *fixture {
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
	src, err := sources.Create(c, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	must(t, err)
	art, err := artifacts.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	must(t, err)
	db, err := c.DB()
	must(t, err)
	f := &fixture{t: t, c: c, db: db, source: src, artifact: art,
		types: map[string]subjecttypes.Type{}, props: map[string]properties.Property{}}
	for _, k := range []string{"person", "event", "place", "participation"} {
		st, err := subjecttypes.Lookup(c, k, subjecttypes.OriginProvenencia)
		must(t, err)
		f.types[k] = st
	}
	for _, k := range []string{"name", "sex_at_birth", "event_type", "date", "toponym"} {
		p, err := properties.Lookup(c, k, properties.OriginProvenencia)
		must(t, err)
		f.props[k] = p
	}
	return f
}

// value is one Observation to file on a new Subject.
type value func(f *fixture, s subjects.Subject) observations.Input

func named(form string) value {
	return func(f *fixture, s subjects.Subject) observations.Input {
		return observations.Input{SubjectID: s.ID, PropertyID: f.props["name"].ID, Name: namevaluestest.Western(form)}
	}
}

// namedParts files a name whose form is written surname-first and whose
// parts say which word is which.
func namedParts(form, given, surname string) value {
	return func(f *fixture, s subjects.Subject) observations.Input {
		return observations.Input{SubjectID: s.ID, PropertyID: f.props["name"].ID, Name: &namevalues.Value{Form: form, Parts: []namevalues.Part{
			{Idx: 0, Value: given, Type: namevalues.PartTypeGiven}, {Idx: 1, Value: surname, Type: namevalues.PartTypeSurname},
		}}}
	}
}

func termed(prop, key string) value {
	return func(f *fixture, s subjects.Subject) observations.Input {
		term, err := propertyterms.Lookup(f.c, f.props[prop].ID, key, propertyterms.OriginProvenencia)
		must(f.t, err)
		return observations.Input{SubjectID: s.ID, PropertyID: f.props[prop].ID, ValueTermID: term.ID}
	}
}

func dated(y int, m, d *int) value {
	return func(f *fixture, s subjects.Subject) observations.Input {
		return observations.Input{SubjectID: s.ID, PropertyID: f.props["date"].ID, Date: &datevalues.Value{
			Kind: datevalues.KindPoint, Calendar: "gregorian", StartYear: ip(y), StartMonth: m, StartDay: d,
		}}
	}
}

func toponym(s string) value {
	return func(f *fixture, sub subjects.Subject) observations.Input {
		return observations.Input{SubjectID: sub.ID, PropertyID: f.props["toponym"].ID, ValueText: s, HasText: true}
	}
}

func (f *fixture) subject(kind string, values ...value) subjects.Subject {
	f.t.Helper()
	s, err := subjects.Create(f.c, userID, subjects.CreateInput{SourceID: f.source.ID, SubjectTypeID: f.types[kind].ID}, nil)
	must(f.t, err)
	if len(values) > 0 {
		var in []observations.Input
		for _, v := range values {
			in = append(in, v(f, s))
		}
		_, err := citations.CreateWithObservations(f.c, userID, citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: locator}, in)
		must(f.t, err)
	}
	return s
}

// handle promotes a new Subject and returns the handle's ref and id.
func (f *fixture) handle(kind string, values ...value) promote.Result {
	f.t.Helper()
	res, err := promote.Save(f.c, userID, promote.Input{SubjectID: f.subject(kind, values...).ID})
	must(f.t, err)
	return res
}

func refs(ms []match.Match) string {
	var out []string
	for _, m := range ms {
		out = append(out, fmt.Sprintf("%s=%.1f", m.Ref, m.Score))
	}
	return fmt.Sprint(out)
}

func TestForSubjectPersons(t *testing.T) {
	f := newFixture(t)
	exact := f.handle("person", named("James Robins"), termed("sex_at_birth", "male"))
	variant := f.handle("person", named("James Robbins"))
	surname := f.handle("person", named("Mary Robins"))
	female := f.handle("person", named("James Robins"), termed("sex_at_birth", "female"))
	f.handle("person", named("Ada Lovelace"))
	f.handle("place", toponym("James Robins")) // another type never matches

	james := f.subject("person", named("James Robins"), termed("sex_at_birth", "male"))
	res, err := matching.ForSubject(f.db, james.ID, matching.Options{})
	must(t, err)
	if res.Kind != "person" || !res.Profiled {
		t.Fatalf("%+v", res)
	}
	// Fixture names are typed (given, surname), so a surname weighs more than a given name.
	// A sex mismatch subtracts 2, so the female James stays in the list at 8.
	want := fmt.Sprintf("[%s=11.0 %s=9.1 %s=8.0 %s=6.0]", exact.Entity.Ref, variant.Entity.Ref, female.Entity.Ref, surname.Entity.Ref)
	if refs(res.Matches) != want {
		t.Fatalf("got %s, want %s", refs(res.Matches), want)
	}
	if r := res.Matches[0].Reasons; len(r) != 2 || r[0].Property.Key != "name" || r[1].Property.Key != "sex_at_birth" {
		t.Fatalf("reasons %+v", r)
	}

	t.Run("limit", func(t *testing.T) {
		res, err := matching.ForSubject(f.db, james.ID, matching.Options{Limit: 1})
		must(t, err)
		if len(res.Matches) != 1 {
			t.Fatalf("%s", refs(res.Matches))
		}
	})

	t.Run("a profile override reweighs", func(t *testing.T) {
		p, _ := match.DefaultProfile("person")
		p = p.With(match.Feature{Property: match.Property{Key: "sex_at_birth", Origin: "provenencia"}, Comparer: match.TermComparer{}, Weight: 1})
		res, err := matching.ForSubject(f.db, james.ID, matching.Options{Profile: &p})
		must(t, err)
		if len(res.Matches) != 4 || res.Matches[1].Ref != female.Entity.Ref || res.Matches[1].Score != 10 {
			t.Fatalf("female James should rank once sex no longer contradicts: %s", refs(res.Matches))
		}
	})

	t.Run("the Subject's own handle is excluded", func(t *testing.T) {
		_, err := promote.Save(f.c, userID, promote.Input{SubjectID: james.ID, EntityID: exact.Entity.ID})
		must(t, err)
		res, err := matching.ForSubject(f.db, james.ID, matching.Options{})
		must(t, err)
		if want := fmt.Sprintf("[%s=9.1 %s=8.0 %s=6.0]", variant.Entity.Ref, female.Entity.Ref, surname.Entity.Ref); refs(res.Matches) != want {
			t.Fatalf("got %s, want %s", refs(res.Matches), want)
		}
	})

	t.Run("merged handles are excluded", func(t *testing.T) {
		_, err := f.db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, exact.Entity.ID, variant.Entity.ID)
		must(t, err)
		res, err := matching.ForSubject(f.db, f.subject("person", named("James Robbins")).ID, matching.Options{})
		must(t, err)
		for _, m := range res.Matches {
			if m.Ref == variant.Entity.Ref {
				t.Fatalf("merged handle matched: %s", refs(res.Matches))
			}
		}
	})

	t.Run("a nameless Subject matches nothing", func(t *testing.T) {
		res, err := matching.ForSubject(f.db, f.subject("person").ID, matching.Options{})
		must(t, err)
		if !res.Profiled || len(res.Matches) != 0 {
			t.Fatalf("%+v", res)
		}
	})
}

// Parts travel from Observations into the probe and through the cache into
// candidates, so structured comparison sees them on both sides.
func TestForSubjectComparesNameParts(t *testing.T) {
	f := newFixture(t)
	same := f.handle("person", namedParts("James Robins", "James", "Robins"))
	sibling := f.handle("person", namedParts("Mary Robins", "Mary", "Robins"))
	f.handle("person", namedParts("James Smith", "James", "Smith")) // 10 × 0.2: below MinScore

	res, err := matching.ForSubject(f.db, f.subject("person", namedParts("ROBINS, Jas.", "James", "Robins")).ID, matching.Options{})
	must(t, err)
	if want := fmt.Sprintf("[%s=10.0 %s=6.0]", same.Entity.Ref, sibling.Entity.Ref); refs(res.Matches) != want {
		t.Fatalf("got %s, want %s", refs(res.Matches), want)
	}
}

func TestForSubjectEventsAndPlaces(t *testing.T) {
	f := newFixture(t)
	sameDay := f.handle("event", termed("event_type", "birth"), dated(1817, ip(5), ip(14)))
	nextYear := f.handle("event", termed("event_type", "birth"), dated(1818, nil, nil))
	f.handle("event", termed("event_type", "birth"))                             // type alone: below MinScore
	f.handle("event", termed("event_type", "death"), dated(1817, ip(5), ip(14))) // type contradicts

	birth := f.subject("event", termed("event_type", "birth"), dated(1817, ip(5), ip(14)))
	res, err := matching.ForSubject(f.db, birth.ID, matching.Options{})
	must(t, err)
	if want := fmt.Sprintf("[%s=10.0 %s=7.2]", sameDay.Entity.Ref, nextYear.Entity.Ref); refs(res.Matches) != want {
		t.Fatalf("events: got %s, want %s", refs(res.Matches), want)
	}

	york := f.handle("place", toponym("York"))
	upper := f.handle("place", toponym("York, Upper Canada"))
	f.handle("place", toponym("Toronto"))
	res, err = matching.ForSubject(f.db, f.subject("place", toponym("york")).ID, matching.Options{})
	must(t, err)
	if want := fmt.Sprintf("[%s=10.0 %s=3.5]", york.Entity.Ref, upper.Entity.Ref); refs(res.Matches) != want {
		t.Fatalf("places: got %s, want %s", refs(res.Matches), want)
	}
}

func TestForEntityMergeHints(t *testing.T) {
	f := newFixture(t)
	a := f.handle("person", named("James Robins"))
	b := f.handle("person", named("james robins"))
	f.handle("person", named("Ada Lovelace"))
	res, err := matching.ForEntity(f.db, a.Entity.ID, matching.Options{})
	must(t, err)
	if want := fmt.Sprintf("[%s=10.0]", b.Entity.Ref); refs(res.Matches) != want {
		t.Fatalf("got %s, want %s (never itself)", refs(res.Matches), want)
	}
}

func TestUnprofiledAndUnknown(t *testing.T) {
	f := newFixture(t)
	res, err := matching.ForSubject(f.db, f.subject("participation").ID, matching.Options{})
	must(t, err)
	if res.Kind != "participation" || res.Profiled || len(res.Matches) != 0 {
		t.Fatalf("%+v", res)
	}
	for name, id := range map[string][]byte{"unknown": make([]byte, 16), "short": {1}} {
		if _, err := matching.ForSubject(f.db, id, matching.Options{}); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("%s subject: %v", name, err)
		}
		if _, err := matching.ForEntity(f.db, id, matching.Options{}); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("%s entity: %v", name, err)
		}
	}
}

func TestForSubjectQueryCountIsConstant(t *testing.T) {
	f := newFixture(t)
	f.handle("person", named("James Robins"))
	james := f.subject("person", named("James Robins"), named("Jim Robins"), termed("sex_at_birth", "male"))
	one, err := matching.ForSubjectQueryCount(f.db, james.ID)
	must(t, err)
	for i := 0; i < 50; i++ {
		f.handle("person", named(fmt.Sprintf("James Robins %d", i)), termed("sex_at_birth", "male"))
	}
	many, err := matching.ForSubjectQueryCount(f.db, james.ID)
	must(t, err)
	if many != one {
		t.Fatalf("queries: %d for 1 handle, %d for 51", one, many)
	}
}
