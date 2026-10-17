package match

import (
	"math"
	"testing"

	"github.com/mendahu/provenencia/core/namevalue"
)

// counts100 are the counts of AC-1.1: 100 people; smith on 40, quirk on 1,
// john on 10.
func counts100() *WordCounts {
	return &WordCounts{People: 100, Words: map[string]int{"smith": 40, "quirk": 1, "john": 10}}
}

func TestCountNameWords(t *testing.T) {
	val := func(form string) namevalue.Value { return *name(form).Name }
	t.Run("one count per handle per word, folded, no initials", func(t *testing.T) {
		got := NameComparer{}.CountWords([][]namevalue.Value{
			{val("José Smith"), val("Smith")},
			{val("Jose Smith"), val("J.")},
			{val("J.")},
		})
		if got.People != 2 || got.Words["smith"] != 2 || got.Words["jose"] != 2 {
			t.Fatalf("got %+v", got)
		}
		if _, ok := got.Words["j"]; ok {
			t.Fatalf("an initial was counted: %+v", got)
		}
	})
	t.Run("a comparer with its own folds keys by them", func(t *testing.T) {
		c := NameComparer{Words: WordRules{Folds: Set(map[rune]string{'ø': "oe"})}}
		got := c.CountWords([][]namevalue.Value{{val("Søren Smith")}})
		if got.Words["soeren"] != 1 {
			t.Fatalf("got %+v", got)
		}
	})
}

func TestNameDiscount(t *testing.T) {
	tests := []struct {
		name   string
		counts *WordCounts
		a, b   string
		want   float64
	}{
		{"rare surname", counts100(), "John Quirk", "John Quirk", 0.040},
		{"common surname", counts100(), "John Smith", "John Smith", 0.136},
		{"small project is floored", &WordCounts{People: 8, Words: map[string]int{"robins": 8}}, "Robins", "Robins", 0.090},
		{"empty project", &WordCounts{}, "John Smith", "John Smith", 0},
		{"no counts at all", nil, "John Smith", "John Smith", 0},
		{"initial pays nothing, the full word pays its own", &WordCounts{People: 100, Words: map[string]int{"robins": 40, "james": 10}}, "J. Robins", "James Robins", 0.116},
		{"no paired word", counts100(), "John Smith", "Mary Jones", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NameComparer{}
			if tt.counts != nil {
				c.Frequency = tt.counts
			}
			_, got, ok := c.CompareDiscount(name(tt.a), name(tt.b))
			if !ok || math.Abs(got-tt.want) > 0.0005 {
				t.Fatalf("discount %.4f (%v), want %.3f", got, ok, tt.want)
			}
		})
	}
	t.Run("similarity is unchanged", func(t *testing.T) {
		s, _, _ := NameComparer{Frequency: counts100()}.CompareDiscount(name("John Smith"), name("John Smith"))
		if s != 1 {
			t.Fatalf("similarity %v", s)
		}
	})
}
