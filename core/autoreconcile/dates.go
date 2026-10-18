package autoreconcile

import (
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// The date module (conclusion-reconciliation.md §7.1). A date is the span of
// days value compare reads: ABT is a range of one unit of its most precise
// part, a clock time or a calendar is not compared, and missing finer fields
// are unknown, never zero.
//
// Dates with the same span share a key and merge. A wider date folds into a
// strictly narrower one it contains: MAY 1985 into 14 MAY 1985, BEF 1900 into
// 1885, ABT 1985 into 1986. Disjoint points stay apart (APR and MAY, 3 MAY
// and 14 JUN): they are not widened into a range and not collapsed to the year
// they share. An open window folds into each closed date it contains and does
// not glue those dates together.
//
// A phrase with no year is no evidence. Time zones are not compared.

// dateSortOffset shifts a proleptic day number into a zero-padded sort key
// that still orders dates before year 1.
const dateSortOffset = 100_000_000

type dateModule struct{ rules valuecompare.Rules }

func (m dateModule) split(v Value) (map[string]unit, bool) {
	if v.Date == nil {
		return nil, false
	}
	cmp := v.compareValue(propertymeta.ValueTypeDate)
	key, ok := valuecompare.Key(cmp, valuecompare.Same, m.rules)
	if !ok {
		return nil, false
	}
	return map[string]unit{wholeValue: {key: key, data: cmp}}, true
}

// fold: from is the wider date and into is the narrower one it contains.
func (m dateModule) fold(_ string, from, into unit) bool {
	res, det := valuecompare.Compare(from.data.(valuecompare.Value), into.data.(valuecompare.Value), m.rules)
	return res.Relation == valuecompare.Contains && det.Broader == valuecompare.SideA
}

// outvotes is false: a majority never crowds out a disjoint date.
func (dateModule) outvotes(string, unit, unit) bool { return false }

func (dateModule) oneValue() bool { return false }

func (dateModule) assemble(_ []Candidate, settled map[string][]unit) Value {
	d := *settled[wholeValue][0].data.(valuecompare.Value).Date
	return Value{Date: &d}
}
