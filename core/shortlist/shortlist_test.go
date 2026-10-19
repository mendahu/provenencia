package shortlist_test

import (
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertycompare"
	pt "github.com/mendahu/provenencia/core/propertycompare/propertycomparetest"
	"github.com/mendahu/provenencia/core/shortlist"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

type propertyValue = propertycompare.Value

func bef(y int) *datevalue.Value {
	v := datevaluetest.YMD(y, 0, 0)
	v.Qualifier = datevalue.QualifierBEF
	return v
}

func aft(y int) *datevalue.Value {
	v := datevaluetest.YMD(y, 0, 0)
	v.Qualifier = datevalue.QualifierAFT
	return v
}

func date(y int) propertyValue { return pt.Date(datevaluetest.YMD(y, 0, 0)) }

// entry is one handle: a one-byte id and its values.
type entry struct {
	id byte
	v  subjectcompare.Values
}

func person(id byte, form string) entry {
	return entry{id, subjectcompare.Values{vocabkey.PropertyName: {pt.Name(form)}}}
}

func place(id byte, s string) entry {
	return entry{id, subjectcompare.Values{vocabkey.PropertyToponym: {pt.Text(s)}}}
}

func event(id byte, typ string, d *datevalue.Value) entry {
	v := subjectcompare.Values{}
	if typ != "" {
		v[vocabkey.PropertyEventType] = []propertyValue{pt.Term(typ)}
	}
	if d != nil {
		v[vocabkey.PropertyDate] = []propertyValue{pt.Date(d)}
	}
	return entry{id, v}
}

func period(id byte, typ string, start, end *datevalue.Value) entry {
	v := subjectcompare.Values{vocabkey.PropertyEventType: {pt.Term(typ)}}
	if start != nil {
		v[vocabkey.PropertyStartDate] = []propertyValue{pt.Date(start)}
	}
	if end != nil {
		v[vocabkey.PropertyEndDate] = []propertyValue{pt.Date(end)}
	}
	return entry{id, v}
}

func build(rules valuecompare.Rules, kind vocabkey.SubjectType, entries ...entry) *shortlist.Index {
	r := matchregistry.Default()
	ix := shortlist.NewIndex(r.Alignment.Shortlist, r.Subject, rules, subjectcomparetest.Metas())
	for _, e := range entries {
		ix.Add(kind, []byte{e.id}, e.v)
	}
	return ix
}

func ids(got [][]byte) []byte {
	var out []byte
	for _, id := range got {
		out = append(out, id...)
	}
	return out
}

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

func TestLookup(t *testing.T) {
	plain := matchregistry.Default().Value
	events := []entry{
		event(1, "birth", datevaluetest.YMD(1850, 0, 0)),
		event(2, "birth", datevaluetest.YMD(1852, 0, 0)),
		event(3, "birth", datevaluetest.YMD(1870, 0, 0)),
		event(4, "death", datevaluetest.YMD(1850, 0, 0)),
		event(5, "birth", nil),
	}
	tests := []struct {
		name    string
		ac      string
		rules   valuecompare.Rules
		kind    vocabkey.SubjectType
		entries []entry
		probe   entry
		want    []byte
	}{
		{"shared word", "AC-1.1", plain, vocabkey.Person,
			[]entry{person(1, "James Robins"), person(2, "Robins"), person(3, "Mary Robins"), person(4, "James Robbins"), person(5, "Ann Ames")},
			person(0, "J. Robins"), []byte{1, 2, 3}},
		{"alias set", "AC-1.2", billRules(), vocabkey.Person,
			[]entry{person(1, "William"), person(2, "Robert")},
			person(0, "Bill"), []byte{1}},
		{"folded word", "AC-1.3", plain, vocabkey.Person,
			[]entry{person(1, "José Robins")}, person(0, "Jose Robins"), []byte{1}},
		{"whole toponym", "AC-1.4", plain, vocabkey.Place,
			[]entry{place(1, "Saint-Jean"), place(2, "York")}, place(0, "SaintJean"), []byte{1}},
		{"toponym word", "AC-1.4", plain, vocabkey.Place,
			[]entry{place(1, "Saint-Jean"), place(2, "York")}, place(0, "New York"), []byte{2}},
		{"type and year", "AC-1.5", plain, vocabkey.Event, events,
			event(0, "birth", datevaluetest.YMD(1850, 0, 0)), []byte{1, 2}},
		{"open index date", "AC-1.6", plain, vocabkey.Event,
			append(slices.Clone(events), event(6, "birth", bef(1800))),
			event(0, "birth", datevaluetest.YMD(1700, 0, 0)), []byte{6}},
		{"open probe date", "AC-1.6", plain, vocabkey.Event,
			append(slices.Clone(events), event(6, "birth", bef(1800))),
			event(0, "birth", aft(1860)), []byte{1, 2, 3, 6}},
		{"huge range probe", "AC-2.2", plain, vocabkey.Event,
			append(slices.Clone(events), event(6, "birth", datevaluetest.Between(1, 0, 0, 10_000_000, 0, 0))),
			event(0, "birth", datevaluetest.Between(1, 0, 0, 10_000_000, 0, 0)), []byte{1, 2, 3, 6}},
		{"no date to look up", "AC-1.7", plain, vocabkey.Event, events, event(0, "birth", nil), nil},
		{"no lead of the kind", "AC-1.7", plain, vocabkey.Person,
			[]entry{person(1, "James Robins")},
			entry{0, subjectcompare.Values{vocabkey.PropertySexAtBirth: {pt.Term("male")}}}, nil},
		{"kind with no rules", "AC-1.7", plain, vocabkey.SubjectType("source"),
			[]entry{person(1, "James Robins")}, person(0, "James Robins"), nil},
		{"period contains", "AC-1.8", plain, vocabkey.Event,
			[]entry{
				period(7, "residence", datevaluetest.YMD(1840, 0, 0), datevaluetest.YMD(1870, 0, 0)),
				period(8, "residence", datevaluetest.YMD(1850, 0, 0), nil),
			},
			event(0, "residence", datevaluetest.YMD(1855, 0, 0)), []byte{7, 8}},
		{"period misses", "AC-1.8", plain, vocabkey.Event,
			[]entry{
				period(7, "residence", datevaluetest.YMD(1840, 0, 0), datevaluetest.YMD(1870, 0, 0)),
				period(8, "residence", datevaluetest.YMD(1850, 0, 0), nil),
			},
			event(0, "residence", datevaluetest.YMD(1900, 0, 0)), []byte{8}},
		{"date hides the period", "AC-1.9", plain, vocabkey.Event,
			[]entry{{9, subjectcompare.Values{
				vocabkey.PropertyEventType: {pt.Term("residence")},
				vocabkey.PropertyDate:      {date(1855)},
				vocabkey.PropertyStartDate: {date(1700)},
				vocabkey.PropertyEndDate:   {date(1710)},
			}}},
			event(0, "residence", datevaluetest.YMD(1705, 0, 0)), nil},
		{"untyped index event", "AC-1.10", plain, vocabkey.Event,
			append(slices.Clone(events), event(9, "", datevaluetest.YMD(1850, 0, 0))),
			event(0, "birth", datevaluetest.YMD(1850, 0, 0)), []byte{1, 2, 9}},
		{"untyped probe", "AC-1.10", plain, vocabkey.Event,
			append(slices.Clone(events), event(9, "", datevaluetest.YMD(1850, 0, 0))),
			event(0, "", datevaluetest.YMD(1851, 0, 0)), []byte{1, 2, 4, 9}},
	}
	for _, tt := range tests {
		t.Run(tt.ac+"/"+tt.name, func(t *testing.T) {
			got := ids(build(tt.rules, tt.kind, tt.entries...).Lookup(tt.kind, tt.probe.v))
			if !slices.Equal(got, tt.want) {
				t.Errorf("Lookup = %v, want %v", got, tt.want)
			}
		})
	}
	var nilIndex *shortlist.Index
	if got := nilIndex.Lookup(vocabkey.Person, person(0, "James Robins").v); got != nil {
		t.Errorf("nil index Lookup = %v, want nil", got)
	}
}

func TestLookupCopiesAndSorts(t *testing.T) {
	ix := build(matchregistry.Default().Value, vocabkey.Person, person(3, "Robins"), person(1, "Robins"), person(2, "James Robins"))
	got := ix.Lookup(vocabkey.Person, person(0, "Robins").v)
	if want := []byte{1, 2, 3}; !slices.Equal(ids(got), want) {
		t.Fatalf("Lookup = %v, want %v", ids(got), want)
	}
	got[0][0] = 99
	if again := ids(ix.Lookup(vocabkey.Person, person(0, "Robins").v)); !slices.Equal(again, []byte{1, 2, 3}) {
		t.Errorf("a changed result changed the index: %v", again)
	}
}

func TestKeys(t *testing.T) {
	r := matchregistry.Default()
	metas := subjectcomparetest.Metas()
	keys := func(kind vocabkey.SubjectType, v subjectcompare.Values, rules shortlist.Rules) ([]string, []string) {
		return shortlist.Keys(kind, v, metas, rules, r.Subject, r.Value)
	}
	t.Run("negative value", func(t *testing.T) {
		v := subjectcompare.Values{vocabkey.PropertyName: {pt.Neg(pt.Name("James Robins"))}}
		index, lookup := keys(vocabkey.Person, v, r.Alignment.Shortlist)
		if len(index) != 0 || len(lookup) != 0 {
			t.Errorf("keys = %q, %q; want none", index, lookup)
		}
	})
	t.Run("no year bucket", func(t *testing.T) {
		rules := r.Alignment.Shortlist
		rules.YearBucket = 0
		index, lookup := keys(vocabkey.Event, event(0, "birth", datevaluetest.YMD(1850, 0, 0)).v, rules)
		if len(index) != 0 || len(lookup) != 0 {
			t.Errorf("keys = %q, %q; want none", index, lookup)
		}
	})
	t.Run("negative years", func(t *testing.T) {
		index, _ := keys(vocabkey.Event, event(0, "birth", datevaluetest.YMD(1, 0, 0)).v, r.Alignment.Shortlist)
		if !slices.Contains(index, "t:birth|-1") {
			t.Errorf("keys %q lack the bucket below zero", index)
		}
	})
	t.Run("huge range keys as open", func(t *testing.T) {
		huge := event(0, "birth", datevaluetest.Between(1, 0, 0, 10_000_000, 0, 0))
		index, lookup := keys(vocabkey.Event, huge.v, r.Alignment.Shortlist)
		if len(index) >= 10 || len(lookup) >= 10 {
			t.Fatalf("keys = %d, %d; want fewer than 10 each", len(index), len(lookup))
		}
		if !slices.Contains(index, "t:birth|wide") {
			t.Errorf("index %q lacks the wide key", index)
		}
	})
	t.Run("sorted and unique", func(t *testing.T) {
		index, lookup := keys(vocabkey.Person, person(0, "Robins Robins").v, r.Alignment.Shortlist)
		for _, list := range [][]string{index, lookup} {
			if !slices.IsSorted(list) || len(slices.Compact(slices.Clone(list))) != len(list) {
				t.Errorf("%q is not sorted and unique", list)
			}
		}
	})
}
