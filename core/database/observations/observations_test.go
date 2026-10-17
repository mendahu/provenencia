package observations_test

import (
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/observations/observationstest"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ingest"
	"github.com/mendahu/provenencia/core/locator"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

func TestObservations(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	validLocator := `{"version":1,"selectors":[{"type":"page","artifact_page":3}]}`

	type seed struct {
		source      sources.Source
		artifact    artifacts.Artifact
		person      subjects.Subject
		place       subjects.Subject
		event       subjects.Subject
		toponymProp properties.Property
		sexProp     properties.Property
		nameProp    properties.Property
		dateProp    properties.Property
		likeness    properties.Property
		femaleTerm  propertyterms.Term
		maleTerm    propertyterms.Term
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
		src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{
				SourceTypeID: typeID,
				Title:        "Deed book",
			})
		})
		if err != nil {
			t.Fatal(err)
		}
		art, err := runArtifactCreate(c, userID, artifacts.CreateInput{
			SourceID: src.ID,
			Label:    "Page scan",
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
		eventType, err := subjecttypes.Lookup(c, "event", subjecttypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		person, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
			return subjects.Create(tx, userID, subjects.CreateInput{
				SourceID: src.ID, SubjectTypeID: personType.ID, Label: "Bob",
			}, &subjects.Placement{GridX: 0, GridY: 0})
		})
		if err != nil {
			t.Fatal(err)
		}
		place, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
			return subjects.Create(tx, userID, subjects.CreateInput{
				SourceID: src.ID, SubjectTypeID: placeType.ID,
			}, &subjects.Placement{GridX: 4, GridY: 0})
		})
		if err != nil {
			t.Fatal(err)
		}
		event, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
			return subjects.Create(tx, userID, subjects.CreateInput{
				SourceID: src.ID, SubjectTypeID: eventType.ID, Label: "Birth",
			}, &subjects.Placement{GridX: 2, GridY: 4})
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
		nameProp, err := properties.Lookup(c, "name", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		dateProp, err := properties.Lookup(c, "date", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		likeness, err := properties.Lookup(c, "likeness", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		femaleTerm, err := propertyterms.Lookup(c, sexProp.ID, "female", propertyterms.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		maleTerm, err := propertyterms.Lookup(c, sexProp.ID, "male", propertyterms.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		return c, seed{
			source: src, artifact: art, person: person, place: place, event: event,
			toponymProp: toponymProp, sexProp: sexProp, nameProp: nameProp, dateProp: dateProp, likeness: likeness,
			femaleTerm: femaleTerm, maleTerm: maleTerm,
		}
	}
	mustCitation := func(t *testing.T, c *database.Catalog, s seed, obs ...observations.Input) citations.Citation {
		t.Helper()
		res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, citations.CreateInput{
				ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
			}, obs)
		})
		if err != nil {
			t.Fatal(err)
		}
		return res.Citation
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
			name: "add to citation append",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				cit := mustCitation(t, c, s, observations.Input{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				})
				added, err := writes.Call(c, writes.Op{Action: "add_observations", UserID: userID}, func(tx *database.Tx) ([]observations.Observation, []rowchange.Change, error) {
					return observations.AddToCitation(tx, userID, cit.ID, []observations.Input{{
						SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
						ValueTermID: s.maleTerm.ID,
						Polarity:    observations.PolarityNegative,
					}})
				})
				if err != nil || len(added) != 1 {
					t.Fatalf("%v len=%d", err, len(added))
				}
				if added[0].Polarity != observations.PolarityNegative {
					t.Fatalf("polarity %q", added[0].Polarity)
				}
				if latestAction(t, c) != "add_observations" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				list, err := observations.ListBySubject(c, s.person.ID)
				if err != nil || len(list) != 2 {
					t.Fatalf("%v len=%d", err, len(list))
				}
			},
		},
		{
			name: "unbound property refused",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				unbound, err := writes.Call(c, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
					return properties.Create(tx, userID, "Nickname", properties.ValueTypeText, "", "")
				})
				if err != nil {
					t.Fatal(err)
				}
				cit := mustCitation(t, c, s, observations.Input{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				})
				_, err = writes.Call(c, writes.Op{Action: "add_observations", UserID: userID}, func(tx *database.Tx) ([]observations.Observation, []rowchange.Change, error) {
					return observations.AddToCitation(tx, userID, cit.ID, []observations.Input{{
						SubjectID: s.person.ID, PropertyID: unbound.ID,
						ValueText: "Al", HasText: true,
					}})
				})
				if !errors.Is(err, observations.ErrInvalid) {
					t.Fatalf("got %v want ErrInvalid", err)
				}
			},
		},
		{
			name: "list by source",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				cit := mustCitation(t, c, s,
					observations.Input{
						SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
						ValueTermID: s.femaleTerm.ID,
					},
					observations.Input{
						SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Cambridge", HasText: true,
					},
				)
				list, err := observations.ListByCitation(c, cit.ID)
				if err != nil || len(list) != 2 {
					t.Fatalf("%v len=%d", err, len(list))
				}
				var sawText, sawTerm bool
				for _, row := range list {
					if row.PropertyKey == "toponym" && row.HasText && row.ValueText == "Cambridge" {
						sawText = true
					}
					if row.PropertyKey == "sex_at_birth" && len(row.ValueTermID) == 16 {
						if row.ValueTermLabel == "" || row.ValueText != row.ValueTermLabel {
							t.Fatalf("term display: label=%q text=%q", row.ValueTermLabel, row.ValueText)
						}
						sawTerm = true
					}
				}
				if !sawText || !sawTerm {
					t.Fatalf("text=%v term=%v rows=%+v", sawText, sawTerm, list)
				}
			},
		},
		{
			name: "list denormalizes name and date for card display",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				y, m := 1882, 4
				cit := mustCitation(t, c, s,
					observations.Input{
						SubjectID: s.person.ID, PropertyID: s.nameProp.ID,
						Name: &namevalues.Value{
							Form: "Ada Lovelace",
							Parts: []namevalues.Part{
								{Idx: 0, Value: "Ada", Type: "given"},
								{Idx: 1, Value: "Lovelace", Type: "surname"},
							},
						},
					},
					observations.Input{
						SubjectID: s.event.ID, PropertyID: s.dateProp.ID,
						Date: &datevalues.Value{
							Kind: datevalues.KindPoint, Qualifier: datevalues.QualifierABT,
							Calendar: "gregorian", StartYear: &y, StartMonth: &m,
						},
					},
				)
				list, err := observations.ListByCitation(c, cit.ID)
				if err != nil || len(list) != 2 {
					t.Fatalf("%v len=%d", err, len(list))
				}
				var sawName, sawDate bool
				for _, row := range list {
					if row.PropertyKey == "name" {
						if row.ValueNameForm != "Ada Lovelace" || row.ValueText != "Ada Lovelace" {
							t.Fatalf("name display form=%q text=%q", row.ValueNameForm, row.ValueText)
						}
						if row.Name == nil || len(row.Name.Parts) != 2 || row.Name.Parts[0].Type != "given" {
							t.Fatalf("name parts=%v", row.Name)
						}
						sawName = true
					}
					if row.PropertyKey == "date" {
						if row.Date == nil || row.ValueText != "About 1882-04" {
							t.Fatalf("date display text=%q date=%v", row.ValueText, row.Date)
						}
						sawDate = true
					}
				}
				if !sawName || !sawDate {
					t.Fatalf("name=%v date=%v rows=%+v", sawName, sawDate, list)
				}
			},
		},
		{
			name: "blank text observation refused",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				_, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
					return citations.CreateWithObservations(tx, userID, citations.CreateInput{
						ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
					}, []observations.Input{{
						SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "   ", HasText: true,
					}})
				})
				if !errors.Is(err, observations.ErrInvalid) {
					t.Fatalf("got %v want ErrInvalid", err)
				}
				_, err = writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
					return citations.CreateWithObservations(tx, userID, citations.CreateInput{
						ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
					}, []observations.Input{{
						SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						HasText: false,
					}})
				})
				if !errors.Is(err, observations.ErrInvalid) {
					t.Fatalf("got %v want ErrInvalid", err)
				}
			},
		},
		{
			name: "update date in place records date_value change",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				year := 1842
				res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
					return citations.CreateWithObservations(tx, userID, citations.CreateInput{
						ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
					}, []observations.Input{{
						SubjectID: s.event.ID, PropertyID: s.dateProp.ID,
						Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year},
					}})
				})
				if err != nil {
					t.Fatal(err)
				}
				year2 := 1843
				got, err := writes.Call(c, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
					return observations.Update(tx, userID, observations.Input{
						ID: res.Observations[0].ID, SubjectID: s.event.ID, PropertyID: s.dateProp.ID,
						Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year2},
					})
				})
				if err != nil {
					t.Fatal(err)
				}
				if got.Date == nil || got.Date.StartYear == nil || *got.Date.StartYear != 1843 {
					t.Fatalf("date %+v", got.Date)
				}
				if latestAction(t, c) != "update_observation" {
					t.Fatalf("action %q", latestAction(t, c))
				}
			},
		},
		{
			name: "two observations cannot share a date value",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				year := 1842
				res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
					return citations.CreateWithObservations(tx, userID, citations.CreateInput{
						ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
					}, []observations.Input{{
						SubjectID: s.event.ID, PropertyID: s.dateProp.ID,
						Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year},
					}})
				})
				if err != nil {
					t.Fatal(err)
				}
				dateID := res.Observations[0].ValueDateID
				if len(dateID) != 16 {
					t.Fatalf("date id %x", dateID)
				}
				_, err = writes.Call(c, writes.Op{Action: "add_observations", UserID: userID}, func(tx *database.Tx) ([]observations.Observation, []rowchange.Change, error) {
					return observations.AddToCitation(tx, userID, res.Citation.ID, []observations.Input{{
						SubjectID: s.event.ID, PropertyID: s.dateProp.ID,
						ValueDateID: dateID,
					}})
				})
				if err == nil {
					t.Fatal("expected unique value_date_id to refuse a second observation")
				}
			},
		},
		{
			name: "two observations cannot share a name value",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
					return citations.CreateWithObservations(tx, userID, citations.CreateInput{
						ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
					}, []observations.Input{{
						SubjectID: s.person.ID, PropertyID: s.nameProp.ID,
						Name: &namevalues.Value{Form: "Ada Lovelace"},
					}})
				})
				if err != nil {
					t.Fatal(err)
				}
				nameID := res.Observations[0].ValueNameID
				if len(nameID) != 16 {
					t.Fatalf("name id %x", nameID)
				}
				_, err = writes.Call(c, writes.Op{Action: "add_observations", UserID: userID}, func(tx *database.Tx) ([]observations.Observation, []rowchange.Change, error) {
					return observations.AddToCitation(tx, userID, res.Citation.ID, []observations.Input{{
						SubjectID: s.person.ID, PropertyID: s.nameProp.ID,
						ValueNameID: nameID,
					}})
				})
				if err == nil {
					t.Fatal("expected unique value_name_id to refuse a second observation")
				}
			},
		},
		{
			name: "update no-op records no revision",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
					return citations.CreateWithObservations(tx, userID, citations.CreateInput{
						ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
					}, []observations.Input{{
						SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Boston", HasText: true,
					}})
				})
				if err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var before int
				if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions`).Scan(&before); err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
					return observations.Update(tx, userID, observations.Input{
						ID: res.Observations[0].ID, SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Boston", HasText: true,
					})
				}); err != nil {
					t.Fatal(err)
				}
				var after int
				if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions`).Scan(&after); err != nil {
					t.Fatal(err)
				}
				if after != before {
					t.Fatalf("revisions before=%d after=%d", before, after)
				}
			},
		},
		{
			name: "delete records full row and leaves citation",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
					return citations.CreateWithObservations(tx, userID, citations.CreateInput{
						ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
						Notes: []string{"keep me"},
					}, []observations.Input{{
						SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Boston", HasText: true,
						Notes: []string{"obs note"},
					}})
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "delete_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := observations.Delete(tx, userID, res.Observations[0].ID)
					return struct{}{}, changes, err
				}); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "delete_observation" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				list, err := observations.ListByCitation(c, res.Citation.ID)
				if err != nil || len(list) != 0 {
					t.Fatalf("obs leftover %v len=%d", err, len(list))
				}
				if _, err := citations.Get(c, res.Citation.ID); err != nil {
					t.Fatalf("citation should remain: %v", err)
				}
			},
		},
		{
			name: "delete drops owned date and name values",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				year := 1844
				res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
					return citations.CreateWithObservations(tx, userID, citations.CreateInput{
						ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
					}, []observations.Input{
						{
							SubjectID: s.event.ID, PropertyID: s.dateProp.ID,
							Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year},
						},
						{
							SubjectID: s.person.ID, PropertyID: s.nameProp.ID,
							Name: &namevalues.Value{Form: "Ada"},
						},
					})
				})
				if err != nil {
					t.Fatal(err)
				}
				dateID := res.Observations[0].ValueDateID
				nameID := res.Observations[1].ValueNameID
				if len(dateID) != 16 || len(nameID) != 16 {
					t.Fatalf("owned ids date=%x name=%x", dateID, nameID)
				}
				if _, err := writes.Call(c, writes.Op{Action: "delete_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := observations.Delete(tx, userID, res.Observations[0].ID)
					return struct{}{}, changes, err
				}); err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "delete_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := observations.Delete(tx, userID, res.Observations[1].ID)
					return struct{}{}, changes, err
				}); err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var dates, names int
				if err := db.QueryRow(`SELECT COUNT(*) FROM date_values WHERE id = ?`, dateID).Scan(&dates); err != nil {
					t.Fatal(err)
				}
				if err := db.QueryRow(`SELECT COUNT(*) FROM name_values WHERE id = ?`, nameID).Scan(&names); err != nil {
					t.Fatal(err)
				}
				if dates != 0 || names != 0 {
					t.Fatalf("owned leftover dates=%d names=%d", dates, names)
				}
			},
		},
		{
			name: "date replaced by another Property releases its value",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				year := 1844
				res, _, err := observationstest.SaveCitation(c, userID, s.artifact, observations.Input{
					SubjectID: s.event.ID, PropertyID: s.dateProp.ID,
					Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year},
				})
				if err != nil {
					t.Fatal(err)
				}
				oldDateID := res.Observations[0].ValueDateID
				if len(oldDateID) != 16 {
					t.Fatalf("date id %x", oldDateID)
				}
				eventNameProp, err := properties.Lookup(c, "event_name", properties.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				changes, err := observationstest.UpdateObservation(c, userID, observations.Input{
					ID: res.Observations[0].ID, SubjectID: s.event.ID, PropertyID: eventNameProp.ID,
					ValueText: "Birth", HasText: true,
				})
				if err != nil {
					t.Fatal(err)
				}
				if n := rowCount(t, c, `SELECT COUNT(*) FROM date_values WHERE id = ?`, oldDateID); n != 0 {
					t.Fatalf("replaced date value left behind: %d rows", n)
				}
				if got, want := changeKinds(changes, "date_value"), []string{"date_value/delete"}; !sameStrings(got, want) {
					t.Fatalf("changes %v want %v", got, want)
				}
			},
		},
		{
			name: "delete schedules no purge with nothing released",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
					return citations.CreateWithObservations(tx, userID, citations.CreateInput{
						ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
					}, []observations.Input{{
						SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Boston", HasText: true,
					}})
				})
				if err != nil {
					t.Fatal(err)
				}
				sentinel := filepath.Join(c.Dir(), "objects", "ab", "cd", "sentinel")
				if err := os.MkdirAll(filepath.Dir(sentinel), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "delete_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := observations.Delete(tx, userID, res.Observations[0].ID)
					return struct{}{}, changes, err
				}); err != nil {
					t.Fatal(err)
				}
				if _, err := os.Stat(sentinel); err != nil {
					t.Fatalf("sentinel object removed: %v", err)
				}
			},
		},
		{
			name: "edge rows refuse update and delete",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				personProp, err := properties.Lookup(c, "person", properties.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				eventProp, err := properties.Lookup(c, "event", properties.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				roleProp, err := properties.Lookup(c, "role", properties.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				roleTerm, err := propertyterms.Lookup(c, roleProp.ID, "witness", propertyterms.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				bridge, err := writes.Call(c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
					return connect.CreateCitedBridge(tx, userID, connect.CreateInput{
						SourceID: s.source.ID, FromSubjectID: s.person.ID, ToSubjectID: s.event.ID,
						Citation: citations.CreateInput{ArtifactID: s.artifact.ID, LocatorJSON: validLocator},
						Observations: []observations.Input{
							{PropertyID: personProp.ID, ValueSubjectID: s.person.ID},
							{PropertyID: eventProp.ID, ValueSubjectID: s.event.ID},
							{PropertyID: roleProp.ID, ValueTermID: roleTerm.ID},
						},
					})
				})
				if err != nil {
					t.Fatal(err)
				}
				var edge observations.Observation
				var role observations.Observation
				for _, o := range bridge.Observations {
					if string(o.PropertyID) == string(personProp.ID) {
						edge = o
					}
					if string(o.PropertyID) == string(roleProp.ID) {
						role = o
					}
				}
				if len(edge.ID) != 16 || len(role.ID) != 16 {
					t.Fatalf("missing rows edge=%x role=%x", edge.ID, role.ID)
				}
				if _, err := writes.Call(c, writes.Op{Action: "delete_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := observations.Delete(tx, userID, edge.ID)
					return struct{}{}, changes, err
				}); !errors.Is(err, observations.ErrEdgeLocked) {
					t.Fatalf("delete edge %v", err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
					return observations.Update(tx, userID, observations.Input{
						ID: edge.ID, SubjectID: edge.SubjectID, PropertyID: edge.PropertyID,
						ValueSubjectID: s.person.ID,
					})
				}); !errors.Is(err, observations.ErrEdgeLocked) {
					t.Fatalf("update edge no-op %v", err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
					return observations.Update(tx, userID, observations.Input{
						ID: edge.ID, SubjectID: edge.SubjectID, PropertyID: edge.PropertyID,
						ValueSubjectID: s.event.ID,
					})
				}); !errors.Is(err, observations.ErrEdgeLocked) {
					t.Fatalf("update edge retarget %v", err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
					return observations.Update(tx, userID, observations.Input{
						ID: role.ID, SubjectID: role.SubjectID, PropertyID: role.PropertyID,
						ValueTermID: roleTerm.ID,
					})
				}); err != nil {
					t.Fatalf("role update %v", err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "delete_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := observations.Delete(tx, userID, role.ID)
					return struct{}{}, changes, err
				}); err != nil {
					t.Fatalf("role delete %v", err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "add_observations", UserID: userID}, func(tx *database.Tx) ([]observations.Observation, []rowchange.Change, error) {
					return observations.AddToCitation(tx, userID, bridge.Citation.ID, []observations.Input{{
						SubjectID: bridge.Subject.ID, PropertyID: personProp.ID, ValueSubjectID: s.person.ID,
					}})
				}); !errors.Is(err, observations.ErrEdgeLocked) {
					t.Fatalf("add edge %v", err)
				}
				ordinary, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
					return citations.CreateWithObservations(tx, userID, citations.CreateInput{
						ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
					}, []observations.Input{{
						SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Leeds", HasText: true,
					}})
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
					return observations.Update(tx, userID, observations.Input{
						ID: ordinary.Observations[0].ID, SubjectID: bridge.Subject.ID, PropertyID: personProp.ID,
						ValueSubjectID: s.person.ID,
					})
				}); !errors.Is(err, observations.ErrEdgeLocked) {
					t.Fatalf("move ordinary onto edge %v", err)
				}
			},
		},
		{
			name: "likeness saves its region and image",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				art := observationstest.FileArtifact(t, c, userID, s.source.ID)
				res, changes, err := observationstest.SaveCitation(c, userID, art, observations.Input{
					SubjectID: s.person.ID, PropertyID: s.likeness.ID,
					Depiction: &observations.DepictionInput{RegionJSON: observationstest.LikenessRegion, ImagePath: observationstest.Crop(t, 300, 200, 128)},
				})
				if err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var transcription sql.NullString
				if err := db.QueryRow(`SELECT transcription FROM citations WHERE id = ?`, res.Citation.ID).Scan(&transcription); err != nil || transcription.Valid {
					t.Fatalf("transcription %+v %v", transcription, err)
				}
				obs := res.Observations[0]
				var text, integer, dateID, nameID, subjectID, termID, depictionID any
				if err := db.QueryRow(`SELECT value_text, value_integer, value_date_id, value_name_id, value_subject_id, value_term_id, value_depiction_id
					FROM observations WHERE id = ?`, obs.ID).Scan(&text, &integer, &dateID, &nameID, &subjectID, &termID, &depictionID); err != nil {
					t.Fatal(err)
				}
				if text != nil || integer != nil || dateID != nil || nameID != nil || subjectID != nil || termID != nil || depictionID == nil {
					t.Fatalf("columns %v %v %v %v %v %v %v", text, integer, dateID, nameID, subjectID, termID, depictionID)
				}
				var region string
				var imageID []byte
				if err := db.QueryRow(`SELECT region_json, image_file_id FROM depiction_values WHERE id = ?`, obs.ValueDepictionID).Scan(&region, &imageID); err != nil {
					t.Fatal(err)
				}
				file, err := files.Lookup(c, imageID)
				if err != nil || region != observationstest.LikenessRegion || file.MediaType != "image/png" {
					t.Fatalf("region=%q file=%+v %v", region, file, err)
				}
				if got, want := changeKinds(changes, "observation", "file", "depiction_value"),
					[]string{"observation/create", "file/create", "depiction_value/create"}; !sameStrings(got, want) {
					t.Fatalf("changes %v want %v", got, want)
				}
				list, err := observations.ListByCitation(c, res.Citation.ID)
				if err != nil || len(list) != 1 {
					t.Fatalf("%v len=%d", err, len(list))
				}
				row := list[0]
				if string(row.ValueDepictionID) != string(obs.ValueDepictionID) || row.Depiction == nil ||
					row.Depiction.RegionJSON != observationstest.LikenessRegion || string(row.Depiction.ImageFileID) != string(imageID) {
					t.Fatalf("listed %+v depiction %+v", row, row.Depiction)
				}
			},
		},
		{
			name: "likeness read drops a stored region that is not valid",
			run: func(t *testing.T) {
				const pt = `{"type":"region","unit":"normalized","points":`
				type damage struct {
					name string
					// update damages the stored row; artFileID is the Artifact's own File.
					update func(db *sql.DB, depictionID, artFileID []byte) error
					// keepRegion and keepImage say which half the read must still return.
					keepRegion, keepImage bool
				}
				setRegion := func(region string) func(*sql.DB, []byte, []byte) error {
					return func(db *sql.DB, depictionID, _ []byte) error {
						_, err := db.Exec(`UPDATE depiction_values SET region_json = ? WHERE id = ?`, region, depictionID)
						return err
					}
				}
				var damages []damage
				for _, bad := range []struct{ name, region string }{
					{"not json", `not json`},
					{"point out of range", `{"version":1,"selectors":[` + pt + `[{"x":5,"y":5},{"x":0.4,"y":0.1},{"x":0.4,"y":0.5}]}]}`},
					{"self-intersecting", `{"version":1,"selectors":[` + pt + `[{"x":0.1,"y":0.1},{"x":0.5,"y":0.5},{"x":0.5,"y":0.1},{"x":0.1,"y":0.5}]}]}`},
					{"no region selector", `{"version":1,"selectors":[{"type":"page","artifact_page":3}]}`},
					{"page differs from the Citation's", strings.Replace(observationstest.LikenessRegion, `"artifact_page":1`, `"artifact_page":9`, 1)},
				} {
					damages = append(damages, damage{name: bad.name, update: setRegion(bad.region), keepImage: true})
				}
				damages = append(damages, damage{
					name: "image is not a stored PNG",
					update: func(db *sql.DB, depictionID, artFileID []byte) error {
						if _, err := db.Exec(`UPDATE files SET media_type = 'application/pdf' WHERE id = ?`, artFileID); err != nil {
							return err
						}
						_, err := db.Exec(`UPDATE depiction_values SET image_file_id = ? WHERE id = ?`, artFileID, depictionID)
						return err
					},
					keepRegion: true,
				})
				for _, bad := range damages {
					t.Run(bad.name, func(t *testing.T) {
						c, s := mustSeed(t)
						art := observationstest.FileArtifact(t, c, userID, s.source.ID)
						res, _, err := observationstest.SaveCitation(c, userID, art, observations.Input{
							SubjectID: s.person.ID, PropertyID: s.likeness.ID,
							Depiction: &observations.DepictionInput{RegionJSON: observationstest.LikenessRegion, ImagePath: observationstest.Crop(t, 300, 200, 128)},
						})
						if err != nil {
							t.Fatal(err)
						}
						obs := res.Observations[0]
						db, err := c.DB()
						if err != nil {
							t.Fatal(err)
						}
						var imageID []byte
						if err := db.QueryRow(`SELECT image_file_id FROM depiction_values WHERE id = ?`, obs.ValueDepictionID).Scan(&imageID); err != nil {
							t.Fatal(err)
						}
						if err := bad.update(db, obs.ValueDepictionID, art.FileID); err != nil {
							t.Fatal(err)
						}
						byCitation, err := observations.ListByCitation(c, res.Citation.ID)
						if err != nil {
							t.Fatal(err)
						}
						bySubject, err := observations.ListBySubject(c, s.person.ID)
						if err != nil {
							t.Fatal(err)
						}
						for name, list := range map[string][]observations.Listed{"citation": byCitation, "subject": bySubject} {
							if len(list) != 1 {
								t.Fatalf("%s: len=%d", name, len(list))
							}
							row := list[0]
							if string(row.ValueDepictionID) != string(obs.ValueDepictionID) || row.Depiction == nil {
								t.Fatalf("%s: listed %+v depiction %+v", name, row, row.Depiction)
							}
							if keptImage := string(row.Depiction.ImageFileID) == string(imageID); keptImage != bad.keepImage ||
								(len(row.Depiction.ImageFileID) > 0) != bad.keepImage {
								t.Fatalf("%s: image %x, keep=%v", name, row.Depiction.ImageFileID, bad.keepImage)
							}
							if kept := row.Depiction.RegionJSON == observationstest.LikenessRegion; kept != bad.keepRegion ||
								(row.Depiction.RegionJSON != "") != bad.keepRegion {
								t.Fatalf("%s: region %q, keep=%v", name, row.Depiction.RegionJSON, bad.keepRegion)
							}
						}
					})
				}
			},
		},
		{
			name: "likeness refuses",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				art := observationstest.FileArtifact(t, c, userID, s.source.ID)
				good := observationstest.Crop(t, 300, 200, 128)
				symlink := filepath.Join(t.TempDir(), "link.png")
				if err := os.Symlink(good, symlink); err != nil {
					t.Fatal(err)
				}
				html := filepath.Join(t.TempDir(), "page.png")
				if err := os.WriteFile(html, []byte("<html><body>not an image</body></html>"), 0o644); err != nil {
					t.Fatal(err)
				}
				page2 := `{"version":1,"selectors":[{"type":"page","artifact_page":2},{"type":"region","unit":"normalized","points":[{"x":0.1,"y":0.1},{"x":0.4,"y":0.1},{"x":0.4,"y":0.5}]}]}`
				noRegion := `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`
				like := func(region, path string) observations.Input {
					return observations.Input{SubjectID: s.person.ID, PropertyID: s.likeness.ID,
						Depiction: &observations.DepictionInput{RegionJSON: region, ImagePath: path}}
				}
				withText := like(observationstest.LikenessRegion, good)
				withText.ValueText, withText.HasText = "x", true
				onName := observations.Input{SubjectID: s.person.ID, PropertyID: s.nameProp.ID,
					Depiction: &observations.DepictionInput{RegionJSON: observationstest.LikenessRegion, ImagePath: good}}
				unstaged := like(observationstest.LikenessRegion, good)
				tests := []struct {
					name   string
					art    artifacts.Artifact
					inputs []observations.Input
					want   error
					// unstaged skips StageDepictions
					unstaged bool
				}{
					{name: "symlink crop", art: art, inputs: []observations.Input{like(observationstest.LikenessRegion, symlink)}, want: ingest.ErrSymlink},
					{name: "html named png", art: art, inputs: []observations.Input{like(observationstest.LikenessRegion, html)}, want: ingest.ErrUnsupportedImage},
					{name: "region on another page", art: art, inputs: []observations.Input{like(page2, good)}, want: locator.ErrInvalid},
					{name: "region with no region selector", art: art, inputs: []observations.Input{like(noRegion, good)}, want: locator.ErrInvalid},
					{name: "region not a locator", art: art, inputs: []observations.Input{like(`not json`, good)}, want: locator.ErrInvalid},
					{name: "artifact with no file", art: s.artifact, inputs: []observations.Input{like(observationstest.LikenessRegion, good)}, want: observations.ErrInvalid},
					{name: "likeness with text", art: art, inputs: []observations.Input{withText}, want: observations.ErrInvalid},
					{name: "likeness with no depiction", art: art, inputs: []observations.Input{{SubjectID: s.person.ID, PropertyID: s.likeness.ID}}, want: observations.ErrInvalid},
					{name: "depiction on a name", art: art, inputs: []observations.Input{onName}, want: observations.ErrInvalid},
					{name: "unstaged depiction", art: art, inputs: []observations.Input{unstaged}, want: observations.ErrInvalid, unstaged: true},
					{name: "second input bad", art: art, want: observations.ErrInvalid, inputs: []observations.Input{
						like(observationstest.LikenessRegion, good),
						{SubjectID: s.person.ID, PropertyID: s.sexProp.ID},
					}},
				}
				for _, tt := range tests {
					t.Run(tt.name, func(t *testing.T) {
						filesBefore := rowCount(t, c, `SELECT COUNT(*) FROM files`)
						valuesBefore := rowCount(t, c, `SELECT COUNT(*) FROM depiction_values`)
						objectsBefore := objectCount(t, c)
						var err error
						if tt.unstaged {
							_, err = writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
								return citations.CreateWithObservations(tx, userID, citations.CreateInput{ArtifactID: tt.art.ID, LocatorJSON: observationstest.LikenessLocator}, tt.inputs)
							})
						} else {
							_, _, err = observationstest.SaveCitation(c, userID, tt.art, tt.inputs...)
						}
						if !errors.Is(err, tt.want) {
							t.Fatalf("got %v want %v", err, tt.want)
						}
						if got := rowCount(t, c, `SELECT COUNT(*) FROM files`); got != filesBefore {
							t.Fatalf("files %d want %d", got, filesBefore)
						}
						if got := rowCount(t, c, `SELECT COUNT(*) FROM depiction_values`); got != valuesBefore {
							t.Fatalf("depiction_values %d want %d", got, valuesBefore)
						}
						if got := objectCount(t, c); got != objectsBefore {
							t.Fatalf("objects %d want %d", got, objectsBefore)
						}
					})
				}
			},
		},
		{
			name: "likeness update replaces the image",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				art := observationstest.FileArtifact(t, c, userID, s.source.ID)
				res, _, err := observationstest.SaveCitation(c, userID, art, observations.Input{
					SubjectID: s.person.ID, PropertyID: s.likeness.ID,
					Depiction: &observations.DepictionInput{RegionJSON: observationstest.LikenessRegion, ImagePath: observationstest.Crop(t, 300, 200, 128)},
				})
				if err != nil {
					t.Fatal(err)
				}
				obs := res.Observations[0]
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				lookup := func() (valueID []byte, region string, file files.File) {
					var imageID []byte
					if err := db.QueryRow(`SELECT o.value_depiction_id, d.region_json, d.image_file_id
						FROM observations o JOIN depiction_values d ON d.id = o.value_depiction_id WHERE o.id = ?`, obs.ID).Scan(&valueID, &region, &imageID); err != nil {
						t.Fatal(err)
					}
					file, err := files.Lookup(c, imageID)
					if err != nil {
						t.Fatal(err)
					}
					return valueID, region, file
				}
				oldValue, _, oldFile := lookup()
				if !objectExists(t, c, oldFile) {
					t.Fatal("old object missing before the update")
				}

				newRegion := `{"version":1,"selectors":[{"type":"page","artifact_page":1},{"type":"region","unit":"normalized","points":[{"x":0.2,"y":0.2},{"x":0.6,"y":0.2},{"x":0.6,"y":0.7}]}]}`
				changes, err := observationstest.UpdateObservation(c, userID, observations.Input{
					ID: obs.ID, SubjectID: s.person.ID, PropertyID: s.likeness.ID,
					Depiction: &observations.DepictionInput{RegionJSON: newRegion, ImagePath: observationstest.Crop(t, 200, 200, 200)},
				})
				if err != nil {
					t.Fatal(err)
				}
				newValue, region, newFile := lookup()
				if string(newValue) == string(oldValue) || region != newRegion || string(newFile.ID) == string(oldFile.ID) {
					t.Fatalf("value %x->%x region=%q file %x->%x", oldValue, newValue, region, oldFile.ID, newFile.ID)
				}
				if rowCount(t, c, `SELECT COUNT(*) FROM depiction_values WHERE id = ?`, oldValue) != 0 ||
					rowCount(t, c, `SELECT COUNT(*) FROM files WHERE id = ?`, oldFile.ID) != 0 {
					t.Fatal("old value row or file row left behind")
				}
				if objectExists(t, c, oldFile) || !objectExists(t, c, newFile) {
					t.Fatal("old object should be gone and the new one stored")
				}
				if got, want := changeKinds(changes, "observation", "depiction_value", "file"),
					[]string{"observation/update", "depiction_value/delete", "file/create", "depiction_value/create"}; !sameStrings(got, want) {
					t.Fatalf("changes %v want %v", got, want)
				}

				// A crop that re-encodes to the same PNG keeps the File and its object.
				keepValue, _, keepFile := lookup()
				if _, err := observationstest.UpdateObservation(c, userID, observations.Input{
					ID: obs.ID, SubjectID: s.person.ID, PropertyID: s.likeness.ID,
					Depiction: &observations.DepictionInput{RegionJSON: newRegion, ImagePath: observationstest.Crop(t, 200, 200, 200)},
				}); err != nil {
					t.Fatal(err)
				}
				againValue, _, againFile := lookup()
				if string(againValue) == string(keepValue) || string(againFile.ID) != string(keepFile.ID) || !objectExists(t, c, againFile) {
					t.Fatalf("same crop: value %x->%x file %x->%x", keepValue, againValue, keepFile.ID, againFile.ID)
				}

				// A polarity change keeps the value row and its File.
				beforeValue, _, beforeFile := lookup()
				changes, err = observationstest.UpdateObservation(c, userID, observations.Input{
					ID: obs.ID, SubjectID: s.person.ID, PropertyID: s.likeness.ID,
					Polarity: observations.PolarityNegative, ValueDepictionID: beforeValue,
				})
				if err != nil {
					t.Fatal(err)
				}
				afterValue, _, afterFile := lookup()
				if string(afterValue) != string(beforeValue) || string(afterFile.ID) != string(beforeFile.ID) || !objectExists(t, c, afterFile) {
					t.Fatal("polarity change touched the value")
				}
				if got := changeKinds(changes, "depiction_value", "file"); len(got) != 0 {
					t.Fatalf("value changes on a polarity update: %v", got)
				}

				// A value row that is not the Observation's own is refused.
				if _, err := observationstest.UpdateObservation(c, userID, observations.Input{
					ID: obs.ID, SubjectID: s.person.ID, PropertyID: s.likeness.ID, ValueDepictionID: oldValue,
				}); !errors.Is(err, observations.ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "likeness and name swap releases the replaced value",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				art := observationstest.FileArtifact(t, c, userID, s.source.ID)
				res, _, err := observationstest.SaveCitation(c, userID, art, observations.Input{
					SubjectID: s.person.ID, PropertyID: s.nameProp.ID, Name: &namevalues.Value{Form: "Ada"},
				})
				if err != nil {
					t.Fatal(err)
				}
				obs := res.Observations[0]
				nameID := obs.ValueNameID
				objectsBefore := objectCount(t, c)

				changes, err := observationstest.UpdateObservation(c, userID, observations.Input{
					ID: obs.ID, SubjectID: s.person.ID, PropertyID: s.likeness.ID,
					Depiction: &observations.DepictionInput{RegionJSON: observationstest.LikenessRegion, ImagePath: observationstest.Crop(t, 300, 200, 128)},
				})
				if err != nil {
					t.Fatal(err)
				}
				if rowCount(t, c, `SELECT COUNT(*) FROM name_values WHERE id = ?`, nameID) != 0 ||
					rowCount(t, c, `SELECT COUNT(*) FROM depiction_values`) != 1 {
					t.Fatal("name value should be released and one depiction value created")
				}
				if got, want := changeKinds(changes, "name_value", "depiction_value"),
					[]string{"name_value/delete", "depiction_value/create"}; !sameStrings(got, want) {
					t.Fatalf("changes %v want %v", got, want)
				}
				if got := objectCount(t, c); got != objectsBefore+1 {
					t.Fatalf("objects %d want %d", got, objectsBefore+1)
				}

				changes, err = observationstest.UpdateObservation(c, userID, observations.Input{
					ID: obs.ID, SubjectID: s.person.ID, PropertyID: s.nameProp.ID, Name: &namevalues.Value{Form: "Ada"},
				})
				if err != nil {
					t.Fatal(err)
				}
				if rowCount(t, c, `SELECT COUNT(*) FROM depiction_values`) != 0 ||
					rowCount(t, c, `SELECT COUNT(*) FROM files`) != 1 {
					t.Fatal("depiction value and its file should be released")
				}
				if got := changeKinds(changes, "depiction_value"); !sameStrings(got, []string{"depiction_value/delete"}) {
					t.Fatalf("changes %v", got)
				}
				if got := objectCount(t, c); got != objectsBefore {
					t.Fatalf("objects %d want %d", got, objectsBefore)
				}
			},
		},
		{
			name: "likeness delete releases its image",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				art := observationstest.FileArtifact(t, c, userID, s.source.ID)
				objectsBefore := objectCount(t, c)
				res, _, err := observationstest.SaveCitation(c, userID, art, observations.Input{
					SubjectID: s.person.ID, PropertyID: s.likeness.ID,
					Depiction: &observations.DepictionInput{RegionJSON: observationstest.LikenessRegion, ImagePath: observationstest.Crop(t, 300, 200, 128)},
				})
				if err != nil {
					t.Fatal(err)
				}
				if objectCount(t, c) != objectsBefore+1 {
					t.Fatal("image object not stored")
				}
				if err := deleteObservation(c, userID, res.Observations[0].ID); err != nil {
					t.Fatal(err)
				}
				if rowCount(t, c, `SELECT COUNT(*) FROM depiction_values`) != 0 || rowCount(t, c, `SELECT COUNT(*) FROM files`) != 1 {
					t.Fatal("value row or image file left behind")
				}
				if objectCount(t, c) != objectsBefore {
					t.Fatal("image object left behind")
				}
				if _, _, err := writes.Run(c, writes.Op{Action: "delete_citation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := citations.Delete(tx, userID, res.Citation.ID)
					return struct{}{}, changes, err
				}); err != nil {
					t.Fatal(err)
				}

				// Two likenesses whose crops re-encode to the same PNG share one File.
				same := observationstest.Crop(t, 300, 200, 90)
				res, _, err = observationstest.SaveCitation(c, userID, art,
					observations.Input{SubjectID: s.person.ID, PropertyID: s.likeness.ID, Depiction: &observations.DepictionInput{RegionJSON: observationstest.LikenessRegion, ImagePath: same}},
					observations.Input{SubjectID: s.person.ID, PropertyID: s.likeness.ID, Depiction: &observations.DepictionInput{RegionJSON: observationstest.LikenessRegion, ImagePath: same}},
				)
				if err != nil {
					t.Fatal(err)
				}
				if rowCount(t, c, `SELECT COUNT(*) FROM files`) != 2 || objectCount(t, c) != objectsBefore+1 {
					t.Fatal("two likenesses of one crop should share one file")
				}
				if err := deleteObservation(c, userID, res.Observations[0].ID); err != nil {
					t.Fatal(err)
				}
				if rowCount(t, c, `SELECT COUNT(*) FROM files`) != 2 || objectCount(t, c) != objectsBefore+1 {
					t.Fatal("the shared file should stay while the other likeness uses it")
				}
				if err := deleteObservation(c, userID, res.Observations[1].ID); err != nil {
					t.Fatal(err)
				}
				if rowCount(t, c, `SELECT COUNT(*) FROM files`) != 1 || objectCount(t, c) != objectsBefore {
					t.Fatal("the shared file should go with the last likeness")
				}
			},
		},
		{
			name: "likeness blocks subject delete",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				art := observationstest.FileArtifact(t, c, userID, s.source.ID)
				objectsBefore := objectCount(t, c)
				res, _, err := observationstest.SaveCitation(c, userID, art, observations.Input{
					SubjectID: s.person.ID, PropertyID: s.likeness.ID,
					Depiction: &observations.DepictionInput{RegionJSON: observationstest.LikenessRegion, ImagePath: observationstest.Crop(t, 300, 200, 128)},
				})
				if err != nil {
					t.Fatal(err)
				}
				deleteSubject := func() error {
					_, _, err := writes.Run(c, writes.Op{Action: "delete_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
						changes, err := subjects.Delete(tx, userID, s.person.ID)
						return struct{}{}, changes, err
					})
					return err
				}
				if err := deleteSubject(); !errors.Is(err, subjects.ErrInUse) {
					t.Fatalf("got %v want ErrInUse", err)
				}
				if rowCount(t, c, `SELECT COUNT(*) FROM depiction_values`) != 1 || objectCount(t, c) != objectsBefore+1 {
					t.Fatal("a refused delete must leave the value, file, and object")
				}
				if err := deleteObservation(c, userID, res.Observations[0].ID); err != nil {
					t.Fatal(err)
				}
				if err := deleteSubject(); err != nil {
					t.Fatal(err)
				}
				if rowCount(t, c, `SELECT COUNT(*) FROM depiction_values`) != 0 || rowCount(t, c, `SELECT COUNT(*) FROM files`) != 1 || objectCount(t, c) != objectsBefore {
					t.Fatal("value row, file, or object left behind")
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

// objectCount is the number of regular files under the catalog's objects/.
func objectCount(t *testing.T, c *database.Catalog) int {
	t.Helper()
	n := 0
	root := filepath.Join(c.Dir(), "objects")
	err := filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.Type().IsRegular() {
			n++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func rowCount(t *testing.T, c *database.Catalog, query string, args ...any) int {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func deleteObservation(c *database.Catalog, userID, id []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "delete_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		changes, err := observations.Delete(tx, userID, id)
		return struct{}{}, changes, err
	})
	return err
}

// changeKinds is "entity/action" for each change of the given entity types, in order.
func changeKinds(changes []rowchange.Change, entities ...string) []string {
	var out []string
	for _, ch := range changes {
		for _, e := range entities {
			if ch.EntityType == e {
				out = append(out, ch.EntityType+"/"+string(ch.Action))
			}
		}
	}
	return out
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// objectExists reports whether the stored object of file f is on disk.
func objectExists(t *testing.T, c *database.Catalog, f files.File) bool {
	t.Helper()
	rel, err := files.StorageRelPath(f.ChecksumSHA256, f.MediaType)
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Stat(filepath.Join(c.Dir(), rel))
	return err == nil
}

func runArtifactCreate(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error) {
	a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, in)
		})
	return a, err
}
