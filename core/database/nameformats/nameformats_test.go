package nameformats

import (
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func TestMissingDefaultUsesWestern(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(db)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.Builtin || loaded.Key != valuecompare.WesternPattern {
		t.Fatalf("got builtin=%v key=%q", loaded.Builtin, loaded.Key)
	}
	a := nameValue("Vincent", "van", "Gogh")
	b := nameValue("Vincent", "van", "Gogh")
	a.NamePattern = loaded.Key
	b.NamePattern = loaded.Key
	got, ok := loaded.Comparer().Compare(a, b)
	want, wantOK := match.DefaultNames.Compare(a, b)
	if ok != wantOK || got != want {
		t.Fatalf("score %v %v, built-in %v %v", got, ok, want, wantOK)
	}
}

func TestInstallLoadsWesternAndFormats(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	if err := Install(c); err != nil {
		t.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(db)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Builtin || loaded.Key != valuecompare.WesternPattern {
		t.Fatalf("got %+v", loaded)
	}
	if loaded.Profile.Label != "Western" {
		t.Fatalf("label %q", loaded.Profile.Label)
	}
	roles := map[string]bool{}
	for _, pt := range namevalue.PartTypes() {
		if _, ok := loaded.Profile.Pattern.PartRoles[pt.Key]; !ok {
			t.Errorf("part type %q has no role", pt.Key)
		}
		roles[pt.Key] = true
	}
	if len(roles) != len(namevalue.PartTypes()) {
		t.Fatalf("roles %d", len(roles))
	}
	vincent := namevalue.Value{Parts: []namevalue.Part{
		{Idx: 0, Value: "Vincent", Type: namevalue.PartTypeGiven},
		{Idx: 1, Value: "van", Type: namevalue.PartTypeSurnamePrefix},
		{Idx: 2, Value: "Gogh", Type: namevalue.PartTypeSurname},
	}}
	if got := loaded.Profile.Format(vincent, StyleSorted); got != "van Gogh, Vincent" {
		t.Fatalf("sorted %q", got)
	}
	if got := loaded.Profile.Format(vincent, StyleNatural); got != "Vincent van Gogh" {
		t.Fatalf("natural %q", got)
	}
	bare := namevalue.Value{Form: "Vincent van Gogh"}
	if got := loaded.Profile.Format(bare, StyleSorted); got != "Vincent van Gogh" {
		t.Fatalf("form %q", got)
	}
	rewritten := vincent
	rewritten.Form = "do not use"
	if got := loaded.Profile.Format(rewritten, StyleNatural); got != "Vincent van Gogh" {
		t.Fatalf("rewrote form: %q", got)
	}
}

func TestRowOnlyProfileFormatsWithoutCode(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO name_format_profiles (id, key, origin, label) VALUES (?, 'given-first', 'user', 'Given first')`, id[:]); err != nil {
		t.Fatal(err)
	}
	parts := []struct {
		idx   int
		typ   string
		trail string
	}{
		{0, namevalue.PartTypeGiven, ""},
		{1, namevalue.PartTypeSurname, " — "},
	}
	for _, row := range parts {
		if _, err := db.Exec(`INSERT INTO name_format_profile_parts
			(profile_id, style, idx, part_type, trailing_separator) VALUES (?, 'sorted', ?, ?, ?)`,
			id[:], row.idx, row.typ, row.trail); err != nil {
			t.Fatal(err)
		}
	}
	p, err := LoadID(db, id[:])
	if err != nil {
		t.Fatal(err)
	}
	name := namevalue.Value{Form: "unused", Parts: []namevalue.Part{
		{Idx: 0, Value: "Gogh", Type: namevalue.PartTypeSurname},
		{Idx: 1, Value: "Vincent", Type: namevalue.PartTypeGiven},
		{Idx: 2, Value: "painter", Type: namevalue.PartTypeUndetermined},
	}}
	if got := p.Format(name, StyleSorted); got != "Vincent Gogh — painter" {
		t.Fatalf("got %q", got)
	}
}

func TestLoadIDRefusesBadWeight(t *testing.T) {
	for name, weight := range map[string]float64{"negative": -1, "infinite": math.Inf(1), "huge": 1e308} {
		t.Run(name, func(t *testing.T) {
			c, err := database.Create(t.TempDir(), "t.provenencia")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = c.Close() })
			db, err := c.DB()
			if err != nil {
				t.Fatal(err)
			}
			id, err := uuid.NewV7()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`INSERT INTO name_format_profiles (id, key, origin, label) VALUES (?, 'bad', 'user', 'Bad')`, id[:]); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`INSERT INTO name_format_profile_weights (profile_id, slot, weight) VALUES (?, 'family', ?)`, id[:], weight); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadID(db, id[:]); err == nil {
				t.Fatalf("weight %v loaded", weight)
			}
		})
	}
}

func nameValue(given, particle, surname string) match.Value {
	return match.Value{Name: &namevalue.Value{Parts: []namevalue.Part{
		{Idx: 0, Value: given, Type: namevalue.PartTypeGiven},
		{Idx: 1, Value: particle, Type: namevalue.PartTypeSurnamePrefix},
		{Idx: 2, Value: surname, Type: namevalue.PartTypeSurname},
	}}}
}

// The seeded Western format and the registry's Western pattern are one
// pattern: same roles per product part type, same weights.
func TestSeedMatchesRegistryWestern(t *testing.T) {
	want := matchregistry.Default().Value.Names.Western
	var seed *seedProfile
	for i := range seedProfiles {
		if seedProfiles[i].Key == valuecompare.WesternPattern {
			seed = &seedProfiles[i]
		}
	}
	if seed == nil {
		t.Fatal("no western seed profile")
	}
	for _, pt := range namevalue.PartTypes() {
		var got valuecompare.NameRole
		for _, r := range seed.Roles {
			if r.PartType == pt.Key {
				got = r.Role
			}
		}
		if got != want.Role(pt.Key) {
			t.Errorf("part type %q: seed role %q, registry %q", pt.Key, got, want.Role(pt.Key))
		}
	}
	weights := map[string]float64{
		"family":      want.Weights.Family,
		"first_given": want.Weights.FirstGiven,
		"other_given": want.Weights.OtherGiven,
		"nick":        want.Weights.Nick,
		"untyped":     want.Weights.Untyped,
	}
	if len(seed.Weights) != len(weights) {
		t.Errorf("seed has %d weights, registry %d", len(seed.Weights), len(weights))
	}
	for _, w := range seed.Weights {
		if got, ok := weights[w.Slot]; !ok || got != w.Weight {
			t.Errorf("weight %q: seed %v, registry %v (%v)", w.Slot, w.Weight, got, ok)
		}
	}
}
