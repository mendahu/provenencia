package sourcegraphprogress

import (
	"bytes"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
)

const validLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":12,"page_label":"10"}]}`

func TestSourceGraphProgress(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	type seed struct {
		c       *database.Catalog
		worked  sources.Source
		empty   sources.Source
		person  subjects.Subject
		place   subjects.Subject
		toponym properties.Property
	}

	mustSeed := func(t *testing.T) seed {
		t.Helper()
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = c.Close() })
		ur, err := ref.Mint(ref.PrefixUser)
		if err != nil {
			t.Fatal(err)
		}
		if err := users.Upsert(c, userID, "Tester", ur); err != nil {
			t.Fatal(err)
		}
		if err := subjectvocab.Install(c); err != nil {
			t.Fatal(err)
		}
		typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
			Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
		})
		if err != nil {
			t.Fatal(err)
		}
		worked, err := sources.Create(c, userID, sources.CreateInput{
			SourceTypeID: typeID, Title: "Worked",
		})
		if err != nil {
			t.Fatal(err)
		}
		empty, err := sources.Create(c, userID, sources.CreateInput{
			SourceTypeID: typeID, Title: "Empty",
		})
		if err != nil {
			t.Fatal(err)
		}
		personType, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		placeType, err := subjecttypes.Lookup(c, "place", subjecttypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		sourceType, err := subjecttypes.Lookup(c, "source", subjecttypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		person, err := subjects.Create(c, userID, subjects.CreateInput{
			SourceID: worked.ID, SubjectTypeID: personType.ID, Label: "Alice",
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		place, err := subjects.Create(c, userID, subjects.CreateInput{
			SourceID: worked.ID, SubjectTypeID: placeType.ID, Label: "Boston",
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := subjects.Create(c, userID, subjects.CreateInput{
			SourceID: worked.ID, SubjectTypeID: sourceType.ID, Label: "Reify",
		}, nil); err != nil {
			t.Fatal(err)
		}
		art, err := artifacts.Create(c, userID, artifacts.CreateInput{
			SourceID: worked.ID, Label: "Scan",
		})
		if err != nil {
			t.Fatal(err)
		}
		toponym, err := properties.Lookup(c, "toponym", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
			ArtifactID: art.ID, LocatorJSON: validLocator,
		}, []observations.Input{{
			SubjectID: place.ID, PropertyID: toponym.ID,
			ValueText: "Boston", HasText: true,
		}}); err != nil {
			t.Fatal(err)
		}
		return seed{c: c, worked: worked, empty: empty, person: person, place: place, toponym: toponym}
	}

	find := func(rows []Progress, sourceID []byte) (Progress, bool) {
		for _, p := range rows {
			if bytes.Equal(p.SourceID, sourceID) {
				return p, true
			}
		}
		return Progress{}, false
	}

	t.Run("migration creates observation subject index", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var ver int
		if err := db.QueryRow(`PRAGMA user_version`).Scan(&ver); err != nil {
			t.Fatal(err)
		}
		if ver < 29 {
			t.Fatalf("user_version=%d want >= 29", ver)
		}
		var name string
		if err := db.QueryRow(
			`SELECT name FROM sqlite_schema WHERE type='index' AND name=?`,
			"observations_subject_id_idx",
		).Scan(&name); err != nil || name != "observations_subject_id_idx" {
			t.Fatalf("index: %v %q", err, name)
		}
	})

	t.Run("list omits empty and excludes source type", func(t *testing.T) {
		s := mustSeed(t)
		rows, err := List(s.c)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := find(rows, s.empty.ID); ok {
			t.Fatalf("empty source should be omitted: %+v", rows)
		}
		got, ok := find(rows, s.worked.ID)
		if !ok {
			t.Fatalf("worked missing: %+v", rows)
		}
		if got.SubjectCount != 2 || got.ObservationCount != 1 {
			t.Fatalf("got %+v want subjects=2 obs=1", got)
		}
	})

	t.Run("get zeros for empty source", func(t *testing.T) {
		s := mustSeed(t)
		got, err := Get(s.c, s.empty.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.SubjectCount != 0 || got.ObservationCount != 0 {
			t.Fatalf("got %+v", got)
		}
		if !bytes.Equal(got.SourceID, s.empty.ID) {
			t.Fatalf("source id %+v", got.SourceID)
		}
	})

	t.Run("get matches list for worked source", func(t *testing.T) {
		s := mustSeed(t)
		got, err := Get(s.c, s.worked.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.SubjectCount != 2 || got.ObservationCount != 1 {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("rejects short source id", func(t *testing.T) {
		s := mustSeed(t)
		if _, err := Get(s.c, []byte{1}); !errors.Is(err, ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("rejects closed catalog", func(t *testing.T) {
		s := mustSeed(t)
		s.c.Close()
		if _, err := List(s.c); !errors.Is(err, database.ErrClosed) {
			t.Fatalf("list %v", err)
		}
		if _, err := Get(s.c, s.worked.ID); !errors.Is(err, database.ErrClosed) {
			t.Fatalf("get %v", err)
		}
	})
}
