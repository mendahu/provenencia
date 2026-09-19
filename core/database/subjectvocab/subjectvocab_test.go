package subjectvocab

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
)

func TestSubjectVocab(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "install seeds types properties bindings",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				types, err := subjecttypes.List(c)
				if err != nil || len(types) != 7 {
					t.Fatalf("types %v len=%d", err, len(types))
				}
				props, err := properties.List(c)
				if err != nil || len(props) != len(seedProperties) {
					t.Fatalf("props %v len=%d want %d", err, len(props), len(seedProperties))
				}
				person, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				bindings, err := ListBindings(c, person.ID)
				if err != nil || len(bindings) != 3 {
					t.Fatalf("person bindings %v len=%d", err, len(bindings))
				}
				participation, err := subjecttypes.Lookup(c, "participation", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				pb, err := ListBindings(c, participation.ID)
				if err != nil {
					t.Fatal(err)
				}
				locked := 0
				for _, b := range pb {
					if b.Locked {
						locked++
					}
				}
				if locked != 2 {
					t.Fatalf("participation locked=%d want 2", locked)
				}
			},
		},
		{
			name: "open does not heal deleted property",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				prop, err := properties.Lookup(c, "occupation", properties.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				// Unbind person.occupation first so delete is allowed.
				person, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if err := DeleteBinding(c, person.ID, prop.ID); err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`DELETE FROM properties WHERE id = ?`, prop.ID); err != nil {
					t.Fatal(err)
				}
				dir := c.Dir()
				_ = c.Close()
				reopened, err := database.Open(dir)
				if err != nil {
					t.Fatal(err)
				}
				defer reopened.Close()
				_, err = properties.Lookup(reopened, "occupation", properties.OriginProvenencia)
				if err == nil {
					t.Fatal("healed occupation on open")
				}
			},
		},
		{
			name: "refuse remove locked binding",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				ptn, err := subjecttypes.Lookup(c, "participation", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				personProp, err := properties.Lookup(c, "person", properties.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if err := DeleteBinding(c, ptn.ID, personProp.ID); !errors.Is(err, ErrLocked) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "registry placeable presentation connect",
			run: func(t *testing.T, _ *database.Catalog) {
				placeable := PlaceableTypes()
				if len(placeable) != 3 {
					t.Fatalf("placeable=%d", len(placeable))
				}
				if placeable[0].Key != "person" || placeable[1].Key != "event" || placeable[2].Key != "place" {
					t.Fatalf("%+v", placeable)
				}
				pres, ok := PresentationFor("person")
				if !ok || pres.InkToken != "subjectPersonInk" || pres.IconSymbol != "person" {
					t.Fatalf("%+v ok=%v", pres, ok)
				}
				for _, info := range AllTypes() {
					if info.Presentation.L10nKey == "" || info.Presentation.IconSymbol == "" {
						t.Fatalf("missing presentation %+v", info)
					}
				}
				if !LockedBinding("participation", "person") {
					t.Fatal("expected locked")
				}
				rule := Connect("person", "event")
				if rule.Refuse || rule.BridgeTypeKey != "participation" || rule.Disambiguation != DisambiguationRole {
					t.Fatalf("%+v", rule)
				}
				refuse := Connect("person", "place")
				if !refuse.Refuse {
					t.Fatalf("%+v", refuse)
				}
				unknown := Connect("person", "source")
				if !unknown.Refuse {
					t.Fatalf("%+v", unknown)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := database.Create(t.TempDir(), "t.provenencia")
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			tt.run(t, c)
		})
	}
}
