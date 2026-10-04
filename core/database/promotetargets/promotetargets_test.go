package promotetargets_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promotetargets"
	"github.com/mendahu/provenencia/core/database/properties"
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
	source   sources.Source
	artifact artifacts.Artifact
	name     properties.Property
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
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	must(t, err)
	src, err := sources.Create(c, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	must(t, err)
	art, err := artifacts.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	must(t, err)
	name, err := properties.Lookup(c, "name", properties.OriginProvenencia)
	must(t, err)
	return &fixture{t: t, c: c, source: src, artifact: art, name: name}
}

// subject creates an unpromoted Subject of kind named by forms.
func (f *fixture) subject(kind string, forms ...string) subjects.Subject {
	f.t.Helper()
	st, err := subjecttypes.Lookup(f.c, kind, subjecttypes.OriginProvenencia)
	must(f.t, err)
	s, err := subjects.Create(f.c, userID, subjects.CreateInput{SourceID: f.source.ID, SubjectTypeID: st.ID}, nil)
	must(f.t, err)
	if len(forms) > 0 {
		var in []observations.Input
		for _, form := range forms {
			in = append(in, observations.Input{SubjectID: s.ID, PropertyID: f.name.ID, Name: &namevalues.Value{Form: form}})
		}
		_, err := citations.CreateWithObservations(f.c, userID, citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: locator}, in)
		must(f.t, err)
	}
	return s
}

// handle promotes a new Subject of kind named by forms and returns its ref.
func (f *fixture) handle(kind string, forms ...string) promote.Result {
	f.t.Helper()
	res, err := promote.Save(f.c, userID, promote.Input{SubjectID: f.subject(kind, forms...).ID})
	must(f.t, err)
	return res
}

func (f *fixture) suggest(s subjects.Subject, limit int) []promotetargets.Suggestion {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	got, err := promotetargets.Suggest(db, s.ID, limit)
	must(f.t, err)
	return got
}

func refsOf(got []promotetargets.Suggestion) string {
	var refs []string
	for _, s := range got {
		refs = append(refs, s.Entity.Ref)
	}
	return fmt.Sprint(refs)
}

// Ranking itself is tested in core/match and core/database/matching; these
// tests cover what Promote's picker gets.
func TestSuggestPersons(t *testing.T) {
	f := newFixture(t)
	exact := f.handle("person", "James Robins")
	shared := f.handle("person", "Mary Robins")
	f.handle("person", "Ada Lovelace")

	got := f.suggest(f.subject("person", "James Robins"), 0)
	if want := fmt.Sprint([]string{exact.Entity.Ref, shared.Entity.Ref}); refsOf(got) != want {
		t.Fatalf("got %s, want %s", refsOf(got), want)
	}
	top := got[0]
	if top.Person == nil || top.Person.Name == nil || top.Person.Name.Form != "James Robins" {
		t.Fatalf("person header %+v", top.Person)
	}
	if string(top.Entity.ID) != string(exact.Entity.ID) || top.Score != 10 {
		t.Fatalf("%+v", top)
	}
	if len(top.Reasons) != 1 || top.Reasons[0].Property.Key != "name" || top.Reasons[0].Similarity != 1 {
		t.Fatalf("reasons %+v", top.Reasons)
	}
	if got := f.suggest(f.subject("person", "James Robins"), 1); len(got) != 1 {
		t.Fatalf("limit: %s", refsOf(got))
	}
}

func TestSuggestDefaultLimit(t *testing.T) {
	f := newFixture(t)
	for i := 0; i < match.DefaultSuggestionLimit+2; i++ {
		f.handle("person", fmt.Sprintf("James Robins %d", i))
	}
	if got := f.suggest(f.subject("person", "James Robins"), 0); len(got) != match.DefaultSuggestionLimit {
		t.Fatalf("got %d", len(got))
	}
}

func TestSuggestPlacesCarryTheHandleOnly(t *testing.T) {
	f := newFixture(t)
	st, err := subjecttypes.Lookup(f.c, "place", subjecttypes.OriginProvenencia)
	must(t, err)
	toponym, err := properties.Lookup(f.c, "toponym", properties.OriginProvenencia)
	must(t, err)
	place := func(name string) subjects.Subject {
		s, err := subjects.Create(f.c, userID, subjects.CreateInput{SourceID: f.source.ID, SubjectTypeID: st.ID}, nil)
		must(t, err)
		_, err = citations.CreateWithObservations(f.c, userID, citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: locator},
			[]observations.Input{{SubjectID: s.ID, PropertyID: toponym.ID, ValueText: name, HasText: true}})
		must(t, err)
		return s
	}
	york, err := promote.Save(f.c, userID, promote.Input{SubjectID: place("York").ID})
	must(t, err)

	got := f.suggest(place("york"), 0)
	if len(got) != 1 || got[0].Entity.Ref != york.Entity.Ref || got[0].Person != nil {
		t.Fatalf("%+v", got)
	}
}

func TestSuggestRefusals(t *testing.T) {
	f := newFixture(t)
	db, err := f.c.DB()
	must(t, err)
	if _, err := promotetargets.Suggest(db, f.subject("participation").ID, 0); !errors.Is(err, promote.ErrUnsupportedType) {
		t.Fatalf("bridge kind: %v", err)
	}
	if _, err := promotetargets.Suggest(db, make([]byte, 16), 0); !errors.Is(err, promote.ErrInvalid) {
		t.Fatalf("unknown subject: %v", err)
	}
	if _, err := promotetargets.Suggest(db, []byte{1}, 0); !errors.Is(err, promote.ErrInvalid) {
		t.Fatalf("short id: %v", err)
	}
}
