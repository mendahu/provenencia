package match

import (
	"math"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/namevalue"
)

// WordCounts is how common each name word is among a project's filed people:
// People is how many handles carry a counted word, Words how many handles
// carry each comparison token. A handle counts once per word however many
// name values spell it. The zero value is an empty project: no word is
// common.
type WordCounts struct {
	People int
	Words  map[string]int
}

// noWordCounts is the comparer default: no counts, so no discount.
var noWordCounts WordCounts

// CountWords counts the words of each handle's name values. One entry of
// people is one handle. Every part word counts, whatever its role; initials
// do not. Keys are comparison tokens: split with the comparer's particles,
// then folded with its folds, so José and Jose share one key. Lookup in
// discount reads the same keys.
func (c NameComparer) CountWords(people [][]namevalue.Value) WordCounts {
	r := c.Words.resolve(false)
	out := WordCounts{Words: map[string]int{}}
	for _, handle := range people {
		seen := map[string]bool{}
		for i := range handle {
			for _, part := range handle[i].Parts {
				for _, w := range splitWords(part.Value, r.particles) {
					if autoreconcile.IsInitial(w) {
						continue
					}
					seen[foldWord(w, r.folds)] = true
				}
			}
		}
		if len(seen) == 0 {
			continue
		}
		out.People++
		for k := range seen {
			out.Words[k]++
		}
	}
	return out
}

// wordDiscount is the share of points one matched word costs:
// FrequencyDiscount × ln(df) / ln(pop). df is the word's handle count, capped
// at pop; pop is the larger of the project's People, FrequencyMinPeople, and
// 2. An initial and a word on one handle or fewer cost nothing.
func (c NameComparer) wordDiscount(text string, r wordRules) float64 {
	if autoreconcile.IsInitial(text) {
		return 0
	}
	counts := setting(c.Frequency, *DefaultNames.Frequency)
	df := counts.Words[foldWord(text, r.folds)]
	if df <= 1 {
		return 0
	}
	pop := max(counts.People, setting(c.FrequencyMinPeople, *DefaultNames.FrequencyMinPeople), 2)
	df = min(df, pop)
	return setting(c.FrequencyDiscount, *DefaultNames.FrequencyDiscount) * math.Log(float64(df)) / math.Log(float64(pop))
}

// discount is the role-weighted mean of the word discounts of the words the
// best assignment paired, on both sides. With no paired word it is 0.
func (c NameComparer) discount(wa []nameWord, usedA []bool, wb []nameWord, usedB []bool, r wordRules) float64 {
	counts := setting(c.Frequency, *DefaultNames.Frequency)
	if len(counts.Words) == 0 {
		return 0
	}
	var sum, weight float64
	add := func(words []nameWord, used []bool) {
		for i, w := range words {
			if i < len(used) && used[i] {
				sum += w.weight * c.wordDiscount(w.text, r)
				weight += w.weight
			}
		}
	}
	add(wa, usedA)
	add(wb, usedB)
	if weight == 0 {
		return 0
	}
	return sum / weight
}
