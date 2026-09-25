package subjectvocab

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
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
				et, err := properties.Lookup(c, "event_type", properties.OriginProvenencia)
				if err != nil || et.ValueType != properties.ValueTypeTerm {
					t.Fatalf("event_type %+v %v", et, err)
				}
				terms, err := propertyterms.ListByProperty(c, et.ID)
				if err != nil || len(terms) == 0 {
					t.Fatalf("event_type terms %v len=%d", err, len(terms))
				}
				birth, err := propertyterms.Lookup(c, et.ID, "birth", propertyterms.OriginProvenencia)
				if err != nil || birth.Label != "Birth" {
					t.Fatalf("birth %+v %v", birth, err)
				}
				person, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				bindings, err := ListBindings(c, person.ID)
				if err != nil {
					t.Fatal(err)
				}
				got := map[string]bool{}
				for _, b := range bindings {
					got[b.Property.Key] = true
				}
				for _, key := range []string{"name", "sex_at_birth"} {
					if !got[key] {
						t.Fatalf("person missing binding %q", key)
					}
				}
				evt, err := subjecttypes.Lookup(c, "event", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				eb, err := ListBindings(c, evt.ID)
				if err != nil || len(eb) != 4 {
					t.Fatalf("event bindings %v len=%d want 4", err, len(eb))
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
				prop, err := properties.Lookup(c, "remark", properties.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				// Unbind source.remark first so the Property can be deleted.
				srcType, err := subjecttypes.Lookup(c, "source", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if err := DeleteBinding(c, srcType.ID, prop.ID); err != nil {
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
				_, err = properties.Lookup(reopened, "remark", properties.OriginProvenencia)
				if err == nil {
					t.Fatal("healed remark on open")
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
				if !ok || pres.InkToken != "subjectPersonInk" || pres.IconSymbol != "subject_person" {
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
				if !LockedBinding("event", "date") || !LockedBinding("event", "start_date") || !LockedBinding("event", "end_date") {
					t.Fatal("expected event date Properties locked")
				}
				rule := Connect("person", "event")
				if rule.Refuse || rule.BridgeTypeKey != "participation" || rule.Disambiguation != DisambiguationRole {
					t.Fatalf("%+v", rule)
				}
				if len(rule.Edges) != 2 || rule.Edges[0].PropertyKey != "person" || rule.Edges[0].EndpointTypeKey != "person" ||
					rule.Edges[1].PropertyKey != "event" || rule.Edges[1].EndpointTypeKey != "event" {
					t.Fatalf("participation edges %+v", rule.Edges)
				}
				rel := Connect("person", "person")
				if len(rel.Edges) != 2 || rel.Edges[0].PropertyKey != "person" || rel.Edges[1].PropertyKey != "related_to" ||
					rel.Edges[1].EndpointTypeKey != "person" {
					t.Fatalf("relationship edges %+v", rel.Edges)
				}
				loc := Connect("event", "place")
				if len(loc.Edges) != 2 || loc.Edges[0].PropertyKey != "event" || loc.Edges[1].PropertyKey != "place" {
					t.Fatalf("location edges %+v", loc.Edges)
				}
				if endpoint, ok := EdgeEndpoint("participation", "person"); !ok || endpoint != "person" {
					t.Fatalf("EdgeEndpoint participation/person %q %v", endpoint, ok)
				}
				if endpoint, ok := EdgeEndpoint("relationship", "related_to"); !ok || endpoint != "person" {
					t.Fatalf("EdgeEndpoint relationship/related_to %q %v", endpoint, ok)
				}
				if endpoint, ok := EdgeEndpoint("location", "place"); !ok || endpoint != "place" {
					t.Fatalf("EdgeEndpoint location/place %q %v", endpoint, ok)
				}
				if _, ok := EdgeEndpoint("participation", "role"); ok {
					t.Fatal("role is not an edge")
				}
				rules := ListConnectRules()
				if len(rules) == 0 || len(rules[0].Edges) != len(rules[0].EdgePropertyKeys) {
					t.Fatalf("ListConnectRules edges %+v", rules)
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
