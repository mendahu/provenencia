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

func (f *fixture) suggest(s subjects.Subject, limit int) []string {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	got, err := promotetargets.Suggest(db, s.ID, limit)
	must(f.t, err)
	var refs []string
	for _, h := range got.Persons {
		refs = append(refs, h.Entity.Ref)
	}
	return refs
}

func TestSuggest(t *testing.T) {
	f := newFixture(t)
	// Exact matches: one with a better-supported "james robins" cluster.
	exactWeak := f.handle("person", "James Robins")
	exactStrong := f.handle("person", "james robins", "James Robins.")
	shared := f.handle("person", "Mary Robins")
	f.handle("person", "Ada Lovelace")
	f.handle("person", "J. Smith") // an initial alone never suggests
	f.handle("event")

	james := f.subject("person", "James Robins", "J. Doe")
	got := f.suggest(james, 0)
	want := []string{exactStrong.Entity.Ref, exactWeak.Entity.Ref, shared.Entity.Ref}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	t.Run("limit", func(t *testing.T) {
		if got := f.suggest(james, 1); len(got) != 1 || got[0] != exactStrong.Entity.Ref {
			t.Fatalf("%v", got)
		}
	})

	t.Run("the Subject's own handle is excluded", func(t *testing.T) {
		_, err := promote.Save(f.c, userID, promote.Input{SubjectID: james.ID, EntityID: exactWeak.Entity.ID})
		must(t, err)
		got := f.suggest(james, 0)
		want := []string{exactStrong.Entity.Ref, shared.Entity.Ref}
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("merged handles are excluded", func(t *testing.T) {
		db, err := f.c.DB()
		must(t, err)
		_, err = db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, exactWeak.Entity.ID, shared.Entity.ID)
		must(t, err)
		got := f.suggest(f.subject("person", "Mary Robins"), 0)
		if len(got) != 2 {
			t.Fatalf("want the two James handles: %v", got)
		}
		for _, r := range got {
			if r == shared.Entity.Ref {
				t.Fatalf("merged handle suggested: %v", got)
			}
		}
	})

	t.Run("a nameless Subject gets none", func(t *testing.T) {
		if got := f.suggest(f.subject("person"), 0); len(got) != 0 {
			t.Fatalf("%v", got)
		}
	})

	t.Run("events and places get none yet", func(t *testing.T) {
		if got := f.suggest(f.subject("event"), 0); len(got) != 0 {
			t.Fatalf("%v", got)
		}
	})
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

func TestSuggestQueryCountIsConstant(t *testing.T) {
	f := newFixture(t)
	f.handle("person", "James Robins")
	james := f.subject("person", "James Robins", "Jim Robins")
	db, err := f.c.DB()
	must(t, err)
	one, err := promotetargets.SuggestQueryCount(db, james.ID)
	must(t, err)
	for i := 0; i < 50; i++ {
		f.handle("person", fmt.Sprintf("James Robins %d", i), "Jim Robins")
	}
	many, err := promotetargets.SuggestQueryCount(db, james.ID)
	must(t, err)
	if many != one {
		t.Fatalf("queries: %d for 1 handle, %d for 51", one, many)
	}
}
