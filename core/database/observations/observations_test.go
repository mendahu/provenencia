package observations_test

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
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
		toponymProp properties.Property
		sexProp     properties.Property
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
		maleTerm, err := propertyterms.Lookup(c, sexProp.ID, "male", propertyterms.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		return c, seed{
			source: src, artifact: art, person: person, place: place,
			toponymProp: toponymProp, sexProp: sexProp,
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
						sawTerm = true
					}
				}
				if !sawText || !sawTerm {
					t.Fatalf("text=%v term=%v rows=%+v", sawText, sawTerm, list)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
