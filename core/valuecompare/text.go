package valuecompare

import (
	"slices"
	"strings"

	"github.com/mendahu/provenencia/core/propertymeta"
)

// text compares free text (toponyms, remarks) word by word, in order.
type text struct{}

func (text) compare(a, b Value, r Rules) (Result, Detail) {
	wa, wb := SplitWords(a.Text, r.Words), SplitWords(b.Text, r.Words)
	if len(wa) == 0 || len(wb) == 0 {
		return Result{Relation: Unknown}, Detail{Cause: CauseEmpty}
	}
	// The same letters in the same order, wherever the words break.
	if ja, jb := strings.Join(wa, ""), strings.Join(wb, ""); ja == jb {
		det := Detail{Cause: CauseCase}
		if a.Text == b.Text {
			det.Cause = CauseExact
		}
		if slices.Equal(wa, wb) {
			det.Parts = positionalParts(wa, wb, nil, nil, r.Words)
		} else {
			det.Parts = []PartMatch{joinedPart(wa, wb, Same, 1, CauseCase)}
		}
		return Result{Relation: Same, Similarity: 1}, det
	}
	// The same letters once accents and letters are folded.
	fa, fb := foldWords(wa, r.Words), foldWords(wb, r.Words)
	if strings.Join(fa, "") == strings.Join(fb, "") {
		det := Detail{Cause: CauseFold}
		if slices.Equal(fa, fb) {
			det.Parts = positionalParts(wa, wb, nil, nil, r.Words)
			return Result{Relation: Equivalent, Similarity: meanSimilarity(det.Parts, textPairWeights(len(det.Parts)), float64(len(wa)+len(wb)))}, det
		}
		det.Parts = []PartMatch{joinedPart(wa, wb, Equivalent, r.Words.FoldCredit, CauseFold)}
		return Result{Relation: Equivalent, Similarity: r.Words.FoldCredit}, det
	}
	// The alignment is quadratic in the words: bound it.
	if len(wa) > r.Text.MaxWords || len(wb) > r.Text.MaxWords {
		return Result{Relation: Unknown}, Detail{Cause: CauseTooLong}
	}
	pairs, onlyA, onlyB, sims := alignWords(wa, wb, fa, fb, r.Words)
	det := Detail{}
	for _, p := range pairs {
		det.Parts = append(det.Parts, PartMatch{
			A: Part{Kind: PartWord, Text: wa[p.i]}, B: Part{Kind: PartWord, Text: wb[p.j]},
			Relation: p.rel, Similarity: p.sim, Cause: p.why,
		})
	}
	det.OnlyA, det.OnlyB = wordParts(wa, nil, onlyA), wordParts(wb, nil, onlyB)
	switch {
	case len(pairs) == 0:
		det.Cause = CauseNoMatch
		return Result{Relation: Different}, det
	case crossesOrder(sims, onlyA, onlyB):
		det.Cause = CauseOrder
		return Result{Relation: Different}, det
	}
	mean := meanSimilarity(det.Parts, textPairWeights(len(det.Parts)), float64(len(wa)+len(wb)))
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

// joinedPart is the one part for two texts whose words break in different
// places: each side's words joined by one space.
func joinedPart(wa, wb []string, rel Relation, sim float64, cause Cause) PartMatch {
	return PartMatch{
		A: Part{Kind: PartWord, Text: strings.Join(wa, " ")}, B: Part{Kind: PartWord, Text: strings.Join(wb, " ")},
		Relation: rel, Similarity: sim, Cause: cause,
	}
}

// TextKey is the key of free text at Same or Equivalent. A text key reads only
// the word rules, never catalog data, so callers outside matching use this
// with the shipped word rules.
func TextKey(s string, level Relation, w WordRules) (string, bool) {
	return Key(Value{Type: propertymeta.ValueTypeText, Text: s}, level, Rules{Words: w})
}

// key: the words joined with no separator, folded at the Equivalent level. They
// are the letters compare tests for Same and Equivalent.
func (text) key(v Value, level Relation, r Rules) (string, bool) {
	words := SplitWords(v.Text, r.Words)
	if len(words) == 0 {
		return "", false
	}
	if level == Equivalent {
		words = foldWords(words, r.Words)
	}
	return strings.Join(words, ""), true
}

func foldWords(words []string, w WordRules) []string {
	out := make([]string, len(words))
	for i, word := range words {
		out[i] = FoldWord(word, w.Folds)
	}
	return out
}

// positionalParts pairs two equal-length word lists position by position: 1
// for a byte-equal pair, FoldCredit for any other. kindsA and kindsB give each
// word's part kind; nil means every word is a PartWord.
func positionalParts(wa, wb []string, kindsA, kindsB []PartKind, w WordRules) []PartMatch {
	parts := make([]PartMatch, len(wa))
	for i := range wa {
		// Not the word table: the initial rule applies only when aligning.
		part := PartMatch{Relation: Same, Similarity: 1, Cause: CauseExact}
		if wa[i] != wb[i] {
			part.Relation, part.Similarity, part.Cause = Equivalent, w.FoldCredit, CauseFold
		}
		part.A, part.B = Part{Kind: kindAt(kindsA, i), Text: wa[i]}, Part{Kind: kindAt(kindsB, i), Text: wb[i]}
		parts[i] = part
	}
	return parts
}

// kindAt is the part kind of word i; nil kinds mean PartWord.
func kindAt(kinds []PartKind, i int) PartKind {
	if kinds == nil {
		return PartWord
	}
	return kinds[i]
}

// wordParts are the words at idx as parts, with their kinds (nil: PartWord).
func wordParts(words []string, kinds []PartKind, idx []int) []Part {
	var out []Part
	for _, i := range idx {
		out = append(out, Part{Kind: kindAt(kinds, i), Text: words[i]})
	}
	return out
}

// textPairWeights are the weights of n text pairs: one word on each side.
func textPairWeights(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = 1 + 1
	}
	return out
}

// meanSimilarity is the weighted mean over every word on both sides: a pair
// earns its similarity times pairWeight[k] (the two words' weights), and total
// is the weight of every word, paired or not. An unpaired word earns 0. Text
// weighs every word 1; names weigh a word by its role.
func meanSimilarity(parts []PartMatch, pairWeight []float64, total float64) float64 {
	var earned float64
	for k, p := range parts {
		earned += p.Similarity * pairWeight[k]
	}
	return earned / total
}

// crossesOrder reports whether an unpaired word of a resembles an unpaired word
// of b: the shared words are in a different order on the two sides. sims is the
// alignment's table.
func crossesOrder(sims [][]wordPair, onlyA, onlyB []int) bool {
	for _, i := range onlyA {
		for _, j := range onlyB {
			if sims[i][j].sim > 0 {
				return true
			}
		}
	}
	return false
}
