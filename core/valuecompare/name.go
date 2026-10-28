package valuecompare

import (
	"math"
	"slices"
	"sort"
	"strings"
)

// name compares two name values by their parts. Each part's role and weight
// come from the name's pattern; any word may pair with any word of the other
// name, discounted when their roles differ. Only Parts are read, never Form.
type name struct{}

// nameWord is one word of a name.
type nameWord struct {
	text, fold string
	role       NameRole
	weight     float64
	first      bool // the first given word in part order
}

func (w nameWord) part() Part { return Part{Kind: PartKind(w.role), Text: w.text} }

// roleOrder is the order words are sorted by, and keys list roles in.
var roleOrder = []NameRole{RoleFamily, RoleGiven, RoleNick, RoleUntyped}

// nameWords reads a name's words under its pattern, sorted by role (family,
// given, nick, untyped) and token, so the order the parts were entered in
// never changes a result. Generation words come back separately, in part
// order. Ignored parts are dropped.
func nameWords(v Value, r Rules) (words []nameWord, generation []string) {
	if v.Name == nil {
		return nil, nil
	}
	pattern, ok := r.Names.Patterns[v.NamePattern]
	if !ok {
		pattern = r.Names.Western
	}
	firstGiven := true
	for _, part := range v.Name.Parts {
		role := pattern.Role(part.Type)
		for _, token := range SplitWords(part.Value, r.Words) {
			switch role {
			case RoleIgnored:
			case RoleGeneration:
				generation = append(generation, token)
			default:
				switch role {
				case RoleFamily, RoleGiven, RoleNick:
				default:
					role = RoleUntyped
				}
				w := nameWord{text: token, fold: FoldWord(token, r.Words.Folds), role: role, weight: pattern.Weight(role, firstGiven)}
				if role == RoleGiven {
					w.first = firstGiven
					firstGiven = false
				}
				words = append(words, w)
			}
		}
	}
	sort.SliceStable(words, func(i, j int) bool {
		if ri, rj := slices.Index(roleOrder, words[i].role), slices.Index(roleOrder, words[j].role); ri != rj {
			return ri < rj
		}
		return words[i].text < words[j].text
	})
	return words, generation
}

// nameKey is the key of a name's words: for each role in the order family,
// given, nick, untyped, generation that has tokens, "role=" and its tokens
// sorted and joined by one space; roles joined by "|". With fold, every token
// but a generation token is folded first.
func nameKey(words []nameWord, generation []string, fold bool) string {
	tokens := map[NameRole][]string{}
	for _, w := range words {
		t := w.text
		if fold {
			t = w.fold
		}
		tokens[w.role] = append(tokens[w.role], t)
	}
	tokens[RoleGeneration] = slices.Clone(generation)
	var parts []string
	for _, role := range append(slices.Clone(roleOrder), RoleGeneration) {
		if list := tokens[role]; len(list) > 0 {
			sort.Strings(list)
			parts = append(parts, string(role)+"="+strings.Join(list, " "))
		}
	}
	return strings.Join(parts, "|")
}

// key: a name with no word has no key. Name keys ignore alias tables: an alias
// joins two words that share a set, and sets do not chain (Chris ~ Christine,
// Chris ~ Christopher, Christine ≁ Christopher), so no key can hold them.
func (name) key(v Value, level Relation, r Rules) (string, bool) {
	words, generation := nameWords(v, r)
	if len(words) == 0 {
		return "", false
	}
	return nameKey(words, generation, level == Equivalent), true
}

// affinity is how far two words' roles agree: 1 for the same role, less when
// the types differ.
func (name) affinity(x, y NameRole, n NameRules) float64 {
	switch {
	case x == y:
		return 1
	case x == RoleUntyped || y == RoleUntyped:
		return n.UntypedAffinity
	case (x == RoleNick && y == RoleGiven) || (x == RoleGiven && y == RoleNick):
		return n.NickAffinity
	}
	return n.CrossRole
}

// contributor is what one pair or leftover says about the relation.
type contributor struct {
	rel     Relation
	why     Cause
	broader Side
}

type wordPairing struct{ a, b int }

// leftover is what one side keeps unpaired: its parts, whether one is a family
// word, and whether one is the first given word.
type leftover struct {
	parts         []Part
	family, first bool
}

// tooManyWords is whether a name has more words, generation words included,
// than the pairing bounds: Compare answers CauseTooLong and wordKeys files no
// prefix keys for it.
func tooManyWords(words []nameWord, generation []string, r Rules) bool {
	return len(words)+len(generation) > r.Names.MaxWords
}

func (n name) compare(a, b Value, r Rules) (Result, Detail) {
	wa, ga := nameWords(a, r)
	wb, gb := nameWords(b, r)
	// (1) Either side empty.
	if len(wa) == 0 || len(wb) == 0 {
		return Result{Relation: Unknown}, Detail{Cause: CauseEmpty}
	}
	keyA, keyB := nameKey(wa, ga, false), nameKey(wb, gb, false)
	// (2) The same words.
	if keyA == keyB {
		return Result{Relation: Same, Similarity: 1}, Detail{Cause: CauseExact, Parts: sameParts(wa, wb, ga, gb, r.Words)}
	}
	// The pairing is exponential in the words: bound it.
	if tooManyWords(wa, ga, r) || tooManyWords(wb, gb, r) {
		return Result{Relation: Unknown}, Detail{Cause: CauseTooLong}
	}
	// (3) Generation words that share none.
	if len(ga) > 0 && len(gb) > 0 && !sharesAny(ga, gb) {
		return Result{Relation: Different}, Detail{Cause: CauseGenerationConflict}
	}
	// (4) Pair. The rows are the side of the smaller key, so the pairing does
	// not depend on which side is A.
	lo, hi, swapped := wa, wb, false
	if keyA > keyB {
		lo, hi, swapped = wb, wa, true
	}
	earned := make([][]float64, len(lo))
	for i, x := range lo {
		earned[i] = make([]float64, len(hi))
		for j, y := range hi {
			sim, _, _, _ := n.word(x, y, r)
			earned[i][j] = sim * n.affinity(x.role, y.role, r.Names) * (x.weight + y.weight)
		}
	}
	_, rowTo := BestAssignment(earned, r.Words.MaxExactPairing)
	var pairing []wordPairing
	usedA, usedB := make([]bool, len(wa)), make([]bool, len(wb))
	for i, j := range rowTo {
		if j < 0 {
			continue
		}
		p := wordPairing{a: i, b: j}
		if swapped {
			p = wordPairing{a: j, b: i}
		}
		pairing = append(pairing, p)
		usedA[p.a], usedB[p.b] = true, true
	}
	// (5) No pair.
	if len(pairing) == 0 {
		return Result{Relation: Different}, Detail{Cause: CauseNoMatch}
	}
	sort.Slice(pairing, func(i, j int) bool { return pairing[i].a < pairing[j].a })

	// Generation words pair with their equal token.
	var genPairs []string
	genUsedA, genUsedB := make([]bool, len(ga)), make([]bool, len(gb))
	for i, t := range ga {
		for j, u := range gb {
			if !genUsedB[j] && t == u {
				genUsedA[i], genUsedB[j] = true, true
				genPairs = append(genPairs, t)
				break
			}
		}
	}

	// (6) Leftovers.
	la, lb := leftoverOf(wa, ga, usedA, genUsedA), leftoverOf(wb, gb, usedB, genUsedB)
	if len(la.parts) > 0 && len(lb.parts) > 0 && ((la.family && lb.family) || (la.first && lb.first)) {
		return Result{Relation: Different}, Detail{Cause: CauseNoMatch}
	}

	// (7) Contributors, and the weight each pair earns from.
	det := Detail{OnlyA: la.parts, OnlyB: lb.parts}
	var contributors []contributor
	var pairWeight []float64
	// Parts list the pairs in A's order; the contributors take the canonical
	// side's order, so the cause does not depend on which side is A.
	pairContributors := make([]contributor, len(pairing))
	for k, p := range pairing {
		x, y := wa[p.a], wb[p.b]
		c, sim := n.pair(x, y, r)
		pairContributors[k] = c
		det.Parts = append(det.Parts, PartMatch{
			A: x.part(), B: y.part(), Relation: c.rel, Similarity: sim, Cause: c.why,
		})
		pairWeight = append(pairWeight, x.weight+y.weight)
	}
	canonical := func(p wordPairing) int {
		if swapped {
			return p.b
		}
		return p.a
	}
	for _, k := range sortedBy(len(pairing), func(i int) int { return canonical(pairing[i]) }) {
		contributors = append(contributors, pairContributors[k])
	}
	for _, t := range genPairs {
		contributors = append(contributors, contributor{rel: Same, why: CauseExact})
		det.Parts = append(det.Parts, PartMatch{
			A: Part{Kind: PartGeneration, Text: t}, B: Part{Kind: PartGeneration, Text: t},
			Relation: Same, Similarity: 1, Cause: CauseExact,
		})
		pairWeight = append(pairWeight, r.Names.Generation+r.Names.Generation)
	}
	switch {
	case len(la.parts) > 0 && len(lb.parts) > 0:
		contributors = append(contributors, contributor{rel: Near, why: CauseOverlap})
	case len(la.parts) > 0:
		contributors = append(contributors, contributor{rel: Contains, why: CauseMissingPart, broader: SideB})
	case len(lb.parts) > 0:
		contributors = append(contributors, contributor{rel: Contains, why: CauseMissingPart, broader: SideA})
	}

	// Weights that sum to nothing give no similarity to take.
	totalWeight := float64(len(ga)+len(gb)) * r.Names.Generation
	for _, w := range wa {
		totalWeight += w.weight
	}
	for _, w := range wb {
		totalWeight += w.weight
	}
	pairedWeight := 0.0
	for _, w := range pairWeight {
		pairedWeight += w
	}
	if totalWeight <= 0 || pairedWeight <= 0 || !finite(totalWeight) || !finite(pairedWeight) {
		return Result{Relation: Unknown}, Detail{Cause: CauseEmpty}
	}

	// (8) The weakest contributor decides the relation, the first at that
	// relation its cause.
	weakest := weakestOf(contributors)
	det.Cause = weakest.why
	if weakest.rel == Contains {
		det.Broader = weakest.broader
	}

	// (9) Similarity: what the pairs earned over the weight of every word.
	similarity := meanSimilarity(det.Parts, pairWeight, totalWeight)

	// (10) A Near name under the minimum is Different.
	if weakest.rel == Near && meanSimilarity(det.Parts, pairWeight, pairedWeight) < r.Names.NearMinimum {
		det.Cause, det.Broader = CauseBelowMinimum, SideNone
		return Result{Relation: Different}, det
	}
	return Result{Relation: weakest.rel, Similarity: similarity}, det
}

// sortedBy is 0…n-1 ordered by key(i).
func sortedBy(n int, key func(int) int) []int {
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool { return key(order[i]) < key(order[j]) })
	return order
}

// pair is what a word pair says about the relation and the similarity it
// earns: a role mismatch is Near, for the role unless the spelling already was.
func (n name) pair(x, y nameWord, r Rules) (contributor, float64) {
	sim, rel, why, broader := n.word(x, y, r)
	c := contributor{rel: rel, why: why}
	switch {
	case x.role != y.role:
		c.rel, c.why = Near, CauseRole
		if rel == Near {
			c.why = CauseSpelling
		}
	case rel == Contains:
		c.broader = broader
	}
	return c, sim * n.affinity(x.role, y.role, r.Names)
}

// leftoverOf is a side's unpaired words, as parts with their kinds, then its
// unpaired generation words.
func leftoverOf(words []nameWord, gen []string, usedWord, usedGen []bool) leftover {
	tokens, kinds := wordTexts(words)
	var idx []int
	var l leftover
	for i, w := range words {
		if !usedWord[i] {
			idx = append(idx, i)
			l.family = l.family || w.role == RoleFamily
			l.first = l.first || w.first
		}
	}
	l.parts = wordParts(tokens, kinds, idx)
	for i, t := range gen {
		if !usedGen[i] {
			l.parts = append(l.parts, Part{Kind: PartGeneration, Text: t})
		}
	}
	return l
}

// weakestOf is the contributor at the weakest relation, the first of those.
// Contains contributors that name both sides as broader are one Near first.
func weakestOf(contributors []contributor) contributor {
	var broaderA, broaderB bool
	for _, c := range contributors {
		if c.rel == Contains {
			broaderA = broaderA || c.broader == SideA
			broaderB = broaderB || c.broader == SideB
		}
	}
	if broaderA && broaderB {
		contributors = slices.DeleteFunc(slices.Clone(contributors), func(c contributor) bool { return c.rel == Contains })
		contributors = append(contributors, contributor{rel: Near, why: CauseOverlap})
	}
	weakest := contributors[0]
	for _, c := range contributors[1:] {
		if slices.Index(relationOrder, c.rel) > slices.Index(relationOrder, weakest.rel) {
			weakest = c
		}
	}
	return weakest
}

// word is the word table for two name words: initials count, the catalog's
// aliases apply, each under its own role, and a word that starts the other
// counts when both take part in the prefix rule.
func (name) word(x, y nameWord, r Rules) (float64, Relation, Cause, Side) {
	credit := 0.0
	if prefixWord(x, r) && prefixWord(y, r) {
		credit = r.Names.PrefixCredit
	}
	return wordSimilarity(x.text, y.text, x.fold, y.fold, r.Words,
		wordOptions{initials: true, aliases: r.Names.Aliases, roleX: x.role, roleY: y.role, prefixCredit: credit})
}

// sameParts pairs two names with the same key: their sorted words position by
// position, then their generation words.
func sameParts(wa, wb []nameWord, ga, gb []string, w WordRules) []PartMatch {
	ta, ka := wordTexts(wa)
	tb, kb := wordTexts(wb)
	parts := positionalParts(ta, tb, ka, kb, w)
	sa, sb := slices.Clone(ga), slices.Clone(gb)
	slices.Sort(sa)
	slices.Sort(sb)
	for i := range sa {
		parts = append(parts, PartMatch{
			A: Part{Kind: PartGeneration, Text: sa[i]}, B: Part{Kind: PartGeneration, Text: sb[i]},
			Relation: Same, Similarity: 1, Cause: CauseExact,
		})
	}
	return parts
}

// wordTexts are the words' tokens and part kinds, in order.
func wordTexts(words []nameWord) ([]string, []PartKind) {
	t, k := make([]string, len(words)), make([]PartKind, len(words))
	for i, word := range words {
		t[i], k[i] = word.text, PartKind(word.role)
	}
	return t, k
}

func sharesAny(a, b []string) bool {
	for _, x := range a {
		if slices.Contains(b, x) {
			return true
		}
	}
	return false
}

// finite reports whether f is neither NaN nor infinite.
func finite(f float64) bool { return !math.IsNaN(f) && !math.IsInf(f, 0) }
