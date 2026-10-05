package conclusionheaders_test

import (
	"fmt"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
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

func (f *fixture) typeID(key string) []byte {
	st, err := subjecttypes.Lookup(f.c, key, subjecttypes.OriginProvenencia)
	must(f.t, err)
	return st.ID
}

// person promotes a new person Subject named by forms (one Observation each).
func (f *fixture) person(forms ...string) (subjects.Subject, []observations.Observation, []byte) {
	f.t.Helper()
	s, err := subjects.Create(f.c, userID, subjects.CreateInput{SourceID: f.source.ID, SubjectTypeID: f.typeID("person")}, nil)
	must(f.t, err)
	var obs []observations.Observation
	if len(forms) > 0 {
		var in []observations.Input
		for _, form := range forms {
			in = append(in, observations.Input{SubjectID: s.ID, PropertyID: f.name.ID, Name: namevaluestest.Western(form)})
		}
		res, err := citations.CreateWithObservations(f.c, userID, citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: locator}, in)
		must(f.t, err)
		obs = res.Observations
	}
	p, err := promote.Save(f.c, userID, promote.Input{SubjectID: s.ID})
	must(f.t, err)
	return s, obs, p.Entity.ID
}

func (f *fixture) list() []conclusionheaders.PersonHeader {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	h, err := conclusionheaders.ListPersons(db)
	must(f.t, err)
	return h
}

func TestListPersons(t *testing.T) {
	f := newFixture(t)
	if got := f.list(); len(got) != 0 {
		t.Fatalf("empty project listed %d", len(got))
	}

	f.person("Mary Smith")
	_, jamesObs, _ := f.person("James Robins", "Jim Robins", "james robins")
	labelled, err := canonicalentities.Create(f.c, userID, canonicalentities.CreateInput{SubjectTypeID: f.typeID("person"), Label: "Mother of James"})
	must(t, err)
	bare, err := canonicalentities.Create(f.c, userID, canonicalentities.CreateInput{SubjectTypeID: f.typeID("person")})
	must(t, err)
	_, err = canonicalentities.Create(f.c, userID, canonicalentities.CreateInput{SubjectTypeID: f.typeID("place"), Label: "York"})
	must(t, err)

	got := f.list()
	if len(got) != 4 {
		t.Fatalf("listed %d Persons, want 4 (no Place)", len(got))
	}
	// Named first by sort key (james < mary), then unnamed by ref. Jim
	// Robins is outvoted two to one: it keeps its cache row but isn't a
	// displayed value, so it isn't counted (S9-13).
	if got[0].Name == nil || got[0].Name.Form != "James Robins" || got[0].NameClusterCount != 1 {
		t.Fatalf("first %+v", got[0])
	}
	if got[1].Name == nil || got[1].Name.Form != "Mary Smith" || got[1].NameClusterCount != 1 {
		t.Fatalf("second %+v", got[1])
	}
	unnamed := map[string]conclusionheaders.PersonHeader{got[2].Entity.Ref: got[2], got[3].Entity.Ref: got[3]}
	if h := unnamed[labelled.Ref]; h.Name != nil || h.NameClusterCount != 0 || h.Entity.Label != "Mother of James" {
		t.Fatalf("labelled %+v", h)
	}
	if h := unnamed[bare.Ref]; h.Name != nil || h.Entity.Label != "" {
		t.Fatalf("bare %+v", h)
	}
	if got[2].Entity.Ref > got[3].Entity.Ref {
		t.Fatalf("unnamed not ordered by ref: %s, %s", got[2].Entity.Ref, got[3].Entity.Ref)
	}

	t.Run("a name edit reaches the header", func(t *testing.T) {
		// Jim → James merges the two clusters.
		_, err := observations.Update(f.c, userID, observations.Input{
			ID: jamesObs[1].ID, SubjectID: jamesObs[1].SubjectID, PropertyID: f.name.ID,
			Name: namevaluestest.Western("James Robins"),
		})
		must(t, err)
		if h := f.list()[0]; h.Name.Form != "James Robins" || h.NameClusterCount != 1 {
			t.Fatalf("after edit %+v", h)
		}
	})
}

// +N counts displayed names only: two names one to one are both displayed;
// an outvoted name keeps its row but isn't counted.
func TestPersonNameCountIsDisplayedOnly(t *testing.T) {
	f := newFixture(t)
	f.person("Ann Lee", "Anne Lee")
	f.person("James Robins", "james robins", "Jim Robins")
	counts := map[string]int{}
	for _, h := range f.list() {
		counts[h.Name.Form] = h.NameClusterCount
	}
	if counts["Ann Lee"] != 2 || counts["James Robins"] != 1 {
		t.Fatalf("counts %v", counts)
	}
}

func TestListPersonsQueryCountIsConstant(t *testing.T) {
	f := newFixture(t)
	f.person("Ada Lovelace")
	db, err := f.c.DB()
	must(t, err)
	one, err := conclusionheaders.ListPersonsQueryCount(db)
	must(t, err)
	for i := 0; i < 50; i++ {
		f.person(fmt.Sprintf("Person %d", i), fmt.Sprintf("Alias %d", i))
	}
	many, err := conclusionheaders.ListPersonsQueryCount(db)
	must(t, err)
	if one != 1 || many != one {
		t.Fatalf("queries: %d for 1 Person, %d for 51", one, many)
	}
}

func TestPersonsByIDs(t *testing.T) {
	f := newFixture(t)
	_, _, mary := f.person("Mary Smith")
	_, _, james := f.person("James Robins")
	f.person("Ada Lovelace")
	place, err := canonicalentities.Create(f.c, userID, canonicalentities.CreateInput{SubjectTypeID: f.typeID("place")})
	must(t, err)
	db, err := f.c.DB()
	must(t, err)

	got, err := conclusionheaders.PersonsByIDs(db, [][]byte{mary, james, mary, place.ID, make([]byte, 16)})
	must(t, err)
	if len(got) != 2 || got[0].Name.Form != "James Robins" || got[1].Name.Form != "Mary Smith" {
		t.Fatalf("want James, Mary in list order: %+v", got)
	}
	if got, err := conclusionheaders.PersonsByIDs(db, nil); err != nil || got != nil {
		t.Fatalf("no ids: %v %+v", err, got)
	}
}
