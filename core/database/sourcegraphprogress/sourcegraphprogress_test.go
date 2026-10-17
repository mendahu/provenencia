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
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
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
			Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
		})
		if err != nil {
			t.Fatal(err)
		}
		worked, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{
				SourceTypeID: typeID, Title: "Worked",
			})
		})
		if err != nil {
			t.Fatal(err)
		}
		empty, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{
				SourceTypeID: typeID, Title: "Empty",
			})
		})
		if err != nil {
			t.Fatal(err)
		}
		personType, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		placeType, err := subjecttypes.Lookup(c, "place", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		sourceType, err := subjecttypes.Lookup(c, "source", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		person, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
			return subjects.Create(tx, userID, subjects.CreateInput{
				SourceID: worked.ID, SubjectTypeID: personType.ID, Label: "Alice",
			}, nil)
		})
		if err != nil {
			t.Fatal(err)
		}
		place, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
			return subjects.Create(tx, userID, subjects.CreateInput{
				SourceID: worked.ID, SubjectTypeID: placeType.ID, Label: "Boston",
			}, nil)
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
			return subjects.Create(tx, userID, subjects.CreateInput{
				SourceID: worked.ID, SubjectTypeID: sourceType.ID, Label: "Reify",
			}, nil)
		}); err != nil {
			t.Fatal(err)
		}
		art, err := runArtifactCreate(c, userID, artifacts.CreateInput{
			SourceID: worked.ID, Label: "Scan",
		})
		if err != nil {
			t.Fatal(err)
		}
		toponym, err := properties.Lookup(c, "toponym", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, citations.CreateInput{
				ArtifactID: art.ID, LocatorJSON: validLocator,
			}, []observations.Input{{
				SubjectID: place.ID, PropertyID: toponym.ID,
				ValueText: "Boston", HasText: true,
			}})
		}); err != nil {
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

func TestPrimaryAndFiledCounts(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
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
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	newSource := func(title string) sources.Source {
		t.Helper()
		src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: title})
		})
		if err != nil {
			t.Fatal(err)
		}
		return src
	}
	newSubject := func(src sources.Source, typeKey, label string) subjects.Subject {
		t.Helper()
		st, err := subjecttypes.Lookup(c, typeKey, vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		sub, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
			return subjects.Create(tx, userID, subjects.CreateInput{
				SourceID: src.ID, SubjectTypeID: st.ID, Label: label,
			}, nil)
		})
		if err != nil {
			t.Fatal(err)
		}
		return sub
	}
	entityN := 0
	claim := func(sub subjects.Subject, status string) {
		t.Helper()
		entityN++
		entityID := []byte{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, byte(entityN)}
		if _, err := db.Exec(`INSERT INTO canonical_entities (id, subject_type_id, ref) VALUES (?, ?, ?)`,
			entityID, sub.SubjectTypeID, "CNE-TEST"+string(rune('0'+entityN))); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO identity_claims (id, subject_id, entity_id, subject_type_id, status) VALUES (?, ?, ?, ?, ?)`,
			[]byte{8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, byte(entityN)}, sub.ID, entityID, sub.SubjectTypeID, status); err != nil {
			t.Fatal(err)
		}
	}

	noSubjects := newSource("None")
	unfiled := newSource("Unfiled")
	newSubject(unfiled, "person", "Alice")
	filed := newSource("Filed")
	claim(newSubject(filed, "person", "Bob"), "accepted")
	provisional := newSource("Provisional")
	claim(newSubject(provisional, "person", "Cy"), "provisional")
	bridgeOnly := newSource("Bridge")
	newSubject(bridgeOnly, "participation", "Role")
	mixed := newSource("Mixed")
	claim(newSubject(mixed, "person", "Di"), "accepted")
	newSubject(mixed, "place", "Boston")
	reifyOnly := newSource("Reify")
	newSubject(reifyOnly, "source", "Reify")

	want := map[string]Progress{
		"Unfiled":     {SubjectCount: 1, PrimaryCount: 1, FiledCount: 0},
		"Filed":       {SubjectCount: 1, PrimaryCount: 1, FiledCount: 1},
		"Provisional": {SubjectCount: 1, PrimaryCount: 1, FiledCount: 0},
		"Bridge":      {SubjectCount: 1, PrimaryCount: 0, FiledCount: 0},
		"Mixed":       {SubjectCount: 2, PrimaryCount: 2, FiledCount: 1},
	}
	srcs := map[string]sources.Source{
		"Unfiled": unfiled, "Filed": filed, "Provisional": provisional,
		"Bridge": bridgeOnly, "Mixed": mixed,
	}
	rows, err := List(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != len(want) {
		t.Fatalf("rows %d want %d: %+v", len(rows), len(want), rows)
	}
	for name, w := range want {
		src := srcs[name]
		var got *Progress
		for i := range rows {
			if bytes.Equal(rows[i].SourceID, src.ID) {
				got = &rows[i]
			}
		}
		if got == nil {
			t.Fatalf("%s missing from list", name)
		}
		if got.SubjectCount != w.SubjectCount || got.PrimaryCount != w.PrimaryCount || got.FiledCount != w.FiledCount {
			t.Fatalf("%s list %+v want %+v", name, *got, w)
		}
		one, err := Get(c, src.ID)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(one.SourceID, got.SourceID) || one.SubjectCount != got.SubjectCount ||
			one.ObservationCount != got.ObservationCount || one.PrimaryCount != got.PrimaryCount ||
			one.FiledCount != got.FiledCount {
			t.Fatalf("%s get %+v != list %+v", name, one, *got)
		}
	}
	for _, src := range []sources.Source{noSubjects, reifyOnly} {
		for _, p := range rows {
			if bytes.Equal(p.SourceID, src.ID) {
				t.Fatalf("source %s should be absent from list", src.Title)
			}
		}
		got, err := Get(c, src.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.SubjectCount != 0 || got.PrimaryCount != 0 || got.FiledCount != 0 {
			t.Fatalf("get %s %+v", src.Title, got)
		}
	}
}

func runArtifactCreate(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error) {
	a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, in)
		})
	return a, err
}
