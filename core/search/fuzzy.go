package search

import (
	"math"
	"strings"
	"unicode/utf8"
)

// jaroWinkler returns similarity in [0, 1] over Unicode runes (case-folded).
func jaroWinkler(a, b string) float64 {
	ar := []rune(strings.ToLower(strings.TrimSpace(a)))
	br := []rune(strings.ToLower(strings.TrimSpace(b)))
	if len(ar) == 0 && len(br) == 0 {
		return 1
	}
	if len(ar) == 0 || len(br) == 0 {
		return 0
	}
	j := jaro(ar, br)
	if j < 0.7 {
		return j
	}
	// Common prefix bonus (Winkler), capped at 4 runes.
	prefix := 0
	limit := len(ar)
	if len(br) < limit {
		limit = len(br)
	}
	if limit > 4 {
		limit = 4
	}
	for prefix < limit && ar[prefix] == br[prefix] {
		prefix++
	}
	return j + float64(prefix)*0.1*(1-j)
}

func jaro(ar, br []rune) float64 {
	if len(ar) == 0 && len(br) == 0 {
		return 1
	}
	if len(ar) == 0 || len(br) == 0 {
		return 0
	}
	matchDist := int(math.Max(float64(len(ar)), float64(len(br)))/2) - 1
	if matchDist < 0 {
		matchDist = 0
	}
	aMatches := make([]bool, len(ar))
	bMatches := make([]bool, len(br))
	matches := 0
	for i := range ar {
		start := i - matchDist
		if start < 0 {
			start = 0
		}
		end := i + matchDist + 1
		if end > len(br) {
			end = len(br)
		}
		for j := start; j < end; j++ {
			if bMatches[j] || ar[i] != br[j] {
				continue
			}
			aMatches[i] = true
			bMatches[j] = true
			matches++
			break
		}
	}
	if matches == 0 {
		return 0
	}
	transpositions := 0
	k := 0
	for i := range ar {
		if !aMatches[i] {
			continue
		}
		for !bMatches[k] {
			k++
		}
		if ar[i] != br[k] {
			transpositions++
		}
		k++
	}
	m := float64(matches)
	return (m/float64(len(ar)) + m/float64(len(br)) + (m-float64(transpositions)/2)/m) / 3
}

// bestFuzzyFieldScore returns the best Jaro–Winkler over high-weight identity
// fields (title/label/ref/key), not body text.
func bestFuzzyFieldScore(values map[string]string, tokens []string) (score float64, reason string) {
	fields := []string{"title", "label", "ref", "key"}
	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		if utf8.RuneCountInString(tok) < FuzzyWeights.MinTokenRunes {
			continue
		}
		for _, name := range fields {
			v := values[name]
			if v == "" {
				continue
			}
			// Compare against whole field and whitespace-separated words.
			candidates := []string{v}
			candidates = append(candidates, strings.Fields(v)...)
			for _, cand := range candidates {
				if utf8.RuneCountInString(cand) < 3 {
					continue
				}
				sim := jaroWinkler(tok, cand)
				if sim > score {
					score = sim
					reason = name
				}
			}
		}
	}
	return score, reason
}

// buildTrigramORMatch builds an FTS5 MATCH of OR-quoted character trigrams.
// Empty if the token is too short after sanitization.
func buildTrigramORMatch(token string) string {
	token = sanitizeFTSToken(token)
	if token == "" || utf8.RuneCountInString(token) < FuzzyWeights.MinTokenRunes {
		return ""
	}
	runes := []rune(token)
	parts := make([]string, 0, len(runes)-2)
	seen := make(map[string]struct{}, len(runes))
	for i := 0; i+2 < len(runes); i++ {
		tri := string(runes[i : i+3])
		if _, ok := seen[tri]; ok {
			continue
		}
		seen[tri] = struct{}{}
		// Quote so FTS5 treats the trigram as a phrase token.
		parts = append(parts, `"`+tri+`"`)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " OR ")
}
