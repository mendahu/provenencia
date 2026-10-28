package autoreconcile

import (
	"slices"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// The name module (conclusion-reconciliation.md §7.2). The pipeline does the
// voting; this file says what a name's comparable units are, when one folds
// into another, and how survivors become one name.
//
//   - Structured parts only. `form` is a transcription and is never read; a
//     name with no parts is no evidence.
//   - Format-agnostic. A part type is only an identifier: a name splits into
//     one unit per part type, and units are compared only with units of the
//     same type. No type behaves differently from another. A type's role in
//     the name format decides only whether an alias table or the prefix rule applies (Fold).
//   - A unit is the type's words, read as value compare reads them: each part
//     normalized (case, punctuation and whitespace ignored; accents are not
//     folded), split into words, and a particle rejoined with the word after
//     it, so Mc Donald = McDonald and a hyphen or a space inside a part is the
//     same as two parts. Two units are one value when value compare keys them
//     the same (word order within a type does not matter). The recorded
//     parts are what's displayed.
//   - Fold: a unit folds into a fuller one when value compare finds the fuller
//     one contains it and every word they pair is the same word, an initial
//     of it, or a given name it starts (never a surname). [J] → [James],
//     [Des] → [Desiree], [James] → [James, Kenneth]. A nickname or
//     abbreviation folds into the formal name it stands for, as the product
//     alias seed lists it (Chris → Christopher, Wm → William, Ebbie → Abel;
//     decided 2026-10-19, S11-28). A name alone never folds, and a pair the
//     list gives both ways folds only as nicknameDirections says. An alias
//     applies only to the part roles its table allows; the product tables
//     cover given names and nicknames, so a surname never alias-folds. An
//     accent fold never folds ([Jose] stays beside [José, Kenneth]).
//   - Majority only outvotes spelling variants of the winner (Robbins beside
//     Robins). A different name (Jake beside James, a married surname) is
//     never outvoted: sources often record a nickname as a given name.
//   - One name. Every displayed record contributes to a single structure;
//     each part type keeps every surviving value, best supported first,
//     each distinct part once (given [James, Jake]). Names are never mixed.
//   - Assemble: each value's parts come from its best-ranked carrier, in the
//     type order of the member with the most types. A member that carries
//     exactly the result is returned as is.

// nameModule reads names under rules.
type nameModule struct {
	rules valuecompare.Rules
}

// nameUnit is one part type's parts of one name.
type nameUnit struct {
	parts []string         // comparison words, in idx order
	raw   []namevalue.Part // the parts as recorded
	words [][]string       // each raw part's comparison words
	cmp   valuecompare.Value
}

// compareValue is the unit as a name for value compare: its raw parts with no
// type, renumbered from 0, and no name pattern, so every word is untyped and
// no part type weighs more than another (S11-06 D-4). fold re-reads the part
// type's role for the alias and prefix checks.
func (u nameUnit) compareValue() valuecompare.Value {
	parts := make([]namevalue.Part, len(u.raw))
	for i, p := range u.raw {
		parts[i] = namevalue.Part{Idx: i, Value: p.Value}
	}
	return valuecompare.Value{Type: propertymeta.ValueTypeName, Name: &namevalue.Value{Parts: parts}}
}

// readName is a name's non-empty parts by type, and the types in the order
// they first appear.
func (m nameModule) readName(n *namevalue.Value) (byType map[string]nameUnit, order []string) {
	byType = map[string]nameUnit{}
	parts := append([]namevalue.Part(nil), n.Parts...)
	sort.SliceStable(parts, func(i, j int) bool { return parts[i].Idx < parts[j].Idx })
	for _, p := range parts {
		words := valuecompare.SplitWords(p.Value, m.rules.Words)
		if len(words) == 0 {
			continue
		}
		typ := strings.TrimSpace(p.Type)
		u, seen := byType[typ]
		if !seen {
			order = append(order, typ)
		}
		u.parts = append(u.parts, words...)
		u.raw = append(u.raw, p)
		u.words = append(u.words, words)
		byType[typ] = u
	}
	return byType, order
}

func (m nameModule) split(v Value) (map[string]unit, bool) {
	byType, _ := m.readName(v.Name)
	units := make(map[string]unit, len(byType))
	for typ, u := range byType {
		u.cmp = u.compareValue()
		key, ok := valuecompare.Key(u.cmp, valuecompare.Same, m.rules)
		if !ok {
			continue
		}
		units[typ] = unit{key: key, data: u}
	}
	if len(units) == 0 {
		return nil, false
	}
	return units, true
}

// fold: value compare finds into contains from, or the two equivalent, and
// every word they pair is the same word, an initial of it, a prefix of it (for
// a part type whose role takes prefixes), or a nickname or abbreviation the
// alias list pairs toward the into word. The type only says which part role an
// alias table or a prefix must allow (compareValue strips it).
func (m nameModule) fold(typ string, from, into unit) bool {
	res, det := valuecompare.Compare(from.data.(nameUnit).cmp, into.data.(nameUnit).cmp, m.rules)
	switch {
	case res.Relation == valuecompare.Contains && det.Broader == valuecompare.SideA:
	case res.Relation == valuecompare.Equivalent:
	default:
		return false
	}
	role := m.rules.Names.Western.Role(typ)
	changed := false
	for _, p := range det.Parts {
		switch {
		case p.Relation == valuecompare.Same:
		case p.Cause == valuecompare.CauseInitial:
			changed = true
		case p.Cause == valuecompare.CausePrefix && valuecompare.PrefixRuleRole(role):
			changed = true
		case (p.Cause == valuecompare.CauseAlias || p.Cause == valuecompare.CauseAbbreviation) &&
			m.rules.Names.Aliases.Expands(p.A.Text, p.B.Text, role):
			changed = true
		default:
			return false
		}
	}
	return changed || res.Relation == valuecompare.Contains
}

// outvotes: only a spelling variant of the winner — value compare pairs the
// words and leaves none unpaired, and each pair is the same word or a
// spelling variant (SpellingSimilarity).
func (m nameModule) outvotes(_ string, winner, other unit) bool {
	_, det := valuecompare.Compare(other.data.(nameUnit).cmp, winner.data.(nameUnit).cmp, m.rules)
	// No pairs means value compare found no match at all.
	if len(det.Parts) == 0 || len(det.OnlyA) != 0 || len(det.OnlyB) != 0 {
		return false
	}
	for _, p := range det.Parts {
		if p.A.Text != p.B.Text &&
			valuecompare.SpellingSimilarity(p.A.Text, p.B.Text, m.rules.Words.FuzzyFloor, m.rules.Words.ShortVariantMinLength) == 0 {
			return false
		}
	}
	return true
}

func (nameModule) oneValue() bool { return true }

func containsAll(have, words []string) bool {
	for _, w := range words {
		if !slices.Contains(have, w) {
			return false
		}
	}
	return true
}

func (m nameModule) assemble(members []Candidate, settled map[string][]unit) Value {
	orders := make([][]string, len(members))
	base := 0
	for i, member := range members {
		_, orders[i] = m.readName(member.Value.Name)
		if len(orders[i]) > len(orders[base]) {
			base = i
		}
	}
	order := append([]string(nil), orders[base]...)
	for _, mo := range orders {
		for i, typ := range mo {
			if slices.Contains(order, typ) {
				continue
			}
			at := 0
			for k := i - 1; k >= 0; k-- {
				if p := slices.Index(order, mo[k]); p >= 0 {
					at = p + 1
					break
				}
			}
			order = slices.Insert(order, at, typ)
		}
	}

	// Each type's parts: every value it settled on, in order, skipping a
	// recorded part whose words are all present already (given [James,
	// Kevin] + [James, Kenneth] → James, Kevin, Kenneth).
	parts := map[string][]string{}
	raw := map[string][]namevalue.Part{}
	for _, typ := range order {
		for _, u := range settled[typ] {
			nu := u.data.(nameUnit)
			for i, words := range nu.words {
				if containsAll(parts[typ], words) {
					continue
				}
				parts[typ] = append(parts[typ], words...)
				raw[typ] = append(raw[typ], nu.raw[i])
			}
		}
	}

	// A member carrying exactly these parts, in this type order, is the value.
	for i, member := range members {
		if !slices.Equal(orders[i], order) {
			continue
		}
		byType, _ := m.readName(member.Value.Name)
		same := true
		for _, typ := range order {
			if !slices.Equal(byType[typ].parts, parts[typ]) {
				same = false
				break
			}
		}
		if same {
			return member.Value
		}
	}

	name := &namevalue.Value{}
	var words []string
	for _, typ := range order {
		for _, p := range raw[typ] {
			name.Parts = append(name.Parts, namevalue.Part{Idx: len(name.Parts), Value: p.Value, Type: p.Type})
			words = append(words, strings.TrimSpace(p.Value))
		}
	}
	name.Form = strings.Join(words, " ")
	return Value{Name: name}
}
