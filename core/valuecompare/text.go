package valuecompare

import (
	"slices"
	"strings"
)

// text compares free text (toponyms, remarks) word by word, in order.
type text struct{}

func particleSet(w WordRules) map[string]bool {
	set := make(map[string]bool, len(w.Particles))
	for _, p := range w.Particles {
		set[p] = true
	}
	return set
}

func (text) compare(a, b Value, r Rules) (Result, Detail) {
	particles := particleSet(r.Words)
	wa, wb := SplitWords(a.Text, particles), SplitWords(b.Text, particles)
	if len(wa) == 0 || len(wb) == 0 {
		return Result{Relation: Unknown}, Detail{Cause: CauseEmpty}
	}
	// The same words in the same order.
	if slices.Equal(wa, wb) {
		det := Detail{Cause: CauseCase}
		if a.Text == b.Text {
			det.Cause = CauseExact
		}
		det.Parts = positionalParts(wa, wb, r.Words)
		return Result{Relation: Same, Similarity: 1}, det
	}
	// The same words in the same order once accents and letters are folded.
	if slices.Equal(foldWords(wa, r.Words), foldWords(wb, r.Words)) {
		det := Detail{Cause: CauseFold, Parts: positionalParts(wa, wb, r.Words)}
		return Result{Relation: Equivalent, Similarity: meanSimilarity(det.Parts, len(wa), len(wb))}, det
	}
	pairs, onlyA, onlyB := alignWords(wa, wb, r.Words)
	det := Detail{}
	for _, p := range pairs {
		det.Parts = append(det.Parts, PartMatch{
			A: Part{Kind: PartWord, Text: wa[p.i]}, B: Part{Kind: PartWord, Text: wb[p.j]},
			Relation: p.rel, Similarity: p.sim, Cause: p.why,
		})
	}
	det.OnlyA, det.OnlyB = wordParts(wa, onlyA), wordParts(wb, onlyB)
	switch {
	case len(pairs) == 0:
		det.Cause = CauseNoMatch
		return Result{Relation: Different}, det
	case crossesOrder(wa, wb, onlyA, onlyB, r.Words):
		det.Cause = CauseOrder
		return Result{Relation: Different}, det
	}
	mean := meanSimilarity(det.Parts, len(wa), len(wb))
	switch {
	case len(onlyA) > 0 && len(onlyB) > 0:
		det.Cause = CauseOverlap
		return Result{Relation: Overlaps, Similarity: mean * r.Text.Partial}, det
	case len(onlyA) > 0 || len(onlyB) > 0:
		det.Cause = CauseMissingPart
		det.Broader = SideA
		if len(onlyA) > 0 {
			det.Broader = SideB
		}
		return Result{Relation: Contains, Similarity: mean * r.Text.Partial}, det
	}
	det.Cause = CauseSpelling
	return Result{Relation: Near, Similarity: mean}, det
}

// key: the words joined by one space, folded at the Equivalent level. They are
// the sequences compare tests for Same and Equivalent.
func (text) key(v Value, level Relation, r Rules) (string, bool) {
	words := SplitWords(v.Text, particleSet(r.Words))
	if len(words) == 0 {
		return "", false
	}
	if level == Equivalent {
		words = foldWords(words, r.Words)
	}
	return strings.Join(words, " "), true
}

func foldWords(words []string, w WordRules) []string {
	out := make([]string, len(words))
	for i, word := range words {
		out[i] = FoldWord(word, w.Folds)
	}
	return out
}

// positionalParts pairs two equal-length word lists position by position: 1
// for a byte-equal pair, FoldCredit for any other.
func positionalParts(wa, wb []string, w WordRules) []PartMatch {
	parts := make([]PartMatch, len(wa))
	for i := range wa {
		// Not the word table: the initial rule applies only when aligning.
		part := PartMatch{Relation: Same, Similarity: 1, Cause: CauseExact}
		if wa[i] != wb[i] {
			part.Relation, part.Similarity, part.Cause = Equivalent, w.FoldCredit, CauseFold
		}
		part.A, part.B = Part{Kind: PartWord, Text: wa[i]}, Part{Kind: PartWord, Text: wb[i]}
		parts[i] = part
	}
	return parts
}

func wordParts(words []string, idx []int) []Part {
	var out []Part
	for _, i := range idx {
		out = append(out, Part{Kind: PartWord, Text: words[i]})
	}
	return out
}

// meanSimilarity is the mean over every word on both sides: a word scores its
// pair's similarity, an unpaired word scores 0.
func meanSimilarity(parts []PartMatch, na, nb int) float64 {
	var total float64
	for _, p := range parts {
		total += p.Similarity
		total += p.Similarity
	}
	return total / float64(na+nb)
}

// crossesOrder reports whether an unpaired word of a resembles an unpaired word
// of b: the shared words are in a different order on the two sides.
func crossesOrder(wa, wb []string, onlyA, onlyB []int, w WordRules) bool {
	for _, i := range onlyA {
		for _, j := range onlyB {
			if sim, _, _ := wordSimilarity(wa[i], wb[j], w); sim > 0 {
				return true
			}
		}
	}
	return false
}
