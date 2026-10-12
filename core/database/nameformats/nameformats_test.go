package nameformats

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/match"
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
	if !loaded.Builtin || loaded.Key != match.WesternPattern {
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
	if loaded.Builtin || loaded.Key != match.WesternPattern {
		t.Fatalf("got %+v", loaded)
	}
	if loaded.Profile.Label != "Western" {
		t.Fatalf("label %q", loaded.Profile.Label)
	}
	roles := map[string]bool{}
	for _, pt := range namevalues.PartTypes() {
		if _, ok := loaded.Profile.Pattern.PartRoles[pt.Key]; !ok {
			t.Errorf("part type %q has no role", pt.Key)
		}
		roles[pt.Key] = true
	}
	if len(roles) != len(namevalues.PartTypes()) {
		t.Fatalf("roles %d", len(roles))
	}
	vincent := namevalues.Value{Parts: []namevalues.Part{
		{Idx: 0, Value: "Vincent", Type: namevalues.PartTypeGiven},
		{Idx: 1, Value: "van", Type: namevalues.PartTypeSurnamePrefix},
		{Idx: 2, Value: "Gogh", Type: namevalues.PartTypeSurname},
	}}
	if got := loaded.Profile.Format(vincent, StyleSorted); got != "van Gogh, Vincent" {
		t.Fatalf("sorted %q", got)
	}
	if got := loaded.Profile.Format(vincent, StyleNatural); got != "Vincent van Gogh" {
		t.Fatalf("natural %q", got)
	}
	bare := namevalues.Value{Form: "Vincent van Gogh"}
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
		{0, namevalues.PartTypeGiven, ""},
		{1, namevalues.PartTypeSurname, " — "},
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
	name := namevalues.Value{Form: "unused", Parts: []namevalues.Part{
		{Idx: 0, Value: "Gogh", Type: namevalues.PartTypeSurname},
		{Idx: 1, Value: "Vincent", Type: namevalues.PartTypeGiven},
		{Idx: 2, Value: "painter", Type: namevalues.PartTypeUndetermined},
	}}
	if got := p.Format(name, StyleSorted); got != "Vincent Gogh — painter" {
		t.Fatalf("got %q", got)
	}
}

func nameValue(given, particle, surname string) match.Value {
	return match.Value{Name: &namevalues.Value{Parts: []namevalues.Part{
		{Idx: 0, Value: given, Type: namevalues.PartTypeGiven},
		{Idx: 1, Value: particle, Type: namevalues.PartTypeSurnamePrefix},
		{Idx: 2, Value: surname, Type: namevalues.PartTypeSurname},
	}}}
}
