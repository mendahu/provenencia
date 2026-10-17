package matching_test

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/matching"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
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
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	must(t, err)
	art, err := runArtifactCreate(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
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
		return observations.Input{SubjectID: s.ID, PropertyID: f.props["name"].ID, Name: namevaluetest.Western(form)}
	}
}

// namedParts files a name whose form is written surname-first and whose
// parts say which word is which.
func namedParts(form, given, surname string) value {
	return func(f *fixture, s subjects.Subject) observations.Input {
		return observations.Input{SubjectID: s.ID, PropertyID: f.props["name"].ID, Name: &namevalue.Value{Form: form, Parts: []namevalue.Part{
			{Idx: 0, Value: given, Type: namevalue.PartTypeGiven}, {Idx: 1, Value: surname, Type: namevalue.PartTypeSurname},
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
		return observations.Input{SubjectID: s.ID, PropertyID: f.props["date"].ID, Date: &datevalue.Value{
			Kind: datevalue.KindPoint, Calendar: "gregorian", StartYear: ip(y), StartMonth: m, StartDay: d,
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
	s, err := writes.Call(f.c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{SourceID: f.source.ID, SubjectTypeID: f.types[kind].ID}, nil)
	})
	must(f.t, err)
	if len(values) > 0 {
		var in []observations.Input
		for _, v := range values {
			in = append(in, v(f, s))
		}
		_, err := writes.Call(f.c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: locator}, in)
		})
		must(f.t, err)
	}
	return s
}

// handle promotes a new Subject and returns the handle's ref and id.
func (f *fixture) handle(kind string, values ...value) promote.Result {
	f.t.Helper()
	p6a1 := promote.Input{SubjectID: f.subject(kind, values...).ID}
	res, err := writes.Call(f.c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, p6a1)
	})
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
	// A sex mismatch subtracts 2, so the female James stays in the list at 7.5.
	// The name points are discounted by how common the words are among the
	// filed people (Robins on three of five, James on three): 0.2 ln(3)/ln(100)
	// is about 5%, so the exact name earns 9.5, not 10.
	want := fmt.Sprintf("[%s=10.5 %s=8.8 %s=7.5 %s=5.7]", exact.Entity.Ref, variant.Entity.Ref, female.Entity.Ref, surname.Entity.Ref)
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
		if len(res.Matches) != 4 || res.Matches[1].Ref != female.Entity.Ref || math.Abs(res.Matches[1].Score-9.52) > 0.01 {
			t.Fatalf("female James should rank once sex no longer contradicts: %s", refs(res.Matches))
		}
	})

	t.Run("the Subject's own handle is excluded", func(t *testing.T) {
		_, err := writes.Call(f.c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
			return promote.Save(tx, userID, promote.Input{SubjectID: james.ID, EntityID: exact.Entity.ID})
		})
		must(t, err)
		res, err := matching.ForSubject(f.db, james.ID, matching.Options{})
		must(t, err)
		if want := fmt.Sprintf("[%s=8.8 %s=7.5 %s=5.7]", variant.Entity.Ref, female.Entity.Ref, surname.Entity.Ref); refs(res.Matches) != want {
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
	if want := fmt.Sprintf("[%s=9.7 %s=5.8]", same.Entity.Ref, sibling.Entity.Ref); refs(res.Matches) != want {
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
	if want := fmt.Sprintf("[%s=10.0 %s=7.0]", sameDay.Entity.Ref, nextYear.Entity.Ref); refs(res.Matches) != want {
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
	// Robins is on two of three filed people, so the exact name earns 9.7.
	if want := fmt.Sprintf("[%s=9.7]", b.Entity.Ref); refs(res.Matches) != want {
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

func TestNoStoredDefaultStampsWestern(t *testing.T) {
	f := newFixture(t)
	probe := f.subject("person", namedParts("Vincent van Gogh", "Vincent", "Gogh"))
	handle := f.handle("person", namedParts("Vincent van Gogh", "Vincent", "Gogh"))
	profile, _ := match.DefaultProfile("person")
	prop := match.Property{Key: "name", Origin: "provenencia"}
	vals, err := matching.SubjectsValues(f.db, [][]byte{probe.ID}, profile)
	must(t, err)
	cands, err := matching.CandidatesOfType(f.db, "person", "provenencia", profile)
	must(t, err)
	var candidate match.Value
	for _, c := range cands {
		if c.Ref == handle.Entity.Ref {
			candidate = c.Values[prop][0]
		}
	}
	subject := vals[string(probe.ID)][prop][0]
	if subject.NamePattern != match.WesternPattern || candidate.NamePattern != match.WesternPattern {
		t.Fatalf("patterns %q %q", subject.NamePattern, candidate.NamePattern)
	}
	got, ok := nameformats.Loaded{Key: match.WesternPattern, Builtin: true}.Comparer().Compare(subject, candidate)
	want, wantOK := match.DefaultNames.Compare(subject, candidate)
	if ok != wantOK || got != want {
		t.Fatalf("score %v %v, built-in %v %v", got, ok, want, wantOK)
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

// sixty files 60 people named John <Surname>: 30 Smith, 1 Quirk, 29 others.
func (f *fixture) sixty() {
	f.t.Helper()
	for i := 0; i < 30; i++ {
		f.handle("person", named("John Smith"))
	}
	f.handle("person", named("John Quirk"))
	for i := 1; i <= 29; i++ {
		f.handle("person", named(fmt.Sprintf("John Surname%02d", i)))
	}
}

func topScore(t *testing.T, f *fixture, form string) float64 {
	t.Helper()
	res, err := matching.ForSubject(f.db, f.subject("person", named(form)).ID, matching.Options{})
	must(t, err)
	if len(res.Matches) == 0 {
		t.Fatalf("%s: no matches", form)
	}
	return res.Matches[0].Score
}

func TestForSubjectRanksRareSurnameFirst(t *testing.T) {
	f := newFixture(t)
	f.sixty()
	smith, quirk := topScore(t, f, "John Smith"), topScore(t, f, "John Quirk")
	if quirk <= smith {
		t.Fatalf("quirk %.3f, smith %.3f", quirk, smith)
	}
	if smith >= 10 {
		t.Fatalf("a common name kept every point: %.3f", smith)
	}
}

func TestRankSkipsCountsWithoutANameFeature(t *testing.T) {
	f := newFixture(t)
	f.handle("place", toponym("York"))
	s := f.subject("place", toponym("York"))
	one, err := matching.ForSubjectQueryCount(f.db, s.ID)
	must(t, err)
	for i := 0; i < 20; i++ {
		f.handle("place", toponym("York"))
	}
	many, err := matching.ForSubjectQueryCount(f.db, s.ID)
	must(t, err)
	if one != many {
		t.Fatalf("queries: %d then %d", one, many)
	}
	res, err := matching.ForSubject(f.db, s.ID, matching.Options{})
	must(t, err)
	if len(res.Matches) == 0 || res.Matches[0].Score != 10 {
		t.Fatalf("%v", refs(res.Matches))
	}
}

func runArtifactCreate(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error) {
	a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, in)
		})
	return a, err
}

// negated files the value as a negative Observation ("not male").
func negated(v value) value {
	return func(f *fixture, s subjects.Subject) observations.Input {
		in := v(f, s)
		in.Polarity = observations.PolarityNegative
		return in
	}
}

// addCachedTerm gives a handle another cached term row, as the reconciler
// would have written it.
func (f *fixture) addCachedTerm(entityID []byte, prop string, rank int, key string, support int, reason string) {
	f.t.Helper()
	term, err := propertyterms.Lookup(f.c, f.props[prop].ID, key, propertyterms.OriginProvenencia)
	must(f.t, err)
	_, err = f.db.Exec(`INSERT INTO auto_reconciler_values (entity_id, property_id, rank, value_term_id, support, against, reason)
		VALUES (?, ?, ?, ?, ?, 0, ?)`, entityID, f.props[prop].ID, rank, term.ID, support, reason)
	must(f.t, err)
}

func (f *fixture) setCachedTerm(entityID []byte, prop string, support int, reason string) {
	f.t.Helper()
	_, err := f.db.Exec(`UPDATE auto_reconciler_values SET support = ?, reason = ? WHERE entity_id = ? AND property_id = ? AND rank = 1`,
		support, reason, entityID, f.props[prop].ID)
	must(f.t, err)
}

func sexReason(t *testing.T, res matching.Result, ref string) match.Reason {
	t.Helper()
	for _, m := range res.Matches {
		if m.Ref != ref {
			continue
		}
		for _, r := range m.Reasons {
			if r.Property.Key == "sex_at_birth" {
				return r
			}
		}
	}
	t.Fatalf("no sex reason for %s in %+v", ref, res.Matches)
	return match.Reason{}
}

func TestForSubjectJudgesAgainstTheMajority(t *testing.T) {
	f := newFixture(t)
	h := f.handle("person", named("Mary Robins"), termed("sex_at_birth", "male"))
	f.setCachedTerm(h.Entity.ID, "sex_at_birth", 3, "kept")
	f.addCachedTerm(h.Entity.ID, "sex_at_birth", 2, "female", 1, "kept")

	female := f.subject("person", named("Mary Robins"), termed("sex_at_birth", "female"))
	res, err := matching.ForSubject(f.db, female.ID, matching.Options{})
	must(t, err)
	r := sexReason(t, res, h.Entity.Ref)
	if r.Contribution != -2 || r.Probe.Label != "Female" || r.Candidate.Label != "Male" {
		t.Fatalf("one dissenting reading: %+v", r)
	}

	male := f.subject("person", named("Mary Robins"), termed("sex_at_birth", "male"))
	res, err = matching.ForSubject(f.db, male.ID, matching.Options{})
	must(t, err)
	if r := sexReason(t, res, h.Entity.Ref); r.Contribution != 1 {
		t.Fatalf("the majority matches: %+v", r)
	}

	t.Run("a denied value", func(t *testing.T) {
		d := f.handle("person", named("Mary Robins"), termed("sex_at_birth", "male"))
		f.setCachedTerm(d.Entity.ID, "sex_at_birth", 3, "denied")
		f.addCachedTerm(d.Entity.ID, "sex_at_birth", 2, "female", 1, "kept")
		res, err := matching.ForSubject(f.db, male.ID, matching.Options{})
		must(t, err)
		r := sexReason(t, res, d.Entity.Ref)
		if r.Contribution != -2 || r.Probe.Label != "Male" || r.Candidate.Label != "Male" {
			t.Fatalf("%+v", r)
		}
	})
}

func TestSubjectsValuesNegatives(t *testing.T) {
	f := newFixture(t)
	profile, _ := match.DefaultProfile("person")
	both := f.subject("person", named("Mary Robins"), termed("sex_at_birth", "female"), negated(termed("sex_at_birth", "male")))
	only := f.subject("person", negated(termed("sex_at_birth", "male")))

	got, err := matching.SubjectsValues(f.db, [][]byte{both.ID, only.ID}, profile)
	must(t, err)
	sex := match.Property{Key: "sex_at_birth", Origin: "provenencia"}
	vs := got[string(both.ID)][sex]
	if len(vs) != 2 || vs[0].Negative || vs[0].Term != "female" || !vs[1].Negative || vs[1].Term != "male" || len(vs[1].ObservationID) != 16 {
		t.Fatalf("%+v", vs)
	}
	if string(vs[0].ObservationID) >= string(vs[1].ObservationID) {
		t.Fatalf("not in Observation id order: %x then %x", vs[0].ObservationID, vs[1].ObservationID)
	}
	if o := got[string(only.ID)][sex]; len(o) != 1 || !o[0].Negative {
		t.Fatalf("%+v", o)
	}

	f.handle("person", named("Mary Robins"), termed("sex_at_birth", "male"))
	res, err := matching.ForSubject(f.db, only.ID, matching.Options{})
	must(t, err)
	if len(res.Matches) != 0 {
		t.Fatalf("a Subject with only a negative ranks nothing: %+v", res.Matches)
	}
}

func TestValuesCarryTermLabel(t *testing.T) {
	f := newFixture(t)
	profile, _ := match.DefaultProfile("person")
	sex := match.Property{Key: "sex_at_birth", Origin: "provenencia"}
	check := func(what string, vs []match.Value) {
		t.Helper()
		if len(vs) != 1 || vs[0].Term != "male" || vs[0].TermLabel != "Male" {
			t.Fatalf("%s: %+v", what, vs)
		}
	}

	s := f.subject("person", termed("sex_at_birth", "male"))
	got, err := matching.SubjectsValues(f.db, [][]byte{s.ID}, profile)
	must(t, err)
	check("SubjectsValues", got[string(s.ID)][sex])

	h := f.handle("person", named("Mary Robins"), termed("sex_at_birth", "male"))
	cands, err := matching.CandidatesOfType(f.db, "person", "provenencia", profile)
	must(t, err)
	var found bool
	for _, c := range cands {
		if c.Ref == h.Entity.Ref {
			found = true
			check("CandidatesOfType", c.Values[sex])
		}
	}
	if !found {
		t.Fatalf("handle %s missing from candidates", h.Entity.Ref)
	}

	ev, err := matching.EntitiesValues(f.db, [][]byte{h.Entity.ID}, profile)
	must(t, err)
	check("EntitiesValues", ev[string(h.Entity.ID)][sex])
}
