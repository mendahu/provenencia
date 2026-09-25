package observations_test

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
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
	"github.com/mendahu/provenencia/core/ref"
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
		src, err := sources.Create(c, userID, sources.CreateInput{
			SourceTypeID: typeID,
			Title:        "Deed book",
		})
		if err != nil {
			t.Fatal(err)
		}
		art, err := artifacts.Create(c, userID, artifacts.CreateInput{
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
		person, err := subjects.Create(c, userID, subjects.CreateInput{
			SourceID: src.ID, SubjectTypeID: personType.ID, Label: "Bob",
		})
		if err != nil {
			t.Fatal(err)
		}
		place, err := subjects.Create(c, userID, subjects.CreateInput{
			SourceID: src.ID, SubjectTypeID: placeType.ID,
		})
		if err != nil {
			t.Fatal(err)
		}
		event, err := subjects.Create(c, userID, subjects.CreateInput{
			SourceID: src.ID, SubjectTypeID: eventType.ID, Label: "Birth",
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
			toponymProp: toponymProp, sexProp: sexProp, nameProp: nameProp, dateProp: dateProp,
			femaleTerm: femaleTerm, maleTerm: maleTerm,
		}
	}
	mustCitation := func(t *testing.T, c *database.Catalog, s seed, obs ...observations.Input) citations.Citation {
		t.Helper()
		res, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
			ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
		}, obs)
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
				added, err := observations.AddToCitation(c, userID, cit.ID, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.maleTerm.ID,
					Polarity:    observations.PolarityNegative,
				}})
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
				unbound, err := properties.Create(c, userID, "Nickname", properties.ValueTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				cit := mustCitation(t, c, s, observations.Input{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				})
				_, err = observations.AddToCitation(c, userID, cit.ID, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: unbound.ID,
					ValueText: "Al", HasText: true,
				}})
				if !errors.Is(err, observations.ErrInvalid) {
					t.Fatalf("got %v want ErrInvalid", err)
				}
			},
		},
		{
			name: "list by source",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				mustCitation(t, c, s,
					observations.Input{
						SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
						ValueTermID: s.femaleTerm.ID,
					},
					observations.Input{
						SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Cambridge", HasText: true,
					},
				)
				list, err := observations.ListBySource(c, s.source.ID)
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
				mustCitation(t, c, s,
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
				list, err := observations.ListBySource(c, s.source.ID)
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
				_, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
					ValueText: "   ", HasText: true,
				}})
				if !errors.Is(err, observations.ErrInvalid) {
					t.Fatalf("got %v want ErrInvalid", err)
				}
				_, err = citations.CreateWithObservations(c, userID, citations.CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
					HasText: false,
				}})
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
				res, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.event.ID, PropertyID: s.dateProp.ID,
					Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year},
				}})
				if err != nil {
					t.Fatal(err)
				}
				year2 := 1843
				got, err := observations.Update(c, userID, observations.Input{
					ID: res.Observations[0].ID, SubjectID: s.event.ID, PropertyID: s.dateProp.ID,
					Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year2},
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
			name: "update no-op records no revision",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
					ValueText: "Boston", HasText: true,
				}})
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
				if _, err := observations.Update(c, userID, observations.Input{
					ID: res.Observations[0].ID, SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
					ValueText: "Boston", HasText: true,
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
				res, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
					Notes: []string{"keep me"},
				}, []observations.Input{{
					SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
					ValueText: "Boston", HasText: true,
					Notes: []string{"obs note"},
				}})
				if err != nil {
					t.Fatal(err)
				}
				if err := observations.Delete(c, userID, res.Observations[0].ID); err != nil {
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
				bridge, err := connect.CreateCitedBridge(c, userID, connect.CreateInput{
					SourceID: s.source.ID, FromSubjectID: s.person.ID, ToSubjectID: s.event.ID,
					Citation: citations.CreateInput{ArtifactID: s.artifact.ID, LocatorJSON: validLocator},
					Observations: []observations.Input{
						{PropertyID: personProp.ID, ValueSubjectID: s.person.ID},
						{PropertyID: eventProp.ID, ValueSubjectID: s.event.ID},
						{PropertyID: roleProp.ID, ValueTermID: roleTerm.ID},
					},
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
				if err := observations.Delete(c, userID, edge.ID); !errors.Is(err, observations.ErrEdgeLocked) {
					t.Fatalf("delete edge %v", err)
				}
				if _, err := observations.Update(c, userID, observations.Input{
					ID: edge.ID, SubjectID: edge.SubjectID, PropertyID: edge.PropertyID,
					ValueSubjectID: s.person.ID,
				}); !errors.Is(err, observations.ErrEdgeLocked) {
					t.Fatalf("update edge no-op %v", err)
				}
				if _, err := observations.Update(c, userID, observations.Input{
					ID: edge.ID, SubjectID: edge.SubjectID, PropertyID: edge.PropertyID,
					ValueSubjectID: s.event.ID,
				}); !errors.Is(err, observations.ErrEdgeLocked) {
					t.Fatalf("update edge retarget %v", err)
				}
				if _, err := observations.Update(c, userID, observations.Input{
					ID: role.ID, SubjectID: role.SubjectID, PropertyID: role.PropertyID,
					ValueTermID: roleTerm.ID,
				}); err != nil {
					t.Fatalf("role update %v", err)
				}
				if err := observations.Delete(c, userID, role.ID); err != nil {
					t.Fatalf("role delete %v", err)
				}
				if _, err := observations.AddToCitation(c, userID, bridge.Citation.ID, []observations.Input{{
					SubjectID: bridge.Subject.ID, PropertyID: personProp.ID, ValueSubjectID: s.person.ID,
				}}); !errors.Is(err, observations.ErrEdgeLocked) {
					t.Fatalf("add edge %v", err)
				}
				ordinary, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
					ValueText: "Leeds", HasText: true,
				}})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := observations.Update(c, userID, observations.Input{
					ID: ordinary.Observations[0].ID, SubjectID: bridge.Subject.ID, PropertyID: personProp.ID,
					ValueSubjectID: s.person.ID,
				}); !errors.Is(err, observations.ErrEdgeLocked) {
					t.Fatalf("move ordinary onto edge %v", err)
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
