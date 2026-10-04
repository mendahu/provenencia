package search

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
)

const handleLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

type handleFixture struct {
	t        *testing.T
	c        *database.Catalog
	user     []byte
	source   []byte
	artifact []byte
}

func newHandleFixture(t *testing.T) *handleFixture {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	user := seedUser(t, c)
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	src, err := sources.Create(c, user, sources.CreateInput{SourceTypeID: seedType(t, c, "Book", ""), Title: "Register"})
	if err != nil {
		t.Fatal(err)
	}
	art, err := artifacts.Create(c, user, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}
	return &handleFixture{t: t, c: c, user: user, source: src.ID, artifact: art.ID}
}

func (f *handleFixture) prop(key string) properties.Property {
	f.t.Helper()
	p, err := properties.Lookup(f.c, key, properties.OriginProvenencia)
	if err != nil {
		f.t.Fatal(err)
	}
	return p
}

// subject creates a Subject of kind with the given Observations.
func (f *handleFixture) subject(kind string, in func(s subjects.Subject) []observations.Input) subjects.Subject {
	f.t.Helper()
	st, err := subjecttypes.Lookup(f.c, kind, subjecttypes.OriginProvenencia)
	if err != nil {
		f.t.Fatal(err)
	}
	s, err := subjects.Create(f.c, f.user, subjects.CreateInput{SourceID: f.source, SubjectTypeID: st.ID}, nil)
	if err != nil {
		f.t.Fatal(err)
	}
	if in != nil {
		if _, err := citations.CreateWithObservations(f.c, f.user, citations.CreateInput{ArtifactID: f.artifact, LocatorJSON: handleLocator}, in(s)); err != nil {
			f.t.Fatal(err)
		}
	}
	return s
}

func (f *handleFixture) person(forms ...string) subjects.Subject {
	return f.subject("person", func(s subjects.Subject) []observations.Input {
		var in []observations.Input
		for _, form := range forms {
			in = append(in, observations.Input{SubjectID: s.ID, PropertyID: f.prop("name").ID, Name: &namevalues.Value{Form: form}})
		}
		return in
	})
}

func (f *handleFixture) promote(s subjects.Subject, onto []byte) promote.Result {
	f.t.Helper()
	res, err := promote.Save(f.c, f.user, promote.Input{SubjectID: s.ID, EntityID: onto})
	if err != nil {
		f.t.Fatal(err)
	}
	return res
}

func (f *handleFixture) search(text string, kinds ...string) []Hit {
	f.t.Helper()
	hits, err := DefaultEngine().Search(context.Background(), f.c, Query{Text: text, Kinds: kinds})
	if err != nil {
		f.t.Fatal(err)
	}
	return hits
}

func refsOf(hits []Hit) string {
	var out []string
	for _, h := range hits {
		out = append(out, h.Kind+":"+h.Ref)
	}
	sort.Strings(out)
	return fmt.Sprint(out)
}

func TestHandleSearch(t *testing.T) {
	f := newHandleFixture(t)
	james := f.promote(f.person("James Robins", "Jim Robins"), nil)
	f.promote(f.person("Mary Smith"), nil)

	t.Run("the omnibar default leaves handles out", func(t *testing.T) {
		if hits := f.search("Robins"); len(hits) != 0 {
			t.Fatalf("got %s", refsOf(hits))
		}
	})

	t.Run("by rank-1 name, with location and member count", func(t *testing.T) {
		hits := f.search("James Robins", KindPerson)
		if len(hits) != 1 || hits[0].Ref != james.Entity.Ref || hits[0].Title != "James Robins" {
			t.Fatalf("got %+v", hits)
		}
		loc := hits[0].Location
		if loc.Section != SectionPersons || loc.EntityID != uuidString(james.Entity.ID) || loc.Ref != james.Entity.Ref {
			t.Fatalf("location %+v", loc)
		}
		if hits[0].MemberCount != 1 {
			t.Fatalf("members %d", hits[0].MemberCount)
		}
	})

	t.Run("by another name cluster", func(t *testing.T) {
		if hits := f.search("Jim", KindPerson); refsOf(hits) != fmt.Sprint([]string{"person:" + james.Entity.Ref}) {
			t.Fatalf("got %s", refsOf(hits))
		}
	})

	t.Run("by ref", func(t *testing.T) {
		hits := f.search(james.Entity.Ref, KindPerson)
		if len(hits) == 0 || hits[0].Ref != james.Entity.Ref || hits[0].MatchReason != "ref" {
			t.Fatalf("got %+v", hits)
		}
	})

	t.Run("a join raises the member count and adds its names", func(t *testing.T) {
		f.promote(f.person("J. Robbins"), james.Entity.ID)
		hits := f.search("Robbins", KindPerson)
		if len(hits) != 1 || hits[0].Ref != james.Entity.Ref || hits[0].MemberCount != 2 {
			t.Fatalf("got %+v", hits)
		}
	})

	t.Run("unknown kinds return nothing", func(t *testing.T) {
		if hits := f.search("Robins", "nope"); len(hits) != 0 {
			t.Fatalf("got %s", refsOf(hits))
		}
	})
}

func TestHandleSearchFollowsEdits(t *testing.T) {
	f := newHandleFixture(t)
	s := f.subject("person", func(s subjects.Subject) []observations.Input {
		return []observations.Input{{SubjectID: s.ID, PropertyID: f.prop("name").ID, Name: &namevalues.Value{Form: "Ada Byron"}}}
	})
	ada := f.promote(s, nil)
	obs, err := observations.ListBySubject(f.c, s.ID)
	if err != nil || len(obs) != 1 {
		t.Fatalf("%v %d", err, len(obs))
	}
	if _, err := observations.Update(f.c, f.user, observations.Input{
		ID: obs[0].ID, SubjectID: s.ID, PropertyID: f.prop("name").ID, Name: &namevalues.Value{Form: "Ada Lovelace"},
	}); err != nil {
		t.Fatal(err)
	}
	if hits := f.search("Lovelace", KindPerson); len(hits) != 1 || hits[0].Ref != ada.Entity.Ref {
		t.Fatalf("new name: %s", refsOf(hits))
	}
	if hits := f.search("Byron", KindPerson); len(hits) != 0 {
		t.Fatalf("old name still found: %s", refsOf(hits))
	}
}

func TestHandleSearchEventsAndPlaces(t *testing.T) {
	f := newHandleFixture(t)
	birth, err := propertyterms.Lookup(f.c, f.prop("event_type").ID, "birth", propertyterms.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	y, m := 1817, 5
	event := f.promote(f.subject("event", func(s subjects.Subject) []observations.Input {
		return []observations.Input{
			{SubjectID: s.ID, PropertyID: f.prop("event_type").ID, ValueTermID: birth.ID},
			{SubjectID: s.ID, PropertyID: f.prop("date").ID, Date: &datevalues.Value{Kind: datevalues.KindPoint, Calendar: "gregorian", StartYear: &y, StartMonth: &m}},
		}
	}), nil)
	place := f.promote(f.subject("place", func(s subjects.Subject) []observations.Input {
		return []observations.Input{{SubjectID: s.ID, PropertyID: f.prop("toponym").ID, ValueText: "York, Upper Canada", HasText: true}}
	}), nil)
	f.promote(f.person("York Smith"), nil)

	if hits := f.search("Birth", KindEvent); len(hits) != 1 || hits[0].Ref != event.Entity.Ref || hits[0].Title != "Birth" {
		t.Fatalf("event by type: %+v", hits)
	}
	if hits := f.search("1817", KindEvent); len(hits) != 1 || hits[0].Location.Section != SectionEvents {
		t.Fatalf("event by year: %+v", hits)
	}
	if hits := f.search("York", KindPlace); refsOf(hits) != fmt.Sprint([]string{"place:" + place.Entity.Ref}) {
		t.Fatalf("place only: %s", refsOf(hits))
	}
	if hits := f.search("York", KindPlace, KindPerson); len(hits) != 2 {
		t.Fatalf("both kinds: %s", refsOf(hits))
	}
}

func TestHandleWithoutValuesIsFoundByRef(t *testing.T) {
	f := newHandleFixture(t)
	bare := f.promote(f.person(), nil)
	hits := f.search(bare.Entity.Ref, KindPerson)
	if len(hits) != 1 || hits[0].Title != bare.Entity.Ref {
		t.Fatalf("got %+v", hits)
	}
}

func TestRebuildIndexesHandles(t *testing.T) {
	f := newHandleFixture(t)
	james := f.promote(f.person("James Robins"), nil)
	db, err := f.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := searchindex.ClearAll(db); err != nil {
		t.Fatal(err)
	}
	if err := searchindex.SetProjectionVersion(db, 0); err != nil {
		t.Fatal(err)
	}
	if err := searchindex.EnsureCatalog(f.c); err != nil {
		t.Fatal(err)
	}
	if hits := f.search("Robins", KindPerson); len(hits) != 1 || hits[0].Ref != james.Entity.Ref {
		t.Fatalf("got %s", refsOf(hits))
	}
}

func TestMergedHandleLeavesTheIndex(t *testing.T) {
	f := newHandleFixture(t)
	a := f.promote(f.person("James Robins"), nil)
	b := f.promote(f.person("James Robins"), nil)
	db, err := f.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, a.Entity.ID, b.Entity.ID); err != nil {
		t.Fatal(err)
	}
	if err := searchindex.ReprojectHandles(db, [][]byte{b.Entity.ID}); err != nil {
		t.Fatal(err)
	}
	if hits := f.search("Robins", KindPerson); refsOf(hits) != fmt.Sprint([]string{"person:" + a.Entity.Ref}) {
		t.Fatalf("got %s", refsOf(hits))
	}
}
