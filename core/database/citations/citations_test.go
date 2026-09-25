package citations

import (
	"bytes"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
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
	latestEntities := func(t *testing.T, c *database.Catalog) []string {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		rows, err := db.Query(`
			SELECT c.entity_type || '/' || c.action
			FROM audit_changes c
			JOIN audit_transactions tx ON tx.id = c.audit_transaction_id
			WHERE tx.revision = (SELECT MAX(revision) FROM audit_transactions)
			ORDER BY c.rowid`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var out []string
		for rows.Next() {
			var action string
			if err := rows.Scan(&action); err != nil {
				t.Fatal(err)
			}
			out = append(out, action)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		return out
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
			name: "transcription-only save keeps observation id and ref",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := CreateWithObservations(c, userID, CreateInput{
					ArtifactID:  s.artifact.ID,
					LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
					ValueText: "Boston", HasText: true,
				}})
				if err != nil {
					t.Fatal(err)
				}
				updated, err := UpdateWithObservations(c, userID, res.Citation.ID, CreateInput{
					ArtifactID:    s.artifact.ID,
					LocatorJSON:   validLocator,
					Transcription: "margin note",
				}, []observations.Input{{
					ID:        res.Observations[0].ID,
					SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
					ValueText: "Boston", HasText: true,
				}})
				if err != nil {
					t.Fatal(err)
				}
				if len(updated.Observations) != 1 {
					t.Fatalf("obs %+v", updated.Observations)
				}
				got := updated.Observations[0]
				if !bytes.Equal(got.ID, res.Observations[0].ID) || got.Ref != res.Observations[0].Ref {
					t.Fatalf("id/ref changed %q %q", got.Ref, res.Observations[0].Ref)
				}
				if got.ValueText != "Boston" {
					t.Fatalf("text %q", got.ValueText)
				}
				actions := latestEntities(t, c)
				if len(actions) != 1 || actions[0] != "citation/update" {
					t.Fatalf("audit %v", actions)
				}
			},
		},
		{
			name: "changed observation value keeps id",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := CreateWithObservations(c, userID, CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
					ValueText: "Boston", HasText: true,
				}})
				if err != nil {
					t.Fatal(err)
				}
				updated, err := UpdateWithObservations(c, userID, res.Citation.ID, CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					ID:        res.Observations[0].ID,
					SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
					ValueText: "Salem", HasText: true,
				}})
				if err != nil {
					t.Fatal(err)
				}
				got := updated.Observations[0]
				if !bytes.Equal(got.ID, res.Observations[0].ID) || got.Ref != res.Observations[0].Ref {
					t.Fatalf("id/ref changed")
				}
				if got.ValueText != "Salem" {
					t.Fatalf("text %q", got.ValueText)
				}
				actions := latestEntities(t, c)
				if len(actions) != 1 || actions[0] != "observation/update" {
					t.Fatalf("audit %v", actions)
				}
			},
		},
		{
			name: "new observation inserts and dropped id deletes",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := CreateWithObservations(c, userID, CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
					ValueText: "Boston", HasText: true,
				}})
				if err != nil {
					t.Fatal(err)
				}
				kept := res.Observations[0]
				added, err := UpdateWithObservations(c, userID, res.Citation.ID, CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{
					{
						ID: kept.ID, SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Boston", HasText: true,
					},
					{
						SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Salem", HasText: true,
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				if len(added.Observations) != 2 || !bytes.Equal(added.Observations[0].ID, kept.ID) {
					t.Fatalf("added %+v", added.Observations)
				}
				if bytes.Equal(added.Observations[1].ID, kept.ID) || added.Observations[1].Ref == kept.Ref {
					t.Fatalf("new row reused identity")
				}
				dropped, err := UpdateWithObservations(c, userID, res.Citation.ID, CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					ID: kept.ID, SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
					ValueText: "Boston", HasText: true,
				}})
				if err != nil {
					t.Fatal(err)
				}
				if len(dropped.Observations) != 1 || !bytes.Equal(dropped.Observations[0].ID, kept.ID) {
					t.Fatalf("dropped %+v", dropped.Observations)
				}
			},
		},
		{
			name: "foreign or duplicate observation id is invalid",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := CreateWithObservations(c, userID, CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
					ValueText: "Boston", HasText: true,
				}})
				if err != nil {
					t.Fatal(err)
				}
				foreign, err := uuid.NewV7()
				if err != nil {
					t.Fatal(err)
				}
				_, err = UpdateWithObservations(c, userID, res.Citation.ID, CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					ID: foreign[:], SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
					ValueText: "Salem", HasText: true,
				}})
				if !errors.Is(err, observations.ErrInvalid) {
					t.Fatalf("foreign %v", err)
				}
				_, err = UpdateWithObservations(c, userID, res.Citation.ID, CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{
					{
						ID: res.Observations[0].ID, SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Boston", HasText: true,
					},
					{
						ID: res.Observations[0].ID, SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Salem", HasText: true,
					},
				})
				if !errors.Is(err, observations.ErrInvalid) {
					t.Fatalf("duplicate %v", err)
				}
				listed, err := observations.ListByCitation(c, res.Citation.ID)
				if err != nil || len(listed) != 1 || listed[0].ValueText != "Boston" {
					t.Fatalf("rolled back %+v %v", listed, err)
				}
			},
		},
		{
			name: "date and name values update in place",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				eventType, err := subjecttypes.Lookup(c, "event", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				event, err := subjects.Create(c, userID, subjects.CreateInput{
					SourceID: s.artifact.SourceID, SubjectTypeID: eventType.ID, Label: "Birth",
				})
				if err != nil {
					t.Fatal(err)
				}
				dateProp, err := properties.Lookup(c, "date", properties.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				nameProp, err := properties.Lookup(c, "name", properties.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				year := 1842
				res, err := CreateWithObservations(c, userID, CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{
					{
						SubjectID: event.ID, PropertyID: dateProp.ID,
						Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year},
					},
					{
						SubjectID: s.person.ID, PropertyID: nameProp.ID,
						Name: &namevalues.Value{Form: "Ada"},
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				var dateObs, nameObs observations.Observation
				for _, obs := range res.Observations {
					switch {
					case len(obs.ValueDateID) == 16:
						dateObs = obs
					case len(obs.ValueNameID) == 16:
						nameObs = obs
					}
				}
				if len(dateObs.ID) != 16 || len(nameObs.ID) != 16 {
					t.Fatalf("created %+v", res.Observations)
				}
				year = 1843
				updated, err := UpdateWithObservations(c, userID, res.Citation.ID, CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{
					{
						ID: dateObs.ID, SubjectID: event.ID, PropertyID: dateProp.ID,
						Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year},
					},
					{
						ID: nameObs.ID, SubjectID: s.person.ID, PropertyID: nameProp.ID,
						Name: &namevalues.Value{Form: "Ada Lovelace"},
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				for _, obs := range updated.Observations {
					switch {
					case bytes.Equal(obs.ID, dateObs.ID):
						if !bytes.Equal(obs.ValueDateID, dateObs.ValueDateID) || obs.Ref != dateObs.Ref {
							t.Fatalf("date identity changed")
						}
					case bytes.Equal(obs.ID, nameObs.ID):
						if !bytes.Equal(obs.ValueNameID, nameObs.ValueNameID) || obs.Ref != nameObs.Ref {
							t.Fatalf("name identity changed")
						}
					default:
						t.Fatalf("unexpected observation %q", obs.Ref)
					}
				}
				gotDate, err := datevalues.Lookup(c, dateObs.ValueDateID)
				if err != nil || gotDate.StartYear == nil || *gotDate.StartYear != 1843 {
					t.Fatalf("date %+v %v", gotDate, err)
				}
				gotName, err := namevalues.Lookup(c, nameObs.ValueNameID)
				if err != nil || gotName.Form != "Ada Lovelace" {
					t.Fatalf("name %+v %v", gotName, err)
				}
			},
		},
		{
			name: "unused date value is deleted when the observation stops using it",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				eventType, err := subjecttypes.Lookup(c, "event", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				event, err := subjects.Create(c, userID, subjects.CreateInput{
					SourceID: s.artifact.SourceID, SubjectTypeID: eventType.ID, Label: "Birth",
				})
				if err != nil {
					t.Fatal(err)
				}
				dateProp, err := properties.Lookup(c, "date", properties.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				eventTypeProp, err := properties.Lookup(c, "event_type", properties.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				birth, err := propertyterms.Lookup(c, eventTypeProp.ID, "birth", propertyterms.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				year := 1842
				res, err := CreateWithObservations(c, userID, CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: event.ID, PropertyID: dateProp.ID,
					Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year},
				}})
				if err != nil {
					t.Fatal(err)
				}
				dateID := append([]byte(nil), res.Observations[0].ValueDateID...)
				if _, err := UpdateWithObservations(c, userID, res.Citation.ID, CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					ID: res.Observations[0].ID, SubjectID: event.ID, PropertyID: eventTypeProp.ID,
					ValueTermID: birth.ID,
				}}); err != nil {
					t.Fatal(err)
				}
				if _, err := datevalues.Lookup(c, dateID); err != sql.ErrNoRows {
					t.Fatalf("date row err=%v", err)
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
