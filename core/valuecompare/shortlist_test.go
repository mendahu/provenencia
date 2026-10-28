package valuecompare_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func billRules() valuecompare.Rules {
	rules := matchregistry.Default().Value
	bill := valuecompare.AliasTable{
		Key: "bill", Kind: valuecompare.AliasNickname,
		Roles: []valuecompare.NameRole{valuecompare.RoleGiven},
		Sets:  [][]string{{"bill", "william"}},
	}
	rules.Names.Aliases = valuecompare.NewAliasIndex([]valuecompare.AliasTable{bill}, rules.Words)
	return rules
}

func slName(form string) valuecompare.Value {
	return valuecompare.Value{Type: propertymeta.ValueTypeName, Name: namevaluetest.Western(form), NamePattern: valuecompare.WesternPattern}
}

func slGiven(form string) valuecompare.Value {
	return valuecompare.Value{Type: propertymeta.ValueTypeName, Name: namevaluetest.Given(form), NamePattern: valuecompare.WesternPattern}
}

func slNick(form string) valuecompare.Value {
	return valuecompare.Value{
		Type:        propertymeta.ValueTypeName,
		Name:        &namevalue.Value{Form: form, Parts: []namevalue.Part{{Idx: 0, Value: form, Type: namevalue.PartTypeNick}}},
		NamePattern: valuecompare.WesternPattern,
	}
}

func slGeneration() valuecompare.Value {
	v := slName("John Smith")
	v.Name.Parts = append(v.Name.Parts, namevalue.Part{Idx: 2, Value: "Jr", Type: namevalue.PartTypeSuffix})
	return v
}

func slText(s string) valuecompare.Value {
	return valuecompare.Value{Type: propertymeta.ValueTypeText, Text: s}
}

func slDate(d *datevalue.Value) valuecompare.Value {
	return valuecompare.Value{Type: propertymeta.ValueTypeDate, Date: d}
}

// prefixKey is whether k is a prefix key, the only kind the two sides do not share.
func prefixKey(k string) bool {
	return strings.HasPrefix(k, "p:") || strings.HasPrefix(k, "g:")
}

// plain drops the prefix keys.
func plain(keys []string) []string {
	var out []string
	for _, k := range keys {
		if !prefixKey(k) {
			out = append(out, k)
		}
	}
	return out
}

func TestWordKeys(t *testing.T) {
	rules := billRules()
	billKey := rules.Names.Aliases.SetKeys("bill")[0]
	whole := func(v valuecompare.Value) string {
		k, ok := valuecompare.Key(v, valuecompare.Equivalent, rules)
		if !ok {
			t.Fatalf("no key for %+v", v)
		}
		return "=" + k
	}
	tests := []struct {
		name     string
		v        valuecompare.Value
		want     []string // exact when set
		has      []string
		not      []string
		noPrefix bool // neither list holds a prefix key
	}{
		{name: "initial is no key", v: slName("J. Robins"), want: []string{whole(slName("J. Robins")), "w:robins"}, noPrefix: true},
		{name: "lone initial keys only its whole key", v: slGiven("J."), want: []string{whole(slGiven("J."))}, noPrefix: true},
		{name: "folded word", v: slName("José Robins"), has: []string{"w:jose", "w:robins"}},
		{name: "generation word", v: slGeneration(), has: []string{"w:john", "w:smith"}, not: []string{"w:jr"}},
		{name: "alias set", v: slName("Bill"), has: []string{"a:" + billKey}},
		{name: "text words", v: slText("New York"), has: []string{"w:new", "w:york"}, not: []string{"a:" + billKey}, noPrefix: true},
		{name: "integer", v: valuecompare.Value{Type: propertymeta.ValueTypeInteger, Integer: 4}},
		{name: "empty name", v: valuecompare.Value{Type: propertymeta.ValueTypeName, NamePattern: valuecompare.WesternPattern}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, lookup := valuecompare.WordKeys(tt.v, rules)
			if tt.noPrefix {
				for _, k := range append(slices.Clone(index), lookup...) {
					if prefixKey(k) {
						t.Errorf("keys %q %q hold a prefix key %q", index, lookup, k)
					}
				}
			}
			got := plain(index)
			if !slices.Equal(got, plain(lookup)) {
				t.Fatalf("plain index %q != plain lookup %q", got, plain(lookup))
			}
			if tt.want != nil {
				want := slices.Clone(tt.want)
				slices.Sort(want)
				if !slices.Equal(got, want) {
					t.Fatalf("WordKeys = %q, want %q", got, want)
				}
			}
			if tt.want == nil && tt.has == nil && tt.not == nil && (index != nil || lookup != nil) {
				t.Fatalf("WordKeys = %q, %q, want nil", index, lookup)
			}
			for _, k := range tt.has {
				if !slices.Contains(got, k) {
					t.Errorf("WordKeys %q lacks %q", got, k)
				}
			}
			for _, k := range tt.not {
				if slices.Contains(got, k) {
					t.Errorf("WordKeys %q holds %q", got, k)
				}
			}
			for _, list := range [][]string{index, lookup} {
				if !slices.IsSorted(list) || len(slices.Compact(slices.Clone(list))) != len(list) {
					t.Errorf("WordKeys %q is not sorted and unique", list)
				}
			}
		})
	}
	if index, _ := valuecompare.WordKeys(slText("New York"), rules); slices.ContainsFunc(index, func(k string) bool { return strings.HasPrefix(k, "a:") }) {
		t.Errorf("text keys %q hold an alias key", index)
	}
}

func TestWordKeysReachPrefixes(t *testing.T) {
	tests := []struct {
		name         string
		filed, probe valuecompare.Value
		min          int
		meets        bool
	}{
		{"longer filed, shorter probed", slGiven("Desiree"), slGiven("Des"), 2, true},
		{"shorter filed, longer probed", slGiven("Des"), slGiven("Desiree"), 2, true},
		{"not a prefix", slGiven("Desiree"), slGiven("Desmond"), 2, false},
		{"Albert and Al", slGiven("Albert"), slGiven("Al"), 2, true},
		{"word of the minimum length", slGiven("De"), slGiven("Desiree"), 2, true},
		{"under the minimum", slGiven("Desiree"), slGiven("De"), 3, false},
		{"accent folds", slGiven("Desiree"), slGiven("Dés"), 2, true},
		{"surname filed", slName("Mary MacDonald"), slName("John Mac"), 2, false},
		{"surname filed, given probed", slName("Mary MacDonald"), slGiven("Mac"), 2, false},
		{"nickname filed", slNick("Des"), slGiven("Desiree"), 2, true},
		{"prefix keys off", slGiven("Desiree"), slGiven("Des"), 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := matchregistry.Default().Value
			rules.Names.PrefixMinLength = tt.min
			index, _ := valuecompare.WordKeys(tt.filed, rules)
			_, lookup := valuecompare.WordKeys(tt.probe, rules)
			if got := shares(index, lookup); got != tt.meets {
				t.Errorf("meets = %v, want %v: index %q lookup %q", got, tt.meets, index, lookup)
			}
		})
	}
	t.Run("off keeps today's keys", func(t *testing.T) {
		rules := matchregistry.Default().Value
		rules.Names.PrefixMinLength = 0
		index, lookup := valuecompare.WordKeys(slGiven("Desiree"), rules)
		if !slices.Equal(index, lookup) || slices.ContainsFunc(index, prefixKey) {
			t.Errorf("index %q lookup %q, want equal and no prefix key", index, lookup)
		}
	})
	t.Run("too many words", func(t *testing.T) {
		rules := matchregistry.Default().Value
		words := make([]string, rules.Names.MaxWords+1)
		for i := range words {
			words[i] = "Desiree" + strings.Repeat("a", i)
		}
		index, lookup := valuecompare.WordKeys(slGiven(strings.Join(words, " ")), rules)
		if slices.ContainsFunc(index, prefixKey) || slices.ContainsFunc(lookup, prefixKey) {
			t.Errorf("index %q lookup %q hold a prefix key", index, lookup)
		}
		if !slices.Contains(index, "w:desiree") || !slices.Contains(lookup, "w:desiree") {
			t.Errorf("index %q lookup %q lack the word key", index, lookup)
		}
	})
	t.Run("run-on word", func(t *testing.T) {
		rules := matchregistry.Default().Value
		index, lookup := valuecompare.WordKeys(slGiven(strings.Repeat("x", rules.Words.MaxSpellingLength+1)), rules)
		if slices.ContainsFunc(index, prefixKey) || slices.ContainsFunc(lookup, prefixKey) {
			t.Errorf("index %q lookup %q hold a prefix key", index, lookup)
		}
	})
}

func shares(a, b []string) bool {
	return slices.ContainsFunc(a, func(k string) bool { return slices.Contains(b, k) })
}

func TestWordKeysShareOnSameOrEquivalent(t *testing.T) {
	rules := billRules()
	corpus := []string{
		"James Robins", "J. Robins", "Robins", "José", "Jose", "JOSE", "Mc Donald", "McDonald",
		"Saint-Jean", "SaintJean", "Bill", "William", "Bill Robins", "William Robins", "Robert Jr", "Robert", "Ann-Marie", "Ann Marie",
		"Ó Briain", "O Briain", "J. R.", "J R", "Mary Robins", "Robnis", "Müller", "Muller",
	}
	var equivalentAlias, sameBreak bool
	for _, typ := range []string{propertymeta.ValueTypeName, propertymeta.ValueTypeText} {
		mk := func(f string) valuecompare.Value {
			if typ == propertymeta.ValueTypeName {
				return slName(f)
			}
			return slText(f)
		}
		for _, x := range corpus {
			for _, y := range corpus {
				a, b := mk(x), mk(y)
				res, det := valuecompare.Compare(a, b, rules)
				if res.Relation != valuecompare.Same && res.Relation != valuecompare.Equivalent {
					continue
				}
				ia, la := valuecompare.WordKeys(a, rules)
				ib, lb := valuecompare.WordKeys(b, rules)
				if !shares(ia, lb) || !shares(ib, la) {
					t.Errorf("%s %q ~ %q is %s but the sides do not meet: index %q lookup %q vs index %q lookup %q", typ, x, y, res.Relation,
						ia, la, ib, lb)
				}
				if typ == propertymeta.ValueTypeName && res.Relation == valuecompare.Equivalent && det.Cause == valuecompare.CauseAlias {
					equivalentAlias = true
				}
				if typ == propertymeta.ValueTypeText && res.Relation == valuecompare.Same && x != y && len(valuecompare.SplitWords(x, rules.Words)) != len(valuecompare.SplitWords(y, rules.Words)) {
					sameBreak = true
				}
			}
		}
	}
	if !equivalentAlias {
		t.Error("the corpus met no Equivalent alias pair")
	}
	if !sameBreak {
		t.Error("the corpus met no Same word-break pair")
	}
}

// A pair the comparer gives the cause prefix meets across index and lookup, in
// both directions.
func TestWordKeysMeetOnPrefixPairs(t *testing.T) {
	rules := billRules()
	corpus := []string{"Des Long", "Desiree Long", "Al Robins", "Albert Robins", "Mary Mac", "Mary MacDonald", "Desmond Long"}
	met := false
	for _, x := range corpus {
		for _, y := range corpus {
			a, b := slName(x), slName(y)
			if _, det := valuecompare.Compare(a, b, rules); det.Cause != valuecompare.CausePrefix {
				continue
			}
			met = true
			ia, la := valuecompare.WordKeys(a, rules)
			ib, lb := valuecompare.WordKeys(b, rules)
			if !shares(ia, lb) || !shares(ib, la) {
				t.Errorf("%q ~ %q is a prefix pair but the sides do not meet: index %q lookup %q vs index %q lookup %q", x, y, ia, la, ib, lb)
			}
		}
	}
	if !met {
		t.Error("the corpus met no prefix pair")
	}
}

func TestAliasSetKeys(t *testing.T) {
	w := matchregistry.Default().Value.Words
	given := []valuecompare.NameRole{valuecompare.RoleGiven}
	idx := valuecompare.NewAliasIndex([]valuecompare.AliasTable{
		{Key: "a", Kind: valuecompare.AliasNickname, Roles: given, Sets: [][]string{{"chris", "christine"}}},
		{Key: "b", Kind: valuecompare.AliasNickname, Roles: given, Sets: [][]string{{"chris", "christopher"}}},
	}, w)
	chris := idx.SetKeys("chris")
	if len(chris) != 2 {
		t.Fatalf("SetKeys(chris) = %q, want one key per table", chris)
	}
	if got := idx.SetKeys("christine"); len(got) != 1 || !slices.Contains(chris, got[0]) {
		t.Errorf("SetKeys(christine) = %q, want one of %q", got, chris)
	}
	if got := idx.SetKeys("christopher"); len(got) != 1 || !slices.Contains(chris, got[0]) || slices.Equal(got, idx.SetKeys("christine")) {
		t.Errorf("SetKeys(christopher) = %q, want the other table's key of %q", got, chris)
	}
	if got := (valuecompare.AliasIndex{}).SetKeys("chris"); got != nil {
		t.Errorf("zero index SetKeys = %q, want nil", got)
	}
}

func TestAliasSetKeysDoNotGrowWithTheSet(t *testing.T) {
	w := matchregistry.Default().Value.Words
	given := []valuecompare.NameRole{valuecompare.RoleGiven}
	big := make([]string, 1000)
	for i := range big {
		big[i] = fmt.Sprintf("name%04d", i)
	}
	idx := valuecompare.NewAliasIndex([]valuecompare.AliasTable{
		{Key: "t", Kind: valuecompare.AliasNickname, Roles: given, Sets: [][]string{big, {"zed", "zeddie"}}},
	}, w)
	first, last := idx.SetKeys("name0000"), idx.SetKeys("name0999")
	for _, keys := range [][]string{first, last, idx.SetKeys("zeddie")} {
		for _, k := range keys {
			if len(k) >= 128 {
				t.Errorf("key is %d bytes, want under 128", len(k))
			}
		}
	}
	if !slices.Equal(first, last) || len(first) != 1 {
		t.Errorf("two words of one set keys = %q, %q; want the same single key", first, last)
	}
	if other := idx.SetKeys("zed"); slices.Equal(other, first) {
		t.Errorf("words of two sets share key %q", other)
	}
}

func TestAliasSetKeysDoNotGrowWithTheTokenOrTableKey(t *testing.T) {
	w := matchregistry.Default().Value.Words
	given := []valuecompare.NameRole{valuecompare.RoleGiven}
	long := strings.Repeat("x", 4096)
	idx := valuecompare.NewAliasIndex([]valuecompare.AliasTable{
		{Key: long, Kind: valuecompare.AliasNickname, Roles: given, Sets: [][]string{{long + "a", long + "b"}}},
	}, w)
	first, second := idx.SetKeys(long+"a"), idx.SetKeys(long+"b")
	for _, keys := range [][]string{first, second} {
		for _, k := range keys {
			if len(k) > 64 {
				t.Errorf("key is %d bytes, want at most 64", len(k))
			}
		}
	}
	if len(first) != 1 || !slices.Equal(first, second) {
		t.Errorf("two words of one set keys = %q, %q; want the same single key", first, second)
	}
}

func TestYearReach(t *testing.T) {
	rules := matchregistry.Default().Value
	tests := []struct {
		name      string
		d         *datevalue.Value
		lo, hi    int
		open, ok  bool
		checkLoHi bool
	}{
		{name: "year", d: datevaluetest.YMD(1850, 0, 0), lo: 1850 - int(rules.Dates.Tolerance), hi: 1850 + int(rules.Dates.Tolerance), ok: true, checkLoHi: true},
		{name: "before is open", d: datevaluetest.Before(1850, 0, 0), open: true, ok: true},
		{name: "after is open", d: datevaluetest.After(1850, 0, 0), open: true, ok: true},
		{name: "nil", d: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lo, hi, open, ok := valuecompare.YearReach(slDate(tt.d), rules)
			if open != tt.open || ok != tt.ok {
				t.Fatalf("open, ok = %v, %v; want %v, %v", open, ok, tt.open, tt.ok)
			}
			if tt.checkLoHi && (lo != tt.lo || hi != tt.hi) {
				t.Errorf("reach = %d..%d, want %d..%d", lo, hi, tt.lo, tt.hi)
			}
		})
	}
	// ABT widens the reach by AboutUnits of its most precise part.
	plainLo, plainHi, _, _ := valuecompare.YearReach(slDate(datevaluetest.YMD(1850, 0, 0)), rules)
	aboutLo, aboutHi, open, ok := valuecompare.YearReach(slDate(datevaluetest.About(1850, 0, 0)), rules)
	if !ok || open || aboutLo > plainLo-rules.Dates.AboutUnits || aboutHi < plainHi+rules.Dates.AboutUnits {
		t.Errorf("ABT reach = %d..%d, plain %d..%d, AboutUnits %d", aboutLo, aboutHi, plainLo, plainHi, rules.Dates.AboutUnits)
	}
}

func TestYearReachCoversEveryRelatedPair(t *testing.T) {
	rules := matchregistry.Default().Value
	var corpus []*datevalue.Value
	for y := 1846; y <= 1854; y++ {
		corpus = append(corpus, datevaluetest.YMD(y, 0, 0))
	}
	corpus = append(corpus,
		datevaluetest.YMD(1850, 3, 0), datevaluetest.YMD(1850, 3, 4), datevaluetest.About(1850, 0, 0),
		datevaluetest.Before(1850, 0, 0), datevaluetest.After(1850, 0, 0), datevaluetest.Between(1840, 0, 0, 1845, 0, 0), datevaluetest.Between(1840, 0, 0, 1870, 0, 0),
	)
	for _, x := range corpus {
		for _, y := range corpus {
			a, b := slDate(x), slDate(y)
			res, _ := valuecompare.Compare(a, b, rules)
			if res.Relation == valuecompare.Different || res.Relation == valuecompare.Unknown {
				continue
			}
			alo, ahi, aopen, aok := valuecompare.YearReach(a, rules)
			blo, bhi, bopen, bok := valuecompare.YearReach(b, rules)
			if !aok || !bok {
				t.Errorf("%+v ~ %+v is %s but a reach is not ok", x, y, res.Relation)
				continue
			}
			if !aopen && !bopen && (ahi < blo || bhi < alo) {
				t.Errorf("%+v ~ %+v is %s but reaches %d..%d and %d..%d do not overlap", x, y, res.Relation, alo, ahi, blo, bhi)
			}
		}
	}
}
