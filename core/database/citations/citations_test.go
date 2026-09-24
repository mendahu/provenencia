package citations

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/locator"
	"github.com/mendahu/provenencia/core/ref"
)

func TestCitations(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	validLocator := `{"version":1,"selectors":[{"type":"page","artifact_page":12,"page_label":"10"}]}`

	type seed struct {
		artifact    artifacts.Artifact
		person      subjects.Subject
		place       subjects.Subject
		toponymProp properties.Property
		sexProp     properties.Property
		femaleTerm  propertyterms.Term
	}

	mustSeed := func(t *testing.T) (*database.Catalog, seed) {
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
		src, err := sources.Create(c, userID, sources.CreateInput{
			SourceTypeID: typeID,
			Title:        "Family Bible",
		})
		if err != nil {
			t.Fatal(err)
		}
		art, err := artifacts.Create(c, userID, artifacts.CreateInput{
			SourceID: src.ID,
			Label:    "Scan",
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
		person, err := subjects.Create(c, userID, subjects.CreateInput{
			SourceID: src.ID, SubjectTypeID: personType.ID, Label: "Alice",
		})
		if err != nil {
			t.Fatal(err)
		}
		place, err := subjects.Create(c, userID, subjects.CreateInput{
			SourceID: src.ID, SubjectTypeID: placeType.ID, Label: "Boston",
		})
		if err != nil {
			t.Fatal(err)
		}
		toponymProp, err := properties.Lookup(c, "toponym", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		sexProp, err := properties.Lookup(c, "sex_at_birth", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		femaleTerm, err := propertyterms.Lookup(c, sexProp.ID, "female", propertyterms.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		return c, seed{
			artifact: art, person: person, place: place,
			toponymProp: toponymProp, sexProp: sexProp, femaleTerm: femaleTerm,
		}
	}
	latestAction := func(t *testing.T, c *database.Catalog) string {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var actionType string
		if err := db.QueryRow(
			`SELECT action_type FROM audit_transactions ORDER BY revision DESC LIMIT 1`,
		).Scan(&actionType); err != nil {
			t.Fatal(err)
		}
		return actionType
	}

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "migration creates citation tables",
			run: func(t *testing.T) {
				c, err := database.Create(t.TempDir(), "t.provenencia")
				if err != nil {
					t.Fatal(err)
				}
				defer c.Close()
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				for _, table := range []string{"citations", "citation_notes", "observations", "observation_notes"} {
					var name string
					if err := db.QueryRow(
						`SELECT name FROM sqlite_schema WHERE type='table' AND name=?`, table,
					).Scan(&name); err != nil || name != table {
						t.Fatalf("table %q: %v", table, err)
					}
				}
				var ver int
				if err := db.QueryRow(`PRAGMA user_version`).Scan(&ver); err != nil {
					t.Fatal(err)
				}
				if ver < 26 {
					t.Fatalf("user_version=%d want >= 26", ver)
				}
			},
		},
		{
			name: "create with observations text and term round-trip",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := CreateWithObservations(c, userID, CreateInput{
					ArtifactID:    s.artifact.ID,
					LocatorJSON:   validLocator,
					Transcription: "Alice, female, of Boston",
					Description:   "entry on page 10",
				}, []observations.Input{
					{
						SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Boston", HasText: true,
					},
					{
						SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
						ValueTermID: s.femaleTerm.ID,
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(res.Citation.Ref, "CIT-") {
					t.Fatalf("citation ref %q", res.Citation.Ref)
				}
				if len(res.Observations) != 2 {
					t.Fatalf("obs count %d", len(res.Observations))
				}
				for _, o := range res.Observations {
					if !strings.HasPrefix(o.Ref, "OBS-") {
						t.Fatalf("obs ref %q", o.Ref)
					}
				}
				if latestAction(t, c) != "create_citation_with_observations" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				got, err := Get(c, res.Citation.ID)
				if err != nil {
					t.Fatal(err)
				}
				if got.Transcription != "Alice, female, of Boston" || got.LocatorJSON != validLocator {
					t.Fatalf("got %+v", got)
				}
				bySubject, err := observations.ListBySubject(c, s.person.ID)
				if err != nil || len(bySubject) != 1 {
					t.Fatalf("person obs %v len=%d", err, len(bySubject))
				}
				if string(bySubject[0].ValueTermID) != string(s.femaleTerm.ID) {
					t.Fatalf("term id mismatch")
				}
				placeObs, err := observations.ListBySubject(c, s.place.ID)
				if err != nil || len(placeObs) != 1 {
					t.Fatalf("place obs %v len=%d", err, len(placeObs))
				}
				if !placeObs[0].HasText || placeObs[0].ValueText != "Boston" {
					t.Fatalf("toponym %+v", placeObs[0])
				}
			},
		},
		{
			name: "invalid locator rejected",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				_, err := CreateWithObservations(c, userID, CreateInput{
					ArtifactID:  s.artifact.ID,
					LocatorJSON: `{"version":1,"selectors":[{"type":"page","artifact_page":0}]}`,
				}, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				}})
				if !errors.Is(err, locator.ErrInvalid) {
					t.Fatalf("got %v want locator.ErrInvalid", err)
				}
			},
		},
		{
			name: "list by artifact",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				if _, err := CreateWithObservations(c, userID, CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				}}); err != nil {
					t.Fatal(err)
				}
				list, err := ListByArtifact(c, s.artifact.ID)
				if err != nil || len(list) != 1 {
					t.Fatalf("%v len=%d", err, len(list))
				}
				if list[0].Ref[:4] != "CIT-" {
					t.Fatalf("ref %q", list[0].Ref)
				}
				if list[0].ObservationCount != 1 {
					t.Fatalf("obs count %d", list[0].ObservationCount)
				}
			},
		},
		{
			name: "create with zero observations",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := CreateWithObservations(c, userID, CreateInput{
					ArtifactID:    s.artifact.ID,
					LocatorJSON:   validLocator,
					Transcription: "transcribe first",
				}, nil)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(res.Citation.Ref, "CIT-") {
					t.Fatalf("citation ref %q", res.Citation.Ref)
				}
				if len(res.Observations) != 0 {
					t.Fatalf("obs count %d", len(res.Observations))
				}
				list, err := ListByArtifact(c, s.artifact.ID)
				if err != nil || len(list) != 1 {
					t.Fatalf("%v len=%d", err, len(list))
				}
				if list[0].ObservationCount != 0 {
					t.Fatalf("listed count %d", list[0].ObservationCount)
				}
			},
		},
		{
			name: "update populated citation down to zero observations",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := CreateWithObservations(c, userID, CreateInput{
					ArtifactID:  s.artifact.ID,
					LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				}})
				if err != nil {
					t.Fatal(err)
				}
				updated, err := UpdateWithObservations(c, userID, res.Citation.ID, CreateInput{
					ArtifactID:    s.artifact.ID,
					LocatorJSON:   validLocator,
					Transcription: "cleared",
				}, nil)
				if err != nil {
					t.Fatal(err)
				}
				if updated.Citation.Transcription != "cleared" {
					t.Fatalf("transcription %q", updated.Citation.Transcription)
				}
				if len(updated.Observations) != 0 {
					t.Fatalf("obs count %d", len(updated.Observations))
				}
				listed, err := observations.ListByCitation(c, res.Citation.ID)
				if err != nil || len(listed) != 0 {
					t.Fatalf("persisted %v len=%d", err, len(listed))
				}
			},
		},
		{
			name: "count by source",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				art2, err := artifacts.Create(c, userID, artifacts.CreateInput{
					SourceID: s.artifact.SourceID,
					Label:    "Scan 2",
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := CreateWithObservations(c, userID, CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				}}); err != nil {
					t.Fatal(err)
				}
				if _, err := CreateWithObservations(c, userID, CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				}}); err != nil {
					t.Fatal(err)
				}
				if _, err := CreateWithObservations(c, userID, CreateInput{
					ArtifactID: art2.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				}}); err != nil {
					t.Fatal(err)
				}
				counts, err := CountBySource(c, s.artifact.SourceID)
				if err != nil {
					t.Fatal(err)
				}
				id1 := uuid.Must(uuid.FromBytes(s.artifact.ID)).String()
				id2 := uuid.Must(uuid.FromBytes(art2.ID)).String()
				if counts[id1] != 2 || counts[id2] != 1 {
					t.Fatalf("counts=%v want %s=2 %s=1", counts, id1, id2)
				}
			},
		},
		{
			name: "count by source rejects short id",
			run: func(t *testing.T) {
				c, _ := mustSeed(t)
				if _, err := CountBySource(c, []byte{1}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v want ErrInvalid", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
