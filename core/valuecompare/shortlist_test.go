package valuecompare_test

import (
	"fmt"
	"slices"
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

func slBef(y int) *datevalue.Value {
	v := datevaluetest.YMD(y, 0, 0)
	v.Qualifier = datevalue.QualifierBEF
	return v
}

func slAft(y int) *datevalue.Value {
	v := datevaluetest.YMD(y, 0, 0)
	v.Qualifier = datevalue.QualifierAFT
	return v
}

func TestWordKeys(t *testing.T) {
	rules := billRules()
	whole := func(v valuecompare.Value) string {
		k, ok := valuecompare.Key(v, valuecompare.Equivalent, rules)
		if !ok {
			t.Fatalf("no key for %+v", v)
		}
		return "=" + k
	}
	tests := []struct {
		name string
		v    valuecompare.Value
		want []string // exact when set
		has  []string
		not  []string
	}{
		{name: "initial is no key", v: slName("J. Robins"), want: []string{whole(slName("J. Robins")), "w:robins"}},
		{name: "folded word", v: slName("José Robins"), has: []string{"w:jose", "w:robins"}},
		{name: "generation word", v: slGeneration(), has: []string{"w:john", "w:smith"}, not: []string{"w:jr"}},
		{name: "alias set", v: slName("Bill"), has: []string{"a:bill william"}},
		{name: "text words", v: slText("New York"), has: []string{"w:new", "w:york"}, not: []string{"a:bill william"}},
		{name: "integer", v: valuecompare.Value{Type: propertymeta.ValueTypeInteger, Integer: 4}},
		{name: "empty name", v: valuecompare.Value{Type: propertymeta.ValueTypeName, NamePattern: valuecompare.WesternPattern}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valuecompare.WordKeys(tt.v, rules)
			if tt.want != nil {
				want := slices.Clone(tt.want)
				slices.Sort(want)
				if !slices.Equal(got, want) {
					t.Fatalf("WordKeys = %q, want %q", got, want)
				}
			}
			if tt.want == nil && tt.has == nil && tt.not == nil && got != nil {
				t.Fatalf("WordKeys = %q, want nil", got)
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
			if !slices.IsSorted(got) || len(slices.Compact(slices.Clone(got))) != len(got) {
				t.Errorf("WordKeys %q is not sorted and unique", got)
			}
		})
	}
	if got := valuecompare.WordKeys(slText("New York"), rules); slices.ContainsFunc(got, func(k string) bool { return len(k) > 1 && k[:2] == "a:" }) {
		t.Errorf("text keys %q hold an alias key", got)
	}
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
				if !shares(valuecompare.WordKeys(a, rules), valuecompare.WordKeys(b, rules)) {
					t.Errorf("%s %q ~ %q is %s but shares no key: %q vs %q", typ, x, y, res.Relation,
						valuecompare.WordKeys(a, rules), valuecompare.WordKeys(b, rules))
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

func TestAliasSetKeys(t *testing.T) {
	w := matchregistry.Default().Value.Words
	given := []valuecompare.NameRole{valuecompare.RoleGiven}
	idx := valuecompare.NewAliasIndex([]valuecompare.AliasTable{
		{Key: "a", Kind: valuecompare.AliasNickname, Roles: given, Sets: [][]string{{"chris", "christine"}}},
		{Key: "b", Kind: valuecompare.AliasNickname, Roles: given, Sets: [][]string{{"chris", "christopher"}}},
	}, w)
	if got, want := idx.SetKeys("chris"), []string{"a chris", "b chris"}; !slices.Equal(got, want) {
		t.Errorf("SetKeys(chris) = %q, want %q", got, want)
	}
	if got, want := idx.SetKeys("christine"), []string{"a chris"}; !slices.Equal(got, want) {
		t.Errorf("SetKeys(christine) = %q, want %q", got, want)
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
		{name: "before is open", d: slBef(1850), open: true, ok: true},
		{name: "after is open", d: slAft(1850), open: true, ok: true},
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
		slBef(1850), slAft(1850), datevaluetest.Between(1840, 0, 0, 1845, 0, 0), datevaluetest.Between(1840, 0, 0, 1870, 0, 0),
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
