package search

import (
	"strings"
	"testing"
)

func TestJaroWinklerTypoSimilarity(t *testing.T) {
	sim := jaroWinkler("Ilminstr", "Ilminster")
	if sim < FuzzyWeights.JaroWinklerMin {
		t.Fatalf("Ilminstr~Ilminster = %v, want >= %v", sim, FuzzyWeights.JaroWinklerMin)
	}
	sim = jaroWinkler("zzzzzzzz", "Ilminster")
	if sim >= FuzzyWeights.JaroWinklerMin {
		t.Fatalf("garbage should be low, got %v", sim)
	}
}

func TestBuildTrigramORMatch(t *testing.T) {
	m := buildTrigramORMatch("Ilminstr")
	if m == "" || !strings.Contains(m, `"ilm"`) || !strings.Contains(m, `"str"`) {
		t.Fatalf("unexpected match %q", m)
	}
	if buildTrigramORMatch("abc") != "" { // 3 runes < MinTokenRunes 4
		t.Fatal("short token should be empty")
	}
}
