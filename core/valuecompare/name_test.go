package valuecompare_test

import (
	"fmt"
	"maps"
	"math"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// Default word weights under the Western pattern: family F, first given G1,
// other given G2, nick N, untyped U, a generation word Gen. A pair earns
// similarity × affinity × (wa + wb); a similarity is the earned share of
// every word's weight on both sides. Expectations are written as that
// arithmetic.
const (
	F   = 1.5
	G1  = 1.0
	G2  = 0.5
	N   = 0.3
	U   = 1.0
	Gen = 0.5
)

// Edit-distance ratios used below.
const (
	robbins = 6.0 / 7 // robins ~ robbins
	anne    = 3.0 / 4 // ann ~ anne: one added letter in a short word
)

// nv builds a name value from "code=value" parts separated by "|", with a form
// that says nothing, so the parts have to carry the test.
//
//	g given (initials too) · n nick · s surname · sp surname_prefix ·
//	x suffix · p prefix · u undetermined · _ untyped
func nv(spec string) valuecompare.Value {
	n := &namevalue.Value{Form: "(as written)"}
	for i, item := range strings.Split(spec, "|") {
		code, val, ok := strings.Cut(item, "=")
		if !ok {
			panic("bad part " + item)
		}
		typ, known := partCodes[code]
		if !known {
			panic("bad part code " + code)
		}
		n.Parts = append(n.Parts, namevalue.Part{Idx: i, Type: typ, Value: val})
	}
	return valuecompare.Value{Type: propertymeta.ValueTypeName, Name: n}
}

var partCodes = map[string]string{
	"g": namevalue.PartTypeGiven, "n": namevalue.PartTypeNick,
	"s": namevalue.PartTypeSurname, "sp": namevalue.PartTypeSurnamePrefix, "x": namevalue.PartTypeSuffix,
	"p": namevalue.PartTypePrefix, "u": namevalue.PartTypeUndetermined, "_": "",
}

// formOnly is a name with no parts.
func formOnly(form string) valuecompare.Value {
	return valuecompare.Value{Type: propertymeta.ValueTypeName, Name: &namevalue.Value{Form: form}}
}

// patterned is v read under the pattern key.
func patterned(v valuecompare.Value, pattern string) valuecompare.Value {
	v.NamePattern = pattern
	return v
}

var aliasRoles = []valuecompare.NameRole{valuecompare.RoleGiven, valuecompare.RoleNick, valuecompare.RoleUntyped}

var (
	abbreviationTable = valuecompare.AliasTable{Key: "abbreviations", Kind: valuecompare.AliasAbbreviation, Roles: aliasRoles, Sets: [][]string{{"james", "jas"}}}
	nicknameTable     = valuecompare.AliasTable{Key: "nicknames", Kind: valuecompare.AliasNickname, Roles: aliasRoles, Sets: [][]string{{"james", "jim"}, {"bill", "william"}}}
)

// withTables reads the alias tables.
func withTables(tables ...valuecompare.AliasTable) func(*valuecompare.Rules) {
	return func(r *valuecompare.Rules) { r.Names.Aliases = valuecompare.NewAliasIndex(tables) }
}

// flipped is the Western pattern with given and surname trading roles.
func flipped() valuecompare.NamePattern {
	p := matchregistry.Default().Value.Names.Western
	p.PartRoles = maps.Clone(p.PartRoles)
	p.PartRoles[namevalue.PartTypeGiven], p.PartRoles[namevalue.PartTypeSurname] = valuecompare.RoleFamily, valuecompare.RoleGiven
	return p
}

// withoutWeights is the Western pattern with weights that cancel: a catalog
// row can hold any number, and the comparer never divides by their sum.
func withoutWeights(r *valuecompare.Rules) {
	r.Names.Western.Weights = valuecompare.NameWeights{Family: 1, FirstGiven: -2}
}

func withFlipped(r *valuecompare.Rules) {
	r.Names.Patterns = map[string]valuecompare.NamePattern{"flipped": flipped()}
}

func part(kind valuecompare.PartKind, text string) valuecompare.Part {
	return valuecompare.Part{Kind: kind, Text: text}
}

func pair(kind valuecompare.PartKind, a, b string, rel valuecompare.Relation, sim float64, why valuecompare.Cause) valuecompare.PartMatch {
	return valuecompare.PartMatch{A: part(kind, a), B: part(kind, b), Relation: rel, Similarity: sim, Cause: why}
}

type nameCase struct {
	name    string
	rules   func(*valuecompare.Rules)
	a, b    valuecompare.Value
	rel     valuecompare.Relation
	sim     float64
	why     valuecompare.Cause
	broader valuecompare.Side
	onlyA   []valuecompare.Part
	onlyB   []valuecompare.Part
	parts   []valuecompare.PartMatch // checked when set
}

func same(name, a, b string) nameCase {
	return nameCase{name: name, a: nv(a), b: nv(b), rel: valuecompare.Same, sim: 1, why: valuecompare.CauseExact}
}

func differentWhy(name string, a, b valuecompare.Value, why valuecompare.Cause) nameCase {
	return nameCase{name: name, a: a, b: b, rel: valuecompare.Different, why: why}
}

var nameCases = []nameCase{
	// AC-1.1
	same("AC-1.1 identical", "g=James|s=Robins", "g=James|s=Robins"),
	same("AC-1.1 case and punctuation", "g=JAMES|s=Robins.", "g=james|s=robins"),
	same("AC-1.1 surname listed first", "s=Robins|g=James", "g=James|s=Robins"),
	same("AC-1.1 title ignored", "p=Rev.|g=James|s=Robins", "p=Dr.|g=James|s=Robins"),
	same("AC-1.1 surname prefix ignored", "g=Vincent|sp=van|s=Gogh", "g=Vincent|s=Gogh"),
	same("AC-1.1 non-Latin script", "s=蒋|g=浩", "g=浩|s=蒋"),

	// AC-1.2
	{name: "AC-1.2 fold", a: nv("g=José|s=Núñez"), b: nv("g=Jose|s=Nunez"), rel: valuecompare.Equivalent,
		sim: (2*0.95*G1 + 0.95*2*F) / (2*G1 + 2*F), why: valuecompare.CauseFold},

	// AC-1.3
	{name: "AC-1.3 abbreviation", rules: withTables(abbreviationTable), a: nv("g=Jas."), b: nv("g=James"),
		rel: valuecompare.Equivalent, sim: 0.98, why: valuecompare.CauseAbbreviation},
	{name: "AC-1.3 nickname", rules: withTables(nicknameTable), a: nv("g=Jim"), b: nv("g=James"),
		rel: valuecompare.Equivalent, sim: 0.98, why: valuecompare.CauseAlias},

	// AC-1.4
	differentWhy("AC-1.4 no table", nv("g=Jim"), nv("g=James"), valuecompare.CauseNoMatch),
	{name: "AC-1.4 surnames use no given alias", rules: withTables(nicknameTable), a: nv("s=Bill"), b: nv("s=William"),
		rel: valuecompare.Different, why: valuecompare.CauseNoMatch},

	// AC-1.5
	{name: "AC-1.5 surname only", a: nv("s=Robins"), b: nv("g=James|s=Robins"),
		rel: valuecompare.Contains, sim: 2 * F / (2*F + G1), why: valuecompare.CauseMissingPart, broader: valuecompare.SideA,
		onlyB: []valuecompare.Part{part(valuecompare.PartGiven, "james")},
		parts: []valuecompare.PartMatch{pair(valuecompare.PartFamily, "robins", "robins", valuecompare.Same, 1, valuecompare.CauseExact)}},

	// AC-1.6
	{name: "AC-1.6 given only", a: nv("g=James"), b: nv("g=James|s=Robins"),
		rel: valuecompare.Contains, sim: 2 * G1 / (2*G1 + F), why: valuecompare.CauseMissingPart, broader: valuecompare.SideA,
		onlyB: []valuecompare.Part{part(valuecompare.PartFamily, "robins")}},
	{name: "AC-1.6 added middle name", a: nv("g=James|s=Robins"), b: nv("g=James|g=William|s=Robins"),
		rel: valuecompare.Contains, sim: (2*G1 + 2*F) / (2*G1 + 2*F + G2), why: valuecompare.CauseMissingPart, broader: valuecompare.SideA,
		onlyB: []valuecompare.Part{part(valuecompare.PartGiven, "william")}},
	{name: "AC-1.6 generation on one side", a: nv("g=James|s=Robins|x=Jr."), b: nv("g=James|s=Robins"),
		rel: valuecompare.Contains, sim: (2*G1 + 2*F) / (2*G1 + 2*F + Gen), why: valuecompare.CauseMissingPart, broader: valuecompare.SideB,
		onlyA: []valuecompare.Part{part(valuecompare.PartGeneration, "jr")}},

	// AC-1.7
	{name: "AC-1.7 initial", a: nv("g=J.|s=Robins"), b: nv("g=James|s=Robins"),
		rel: valuecompare.Contains, sim: (0.5*2*G1 + 2*F) / (2*G1 + 2*F), why: valuecompare.CauseInitial, broader: valuecompare.SideA,
		parts: []valuecompare.PartMatch{
			pair(valuecompare.PartFamily, "robins", "robins", valuecompare.Same, 1, valuecompare.CauseExact),
			pair(valuecompare.PartGiven, "j", "james", valuecompare.Contains, 0.5, valuecompare.CauseInitial),
		}},

	// AC-1.8
	{name: "AC-1.8 doubled letter", a: nv("g=James|s=Robbins"), b: nv("g=James|s=Robins"),
		rel: valuecompare.Near, sim: (2*G1 + robbins*2*F) / (2*G1 + 2*F), why: valuecompare.CauseSpelling},
	{name: "AC-1.8 short variant", a: nv("g=Ann|s=Robins"), b: nv("g=Anne|s=Robins"),
		rel: valuecompare.Near, sim: (anne*2*G1 + 2*F) / (2*G1 + 2*F), why: valuecompare.CauseSpelling},

	// AC-1.9
	{name: "AC-1.9 typed against untyped", a: nv("g=James|s=Robins"), b: nv("u=James|u=Robins"),
		rel: valuecompare.Near, sim: (0.8*(G1+U) + 0.8*(F+U)) / (G1 + F + 2*U), why: valuecompare.CauseRole,
		parts: []valuecompare.PartMatch{
			{A: part(valuecompare.PartFamily, "robins"), B: part(valuecompare.PartUntyped, "robins"), Relation: valuecompare.Near, Similarity: 0.8, Cause: valuecompare.CauseRole},
			{A: part(valuecompare.PartGiven, "james"), B: part(valuecompare.PartUntyped, "james"), Relation: valuecompare.Near, Similarity: 0.8, Cause: valuecompare.CauseRole},
		}},

	// AC-1.10
	{name: "AC-1.10 different middle names", a: nv("g=James|g=Kenneth|s=Robins"), b: nv("g=James|g=William|s=Robins"),
		rel: valuecompare.Near, sim: (2*G1 + 2*F) / (2*G1 + 2*G2 + 2*F), why: valuecompare.CauseOverlap,
		onlyA: []valuecompare.Part{part(valuecompare.PartGiven, "kenneth")}, onlyB: []valuecompare.Part{part(valuecompare.PartGiven, "william")}},
	{name: "AC-1.10 different middle initials", a: nv("g=James|g=K.|s=Robins"), b: nv("g=James|g=W.|s=Robins"),
		rel: valuecompare.Near, sim: (2*G1 + 2*F) / (2*G1 + 2*G2 + 2*F), why: valuecompare.CauseOverlap,
		onlyA: []valuecompare.Part{part(valuecompare.PartGiven, "k")}, onlyB: []valuecompare.Part{part(valuecompare.PartGiven, "w")}},

	// AC-1.11
	{name: "AC-1.11 short variant, one surname", a: nv("g=Ann"), b: nv("g=Anne|s=Robins"),
		rel: valuecompare.Near, sim: anne * 2 * G1 / (2*G1 + F), why: valuecompare.CauseSpelling,
		onlyB: []valuecompare.Part{part(valuecompare.PartFamily, "robins")}},

	// AC-1.12
	differentWhy("AC-1.12 other first given", nv("g=Mary|s=Robins"), nv("g=James|s=Robins"), valuecompare.CauseNoMatch),
	differentWhy("AC-1.12 other surname", nv("g=Mary|s=Smith"), nv("g=Mary|s=Robins"), valuecompare.CauseNoMatch),
	differentWhy("AC-1.12 compound surnames", nv("g=Maria|s=Silva Costa"), nv("g=Maria|s=Costa Pereira"), valuecompare.CauseNoMatch),
	differentWhy("AC-1.12 other initial", nv("g=W.|s=Robins"), nv("g=James|s=Robins"), valuecompare.CauseNoMatch),
	differentWhy("AC-1.12 no pair", nv("g=Ada|s=Lovelace"), nv("g=James|s=Robins"), valuecompare.CauseNoMatch),

	// AC-1.13
	{name: "AC-1.13 swapped types", a: nv("g=Robins|s=James"), b: nv("g=James|s=Robins"),
		rel: valuecompare.Different, why: valuecompare.CauseBelowMinimum,
		parts: []valuecompare.PartMatch{
			{A: part(valuecompare.PartFamily, "james"), B: part(valuecompare.PartGiven, "james"), Relation: valuecompare.Near, Similarity: 0.5, Cause: valuecompare.CauseRole},
			{A: part(valuecompare.PartGiven, "robins"), B: part(valuecompare.PartFamily, "robins"), Relation: valuecompare.Near, Similarity: 0.5, Cause: valuecompare.CauseRole},
		}},
	{name: "AC-1.13 initial and spelling", a: nv("g=J.|s=Robbins"), b: nv("g=James|s=Robins"),
		rel: valuecompare.Different, why: valuecompare.CauseBelowMinimum,
		parts: []valuecompare.PartMatch{
			pair(valuecompare.PartFamily, "robbins", "robins", valuecompare.Near, robbins, valuecompare.CauseSpelling),
			pair(valuecompare.PartGiven, "j", "james", valuecompare.Contains, 0.5, valuecompare.CauseInitial),
		}},

	// AC-1.14
	differentWhy("AC-1.14 generation conflict", nv("g=James|s=Robins|x=Jr."), nv("g=James|s=Robins|x=Sr."), valuecompare.CauseGenerationConflict),

	// AC-1.15
	{name: "AC-1.15 form only", a: formOnly("James Robins"), b: nv("g=James"), rel: valuecompare.Unknown, why: valuecompare.CauseEmpty},
	{name: "AC-1.15 title and suffix only", a: nv("p=Rev.|x=Jr."), b: nv("g=James"), rel: valuecompare.Unknown, why: valuecompare.CauseEmpty},
	{name: "AC-1.15 no name", a: valuecompare.Value{Type: propertymeta.ValueTypeName}, b: nv("g=James"), rel: valuecompare.Unknown, why: valuecompare.CauseEmpty},

	// AC-1.16
	{name: "AC-1.16 each name under its own pattern", rules: withFlipped,
		a: patterned(nv("g=Robins|s=James"), "flipped"), b: nv("g=James|s=Robins"),
		rel: valuecompare.Same, sim: 1, why: valuecompare.CauseExact},
	{name: "AC-2.1 cause does not depend on the side", a: nv("s=Robins|g=Jon"), b: nv("g=John|_=Robins"),
		rel: valuecompare.Near, sim: (0.8*(F+U) + 0.75*2*G1) / (F + G1 + G1 + U), why: valuecompare.CauseRole},
	{name: "AC-2.1 cause does not depend on the side, swapped", a: nv("g=John|_=Robins"), b: nv("s=Robins|g=Jon"),
		rel: valuecompare.Near, sim: (0.8*(F+U) + 0.75*2*G1) / (F + G1 + G1 + U), why: valuecompare.CauseRole},
	{name: "AC-1.17 weights that sum to nothing are unknown", rules: withoutWeights,
		a: nv("g=Jan|s=Robins"), b: nv("s=Robins"), rel: valuecompare.Unknown, why: valuecompare.CauseEmpty},
	{name: "AC-1.16 unknown key reads Western", rules: withFlipped,
		a: patterned(nv("g=Robins|s=James"), "missing"), b: nv("g=James|s=Robins"),
		rel: valuecompare.Different, why: valuecompare.CauseBelowMinimum},
}

func TestNameCompare(t *testing.T) {
	for _, tc := range nameCases {
		t.Run(tc.name, func(t *testing.T) {
			rules := matchregistry.Default().Value
			if tc.rules != nil {
				tc.rules(&rules)
			}
			res, det := valuecompare.Compare(tc.a, tc.b, rules)
			if res.Relation != tc.rel || math.Abs(res.Similarity-tc.sim) > 1e-9 || det.Cause != tc.why || det.Broader != tc.broader {
				t.Fatalf("got %s %.6f %s broader %q, want %s %.6f %s broader %q",
					res.Relation, res.Similarity, det.Cause, det.Broader, tc.rel, tc.sim, tc.why, tc.broader)
			}
			if !slices.Equal(det.OnlyA, tc.onlyA) || !slices.Equal(det.OnlyB, tc.onlyB) {
				t.Errorf("only A %v B %v, want A %v B %v", det.OnlyA, det.OnlyB, tc.onlyA, tc.onlyB)
			}
			if tc.parts != nil {
				if len(det.Parts) != len(tc.parts) {
					t.Fatalf("parts %+v, want %+v", det.Parts, tc.parts)
				}
				for i, want := range tc.parts {
					got := det.Parts[i]
					if got.A != want.A || got.B != want.B || got.Relation != want.Relation || got.Cause != want.Cause || math.Abs(got.Similarity-want.Similarity) > 1e-9 {
						t.Errorf("part %d: got %+v, want %+v", i, got, want)
					}
				}
			}
			// Swapped sides: same result, sides mirrored.
			back, backDet := valuecompare.Compare(tc.b, tc.a, rules)
			if back.Relation != res.Relation || math.Abs(back.Similarity-res.Similarity) > 1e-9 || backDet.Cause != det.Cause {
				t.Errorf("swapped %s %.6f %s", back.Relation, back.Similarity, backDet.Cause)
			}
		})
	}
}

// A name key agrees with Compare, and never reads alias tables.
func TestNameKey(t *testing.T) {
	rules := matchregistry.Default().Value
	key := func(v valuecompare.Value, level valuecompare.Relation) (string, bool) {
		return valuecompare.Key(v, level, rules)
	}
	if k, ok := key(nv("s=Robins|g=James"), valuecompare.Same); !ok || k != "family=robins|given=james" {
		t.Errorf("surname first: %q %v", k, ok)
	}
	if k, ok := key(nv("g=James|s=Robins"), valuecompare.Same); !ok || k != "family=robins|given=james" {
		t.Errorf("given first: %q %v", k, ok)
	}
	if k, _ := key(nv("g=James|g=William|s=Robins|x=Jr."), valuecompare.Same); k != "family=robins|given=james william|generation=jr" {
		t.Errorf("generation: %q", k)
	}
	ja, jb := nv("g=José"), nv("g=Jose")
	if ka, _ := key(ja, valuecompare.Same); ka == "" {
		t.Error("no key for José")
	} else if kb, _ := key(jb, valuecompare.Same); ka == kb {
		t.Errorf("José and Jose share the Same key %q", ka)
	}
	ka, _ := key(ja, valuecompare.Equivalent)
	kb, _ := key(jb, valuecompare.Equivalent)
	if ka != kb {
		t.Errorf("José and Jose Equivalent keys %q, %q", ka, kb)
	}
	// Jim ~ James are Equivalent with a nickname table, and their keys differ.
	tables := rules
	tables.Names.Aliases = valuecompare.NewAliasIndex([]valuecompare.AliasTable{nicknameTable})
	if res, _ := valuecompare.Compare(nv("g=Jim"), nv("g=James"), tables); res.Relation != valuecompare.Equivalent {
		t.Fatalf("Jim / James with the table: %s", res.Relation)
	}
	ka, _ = valuecompare.Key(nv("g=Jim"), valuecompare.Equivalent, tables)
	kb, _ = valuecompare.Key(nv("g=James"), valuecompare.Equivalent, tables)
	if ka == kb {
		t.Errorf("alias keys equal: %q", ka)
	}
	for _, v := range []valuecompare.Value{formOnly("James Robins"), nv("p=Dr.|x=Jr.")} {
		if k, ok := key(v, valuecompare.Same); ok || k != "" {
			t.Errorf("a name with no word has key %q, %v", k, ok)
		}
	}
	if k, ok := key(nv("g=James"), valuecompare.Near); ok || k != "" {
		t.Errorf("Near has key %q, %v", k, ok)
	}
}

// Ladders assert order rather than exact scores, so they keep their meaning
// when weights are retuned: each name must resemble the probe strictly less
// than the one before, and the last names are Different.
func TestNameLadders(t *testing.T) {
	rules := matchregistry.Default().Value
	ladders := []struct {
		probe     string
		order     []string
		different []string
	}{
		{"g=James|g=K.|s=Robins",
			[]string{"g=James|g=K.|s=Robins", "g=James|g=K.|s=Robbins", "g=J.|g=K.|s=Robins"},
			[]string{"g=Mary|s=Robins", "g=James|g=K.|s=Smith", "g=Ada|s=Lovelace"}},
		{"g=James|s=Robins|x=Jr.",
			[]string{"g=James|s=Robins|x=Jr.", "g=Jim|n=James|s=Robbins"},
			[]string{"g=Mary|s=Robins", "g=James|s=Robins|x=Sr.", "g=Mary|s=Robins|x=Sr."}},
		{"g=Maria|g=Luisa|s=Silva Costa",
			[]string{"g=Maria|g=Luisa|s=Silva|s=Costa", "g=M.|s=Costa"},
			[]string{"g=Luisa|s=Pereira"}},
		// Across formats: the same words typed differently still rank above
		// a different person typed the same way.
		{"g=James|s=Robins",
			[]string{"g=James|s=Robins", "u=James|u=Robins", "u=Mary|u=Robins"},
			[]string{"g=Robins|s=James", "g=Ada|s=Lovelace"}},
	}
	for _, l := range ladders {
		probe := nv(l.probe)
		prev := math.Inf(1)
		for _, spec := range l.order {
			res, _ := valuecompare.Compare(probe, nv(spec), rules)
			if res.Relation == valuecompare.Different || res.Relation == valuecompare.Unknown {
				t.Fatalf("%s vs %s: %s", l.probe, spec, res.Relation)
			}
			if !(res.Similarity < prev) {
				t.Errorf("%s: %s scored %.4f, not below %.4f", l.probe, spec, res.Similarity, prev)
			}
			prev = res.Similarity
		}
		for _, spec := range l.different {
			if res, _ := valuecompare.Compare(probe, nv(spec), rules); res.Relation != valuecompare.Different || res.Similarity != 0 {
				t.Errorf("%s vs %s: %s %.4f, want Different", l.probe, spec, res.Relation, res.Similarity)
			}
		}
	}
	// The scores the ladders rest on.
	for _, tc := range []struct {
		a, b string
		want float64
	}{
		{"g=James|g=K.|s=Robins", "g=James|g=K.|s=Robbins", (2*G1 + 2*G2 + robbins*2*F) / (2*G1 + 2*G2 + 2*F)},
		{"g=James|g=K.|s=Robins", "g=J.|g=K.|s=Robins", (0.5*2*G1 + 2*G2 + 2*F) / (2*G1 + 2*G2 + 2*F)},
		{"g=James|s=Robins|x=Jr.", "g=Jim|n=James|s=Robbins", (0.9*(G1+N) + robbins*2*F) / (G1 + F + Gen + G1 + N + F)},
		{"g=Maria|g=Luisa|s=Silva Costa", "g=M.|s=Costa", (0.5*2*G1 + 2*F) / (G1 + G2 + 2*F + G1 + F)},
		{"g=James|s=Robins", "u=Mary|u=Robins", 0.8 * (F + U) / (G1 + F + 2*U)},
	} {
		if res, _ := valuecompare.Compare(nv(tc.a), nv(tc.b), rules); math.Abs(res.Similarity-tc.want) > 1e-9 {
			t.Errorf("%s / %s: %.6f, want %.6f", tc.a, tc.b, res.Similarity, tc.want)
		}
	}
}

// Generated permutations from fixed seeds (deterministic): every pair keeps
// the comparer's invariants. A failure names its seed and pair; shrink it into
// nameCases.
func TestNameInvariants(t *testing.T) {
	base := matchregistry.Default().Value
	heavy := matchregistry.Default().Value
	heavy.Names.Western.Weights.Family = 3
	heavy.Names.CrossRole = 0.2
	fuzzyOff := matchregistry.Default().Value
	fuzzyOff.Words.FuzzyFloor = 1
	allRules := []valuecompare.Rules{base, heavy, fuzzyOff}

	withForm := func(v valuecompare.Value, form string) valuecompare.Value {
		n := *v.Name
		n.Form = form
		v.Name = &n
		return v
	}
	reparts := func(v valuecompare.Value, parts []namevalue.Part) valuecompare.Value {
		n := *v.Name
		n.Parts = parts
		v.Name = &n
		return v
	}
	addPart := func(v valuecompare.Value, code, value string) valuecompare.Value {
		parts := append(slices.Clone(v.Name.Parts), namevalue.Part{Idx: len(v.Name.Parts), Type: partCodes[code], Value: value})
		return reparts(v, parts)
	}
	hasGeneration := func(v valuecompare.Value) bool {
		return slices.ContainsFunc(v.Name.Parts, func(p namevalue.Part) bool { return p.Type == partCodes["x"] })
	}
	surnamesFirst := func(v valuecompare.Value) valuecompare.Value {
		var front, rest []namevalue.Part
		for _, p := range v.Name.Parts {
			if p.Type == namevalue.PartTypeSurname {
				front = append(front, p)
			} else {
				rest = append(rest, p)
			}
		}
		return reparts(v, append(front, rest...))
	}
	untyped := func(v valuecompare.Value) valuecompare.Value {
		western := matchregistry.Default().Value.Names.Western
		var parts []namevalue.Part
		for _, p := range v.Name.Parts {
			if r := western.Role(p.Type); r != valuecompare.RoleIgnored && r != valuecompare.RoleGeneration {
				p.Type = ""
			}
			parts = append(parts, p)
		}
		return reparts(v, parts)
	}

	for _, seed := range []int64{1, 2, 3, 4} {
		rng := rand.New(rand.NewSource(seed))
		for i := 0; i < 2000; i++ {
			a := randomName(rng)
			b := variantName(rng, a)
			if rng.Intn(2) == 0 {
				b = randomName(rng)
			}
			rules := allRules[rng.Intn(len(allRules))]
			where := fmt.Sprintf("seed %d pair %d: %+v vs %+v", seed, i, *a.Name, *b.Name)

			got, _ := valuecompare.Compare(a, b, rules)
			// The form is never read.
			if again, _ := valuecompare.Compare(withForm(a, "anything else"), withForm(b, ""), rules); again != got {
				t.Fatalf("%s: the form changed a result: %+v vs %+v", where, got, again)
			}
			// Moving surname parts to the front keeps the result.
			if again, _ := valuecompare.Compare(surnamesFirst(a), b, rules); again != got {
				t.Fatalf("%s: part order changed the result: %+v vs %+v", where, got, again)
			}
			// Giving two suffix-less names different suffixes never raises it.
			if !hasGeneration(a) && !hasGeneration(b) {
				if again, _ := valuecompare.Compare(addPart(a, "x", "Jr."), addPart(b, "x", "Sr."), rules); again.Similarity > got.Similarity+1e-9 {
					t.Fatalf("%s: a generation conflict raised the similarity: %.4f vs %.4f", where, got.Similarity, again.Similarity)
				}
			}
			// A title on one side and a surname prefix on the other never move it.
			if again, _ := valuecompare.Compare(addPart(a, "p", "Rev."), addPart(b, "sp", "van"), rules); again != got {
				t.Fatalf("%s: an ignored part moved the result: %+v vs %+v", where, got, again)
			}
			// The same words with no types still connect, never above 1.
			if res, _ := valuecompare.Compare(a, untyped(a), rules); res.Relation != valuecompare.Unknown {
				if (res.Relation != valuecompare.Same && res.Relation != valuecompare.Near) || res.Similarity <= 0 || res.Similarity > 1 {
					t.Fatalf("%s: retyped copy %+v", where, res)
				}
			}
		}
	}
}
