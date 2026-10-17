package subjectvocab

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
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
				if err != nil || len(types) != 8 {
					t.Fatalf("types %v len=%d", err, len(types))
				}
				props, err := properties.List(c)
				if err != nil || len(props) != len(seedProperties) {
					t.Fatalf("props %v len=%d want %d", err, len(props), len(seedProperties))
				}
				et, err := properties.Lookup(c, "event_type", properties.OriginProvenencia)
				if err != nil || et.ValueType != propertymeta.ValueTypeTerm {
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
				if err != nil || len(eb) != 5 {
					t.Fatalf("event bindings %v len=%d want 5", err, len(eb))
				}
				en, err := properties.Lookup(c, "event_name", properties.OriginProvenencia)
				if err != nil || en.ValueType != propertymeta.ValueTypeText || en.Origin != properties.OriginProvenencia {
					t.Fatalf("event_name %+v %v", en, err)
				}
				var nameBinding *Binding
				for i := range eb {
					if eb[i].Property.Key == "event_name" {
						nameBinding = &eb[i]
						break
					}
				}
				if nameBinding == nil || nameBinding.Locked || nameBinding.SortOrder != 1 {
					t.Fatalf("event_name binding %+v", nameBinding)
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
				if locked != 3 {
					t.Fatalf("participation locked=%d want 3", locked)
				}
				prt, err := properties.Lookup(c, "place_relationship_type", properties.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				partOf, err := propertyterms.Lookup(c, prt.ID, "part_of", propertyterms.OriginProvenencia)
				if err != nil || !partOf.Directed {
					t.Fatalf("part_of %+v %v", partOf, err)
				}
				succ, err := propertyterms.Lookup(c, prt.ID, "succeeded_by", propertyterms.OriginProvenencia)
				if err != nil || !succ.Directed {
					t.Fatalf("succeeded_by %+v %v", succ, err)
				}
				placeType, err := subjecttypes.Lookup(c, "place", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				placeBindings, err := ListBindings(c, placeType.ID)
				if err != nil {
					t.Fatal(err)
				}
				gotPlace := map[string]bool{}
				for _, b := range placeBindings {
					gotPlace[b.Property.Key] = true
				}
				for _, key := range []string{"toponym", "start_date", "end_date"} {
					if !gotPlace[key] {
						t.Fatalf("place missing binding %q", key)
					}
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
					if info.Presentation.IconSymbol != "subject_"+info.Key {
						t.Fatalf("%s IconSymbol %q, want subject_%s (Swift maps marks by this rule)", info.Key, info.Presentation.IconSymbol, info.Key)
					}
				}
				if !LockedBinding("participation", "person") {
					t.Fatal("expected locked")
				}
				if !LockedBinding("event", "event_type") || !LockedBinding("event", "date") || !LockedBinding("event", "start_date") || !LockedBinding("event", "end_date") {
					t.Fatal("expected event_type and event date Properties locked")
				}
				if !LockedBinding("participation", "role") || !LockedBinding("relationship", "relationship_type") {
					t.Fatal("expected role and relationship_type locked")
				}
				if !LockedBinding("person", "name") || !LockedBinding("place", "toponym") {
					t.Fatal("expected name and toponym locked")
				}
				if !LockedBinding("place", "start_date") || !LockedBinding("place", "end_date") {
					t.Fatal("expected place period bindings locked")
				}
				if !LockedBinding("place_relationship", "from") || !LockedBinding("place_relationship", "to") ||
					!LockedBinding("place_relationship", "place_relationship_type") ||
					!LockedBinding("place_relationship", "start_date") || !LockedBinding("place_relationship", "end_date") {
					t.Fatal("expected place_relationship bindings locked")
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
				placeRel := Connect("place", "place")
				if placeRel.Refuse || placeRel.BridgeTypeKey != "place_relationship" ||
					placeRel.Disambiguation != DisambiguationPlaceRelationshipType {
					t.Fatalf("%+v", placeRel)
				}
				if len(placeRel.Edges) != 2 || placeRel.Edges[0].PropertyKey != "from" || placeRel.Edges[1].PropertyKey != "to" {
					t.Fatalf("place_relationship edges %+v", placeRel.Edges)
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
				if endpoint, ok := EdgeEndpoint("place_relationship", "from"); !ok || endpoint != "place" {
					t.Fatalf("EdgeEndpoint place_relationship/from %q %v", endpoint, ok)
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

// An inverse names a directed term of the same Property that names this one
// back, so matching can read either end of one relationship.
func TestSeedInversesPairUp(t *testing.T) {
	byKey := map[string]seedTerm{}
	for _, term := range seedTerms {
		byKey[term.PropertyKey+"|"+term.Key] = term
	}
	pairs := 0
	for _, term := range seedTerms {
		if term.Inverse == "" {
			continue
		}
		pairs++
		other, ok := byKey[term.PropertyKey+"|"+term.Inverse]
		if !ok {
			t.Fatalf("%s: inverse %q is not a %s term", term.Key, term.Inverse, term.PropertyKey)
		}
		if other.Inverse != term.Key || !term.Directed || !other.Directed {
			t.Fatalf("%s ↔ %s: inverses must be directed and name each other", term.Key, other.Key)
		}
	}
	if pairs != 8 {
		t.Fatalf("inverse terms %d, want 8 (four kinship pairs)", pairs)
	}
}

func TestInstallStoresInverses(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	if err := Install(c); err != nil {
		t.Fatal(err)
	}
	rt, err := properties.Lookup(c, "relationship_type", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"parent": "child", "ward": "guardian", "spouse": ""} {
		term, err := propertyterms.Lookup(c, rt.ID, key, propertyterms.OriginProvenencia)
		if err != nil || term.InverseKey != want {
			t.Fatalf("%s inverse %q (%v), want %q", key, term.InverseKey, err, want)
		}
	}
}

func TestInstallFlagsNeutralTerms(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	if err := Install(c); err != nil {
		t.Fatal(err)
	}
	sex, err := properties.Lookup(c, "sex_at_birth", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]bool{"unknown": true, "indeterminate": true, "male": false, "female": false} {
		term, err := propertyterms.Lookup(c, sex.ID, key, propertyterms.OriginProvenencia)
		if err != nil || term.Neutral != want {
			t.Fatalf("%s neutral %v (%v), want %v", key, term.Neutral, err, want)
		}
	}
}

func TestInstallSeedsEngagementTravelInterestsAndGender(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	if err := Install(c); err != nil {
		t.Fatal(err)
	}

	eventType, err := properties.Lookup(c, "event_type", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	for key, label := range map[string]string{"engagement": "Engagement", "travel": "Travel"} {
		term, err := propertyterms.Lookup(c, eventType.ID, key, propertyterms.OriginProvenencia)
		if err != nil || term.Label != label {
			t.Fatalf("%s %+v %v", key, term, err)
		}
	}

	interests, err := properties.Lookup(c, "interests", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	if interests.ValueType != propertymeta.ValueTypeText || interests.Cardinality != propertymeta.CardinalityMultiple {
		t.Fatalf("interests %+v", interests)
	}
	if LockedBinding("person", "interests") {
		t.Fatal("interests binding is locked")
	}

	gender, err := properties.Lookup(c, "gender", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	if gender.ValueType != propertymeta.ValueTypeTerm || gender.Cardinality != propertymeta.CardinalitySingle {
		t.Fatalf("gender %+v", gender)
	}
	if LockedBinding("person", "gender") {
		t.Fatal("gender binding is locked")
	}
	for _, key := range []string{"man", "woman", "non_binary"} {
		term, err := propertyterms.Lookup(c, gender.ID, key, propertyterms.OriginProvenencia)
		if err != nil || term.Neutral {
			t.Fatalf("%s %+v %v", key, term, err)
		}
	}

	person, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	bindings, err := ListBindings(c, person.ID)
	if err != nil {
		t.Fatal(err)
	}
	sortByKey := map[string]int{}
	for _, b := range bindings {
		sortByKey[b.Property.Key] = b.SortOrder
		if (b.Property.Key == "gender" || b.Property.Key == "interests") && b.Locked {
			t.Fatalf("%s binding locked", b.Property.Key)
		}
	}
	if sortByKey["gender"] != 2 || sortByKey["interests"] != 3 {
		t.Fatalf("person sort orders %v", sortByKey)
	}

	userID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	userRef, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Tester", userRef); err != nil {
		t.Fatal(err)
	}
	created, _, err := writes.Run(c, writes.Op{Action: "create_property_term", UserID: userID},
		func(tx *database.Tx) (propertyterms.Term, []rowchange.Change, error) {
			return propertyterms.Create(tx, userID, gender.ID, "Agender", "")
		})
	if err != nil || created.Key != "agender" || created.Origin != propertyterms.OriginUser {
		t.Fatalf("create gender term %+v %v", created, err)
	}
}

func TestInstallSeedsLikeness(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	if err := Install(c); err != nil {
		t.Fatal(err)
	}
	likeness, err := properties.Lookup(c, "likeness", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	if likeness.ValueType != propertymeta.ValueTypeDepiction || likeness.Cardinality != propertymeta.CardinalityMultiple || likeness.Label != "Likeness" {
		t.Fatalf("likeness %+v", likeness)
	}
	if !LockedBinding("person", "likeness") {
		t.Fatal("likeness binding is not locked")
	}
	person, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	bindings, err := ListBindings(c, person.ID)
	if err != nil {
		t.Fatal(err)
	}
	sort := -1
	for _, b := range bindings {
		if b.Property.Key == "likeness" {
			sort = b.SortOrder
		}
	}
	if sort != 4 {
		t.Fatalf("likeness person sort order %d", sort)
	}
}

func TestInstallSeedsFuneralAndPallbearer(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	if err := Install(c); err != nil {
		t.Fatal(err)
	}

	eventType, err := properties.Lookup(c, "event_type", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	funeral, err := propertyterms.Lookup(c, eventType.ID, "funeral", propertyterms.OriginProvenencia)
	if err != nil || funeral.Label != "Funeral" {
		t.Fatalf("funeral %+v %v", funeral, err)
	}

	role, err := properties.Lookup(c, "role", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	pallbearer, err := propertyterms.Lookup(c, role.ID, "pallbearer", propertyterms.OriginProvenencia)
	if err != nil || pallbearer.Label != "Pallbearer" {
		t.Fatalf("pallbearer %+v %v", pallbearer, err)
	}
}

func TestAllTypesPaletteSortFollowsBoardOrder(t *testing.T) {
	want := []string{
		"person", "event", "place", "relationship",
		"participation", "location", "place_relationship", "source",
	}
	types := AllTypes()
	byKey := map[string]int{}
	seen := map[int]string{}
	for _, ty := range types {
		byKey[ty.Key] = ty.PaletteSort
		if other, dup := seen[ty.PaletteSort]; dup {
			t.Errorf("types %q and %q share PaletteSort %d", other, ty.Key, ty.PaletteSort)
		}
		seen[ty.PaletteSort] = ty.Key
	}
	if len(types) != len(want) {
		t.Fatalf("AllTypes() has %d types, want %d", len(types), len(want))
	}
	for i := 1; i < len(want); i++ {
		prev, cur := want[i-1], want[i]
		t.Run(prev+" before "+cur, func(t *testing.T) {
			if byKey[prev] >= byKey[cur] {
				t.Errorf("PaletteSort %s=%d, %s=%d; want %s first", prev, byKey[prev], cur, byKey[cur], prev)
			}
		})
	}
}
