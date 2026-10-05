package autoreconcile

import (
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/mendahu/provenencia/core/database/namevalues"
)

// The name module (conclusion-reconciliation.md §7.2). The pipeline does the
// voting; this file says what a name's comparable units are, when one folds
// into another, and how survivors become one name.
//
//   - Structured parts only. `form` is a transcription and is never read; a
//     name with no parts is no evidence.
//   - Format-agnostic. A part type is only an identifier: a name splits into
//     one unit per part type, and units are compared only with units of the
//     same type. No type behaves differently from another.
//   - A unit is the type's words in idx order: each part normalized (case,
//     punctuation and whitespace ignored; accents are not folded) and split
//     into words, so a hyphen or a space inside a part is the same as two
//     parts. Smith-Jones = "Smith Jones" = Smith + Jones. Words are compared;
//     the recorded parts are what's displayed.
//   - Fold (subsumption): a unit folds into a fuller one when its parts map,
//     in order, onto a subsequence of the fuller one's, each equal or an
//     initial of it. [J] → [James], [James] → [James, Kenneth].
//   - Majority only outvotes spelling variants of the winner (Robbins beside
//     Robins). A different name (Jake beside James, a married surname) is
//     never outvoted: sources often record a nickname as a given name.
//   - One name. Every displayed record contributes to a single structure;
//     each part type keeps every surviving value, best supported first,
//     each distinct part once (given [James, Jake]). Names are never mixed.
//   - Assemble: each value's parts come from its best-ranked carrier, in the
//     type order of the member with the most types. A member that carries
//     exactly the result is returned as is.

// IsInitial reports whether a word is an initial: a lone cased letter ("J").
// A lone character in an uncased script (蒋, 王) is a whole word.
func IsInitial(word string) bool {
	r := []rune(word)
	return len(r) == 1 && unicode.ToUpper(r[0]) != unicode.ToLower(r[0])
}

type nameModule struct{}

// nameUnit is one part type's parts of one name.
type nameUnit struct {
	parts []string          // normalized words, in idx order
	raw   []namevalues.Part // the parts as recorded
	words [][]string        // each raw part's normalized words
}

// readName is a name's non-empty parts by type, and the types in the order
// they first appear.
func readName(n *namevalues.Value) (byType map[string]nameUnit, order []string) {
	byType = map[string]nameUnit{}
	parts := append([]namevalues.Part(nil), n.Parts...)
	sort.SliceStable(parts, func(i, j int) bool { return parts[i].Idx < parts[j].Idx })
	for _, p := range parts {
		words := strings.Fields(NormalizeForm(p.Value))
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

func (nameModule) split(v Value) (map[string]unit, bool) {
	byType, order := readName(v.Name)
	if len(order) == 0 {
		return nil, false
	}
	units := make(map[string]unit, len(byType))
	for typ, u := range byType {
		units[typ] = unit{key: strings.Join(u.parts, "\x01"), data: u}
	}
	return units, true
}

func (nameModule) fold(_ string, from, into unit) bool {
	return subsumes(into.data.(nameUnit).parts, from.data.(nameUnit).parts)
}

// subsumes reports whether short maps, in order, onto a subsequence of long,
// each part equal to or an initial of its match, and the two differ.
func subsumes(long, short []string) bool {
	if slices.Equal(long, short) || len(short) > len(long) {
		return false
	}
	j := 0
	for _, s := range short {
		for j < len(long) && !partFits(s, long[j]) {
			j++
		}
		if j == len(long) {
			return false
		}
		j++
	}
	return true
}

// partFits: the same part, or an initial of it.
func partFits(short, long string) bool {
	if short == long {
		return true
	}
	if !IsInitial(short) || IsInitial(long) {
		return false
	}
	return []rune(short)[0] == []rune(long)[0]
}

// outvotes: only a spelling variant of the winner — the same number of
// words, each equal or a spelling variant (SpellingSimilarity).
func (nameModule) outvotes(_ string, winner, other unit) bool {
	w, o := winner.data.(nameUnit).parts, other.data.(nameUnit).parts
	if len(w) != len(o) {
		return false
	}
	for i := range w {
		if SpellingSimilarity(w[i], o[i], SpellingFloor, SpellingShortMin) == 0 {
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

func (nameModule) assemble(members []Candidate, settled map[string][]unit) Value {
	orders := make([][]string, len(members))
	base := 0
	for i, m := range members {
		_, orders[i] = readName(m.Value.Name)
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
	raw := map[string][]namevalues.Part{}
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
	for i, m := range members {
		if !slices.Equal(orders[i], order) {
			continue
		}
		byType, _ := readName(m.Value.Name)
		same := true
		for _, typ := range order {
			if !slices.Equal(byType[typ].parts, parts[typ]) {
				same = false
				break
			}
		}
		if same {
			return m.Value
		}
	}

	name := &namevalues.Value{}
	var words []string
	for _, typ := range order {
		for _, p := range raw[typ] {
			name.Parts = append(name.Parts, namevalues.Part{Idx: len(name.Parts), Value: p.Value, Type: p.Type})
			words = append(words, strings.TrimSpace(p.Value))
		}
	}
	name.Form = strings.Join(words, " ")
	return Value{Name: name}
}
