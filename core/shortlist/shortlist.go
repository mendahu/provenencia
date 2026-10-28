// Package shortlist is layer 6's first step: the keys a Subject is looked up by
// and the index that maps them to handles. Layer 1 makes every key and layer 3
// picks the values, so shortlisting reads a Subject as comparison does. It is
// pure.
package shortlist

import (
	"slices"
	"sort"
	"strconv"

	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Rules say which leads key a Subject of each kind.
type Rules struct {
	// Words are the leads whose word keys key a Subject of the kind.
	Words map[vocabkey.SubjectType][]vocabkey.PropertyKey
	// Dated are the kinds keyed by a type term with the year buckets of a date.
	Dated map[vocabkey.SubjectType]Dated
	// YearBucket is the years per bucket.
	YearBucket int
	// MaxBuckets is the most buckets a date may key; a wider reach keys as an
	// open one, so a hostile range cannot make a million keys.
	MaxBuckets int
}

// Dated is a kind keyed by a type term and the year buckets of a date.
type Dated struct {
	// Type is the term lead keyed with each bucket.
	Type vocabkey.PropertyKey
	// Date is the date lead whose reach is bucketed; its period comes from the
	// subject section.
	Date vocabkey.PropertyKey
}

const (
	anyType    = "*" // an untyped probe meets every type
	noType     = ""  // an untyped event, met by every typed probe
	datePrefix = "t:"
)

// positive is the positive values Compare reads for lead, typed as the
// property.
func positive(kind vocabkey.SubjectType, lead vocabkey.PropertyKey, v subjectcompare.Values, metas subjectcompare.Metas, subject subjectcompare.Rules) []propertycompare.Value {
	var out []propertycompare.Value
	for _, x := range subjectcompare.Side(kind, lead, v, subject) {
		if x.Positive() {
			out = append(out, propertycompare.Typed(x, metas[lead]))
		}
	}
	return out
}

// Keys are the keys a Subject of kind is filed under (index) and looked up by
// (lookup). Two Subjects that can reach the weak band share an index key and a
// lookup key. Both lists are sorted, with no duplicates.
//
// Word leads file every index word key of every positive value and look up
// every lookup word key (valuecompare.WordKeys). A dated kind keys its type
// tokens with the year buckets its date can still be Near to (a reach wider
// than MaxBuckets counts as open): a closed reach is filed under each bucket it
// covers and looked up under each, an open one meets every closed one through
// "wide" and "any" keys. An event with no type is filed under every type, and a
// probe with no type meets every event.
func Keys(kind vocabkey.SubjectType, v subjectcompare.Values, metas subjectcompare.Metas, r Rules, subject subjectcompare.Rules, value valuecompare.Rules) (index, lookup []string) {
	for _, lead := range r.Words[kind] {
		for _, x := range positive(kind, lead, v, metas, subject) {
			ik, lk := valuecompare.WordKeys(x.Value, value)
			for _, k := range ik {
				index = append(index, string(lead)+"|"+k)
			}
			for _, k := range lk {
				lookup = append(lookup, string(lead)+"|"+k)
			}
		}
	}
	if d, ok := r.Dated[kind]; ok && r.YearBucket >= 1 {
		var types []string
		for _, t := range positive(kind, d.Type, v, metas, subject) {
			if k, ok := valuecompare.Key(t.Value, valuecompare.Equivalent, value); ok {
				types = append(types, k)
			}
		}
		slices.Sort(types)
		types = slices.Compact(types)
		indexTypes := append(slices.Clone(types), anyType)
		lookupTypes := append(slices.Clone(types), noType)
		if len(types) == 0 {
			indexTypes = []string{noType, anyType}
			lookupTypes = []string{anyType}
		}
		for _, date := range positive(kind, d.Date, v, metas, subject) {
			lo, hi, open, ok := valuecompare.YearReach(date.Value, value)
			if !ok {
				continue
			}
			if !open && floorDiv(hi, r.YearBucket)-floorDiv(lo, r.YearBucket)+1 > r.MaxBuckets {
				open = true
			}
			for _, k := range indexTypes {
				base := datePrefix + k + "|"
				if open {
					index = append(index, base+"wide", base+"any")
					continue
				}
				index = append(index, base+"any")
				index = appendBuckets(index, base, lo, hi, r.YearBucket)
			}
			for _, k := range lookupTypes {
				base := datePrefix + k + "|"
				if open {
					lookup = append(lookup, base+"any")
					continue
				}
				lookup = append(lookup, base+"wide")
				lookup = appendBuckets(lookup, base, lo, hi, r.YearBucket)
			}
		}
	}
	return sortedUnique(index), sortedUnique(lookup)
}

// appendBuckets adds base and each bucket from lo to hi.
func appendBuckets(keys []string, base string, lo, hi, width int) []string {
	for b := floorDiv(lo, width); b <= floorDiv(hi, width); b++ {
		keys = append(keys, base+strconv.Itoa(b))
	}
	return keys
}

func sortedUnique(keys []string) []string {
	sort.Strings(keys)
	return slices.Compact(keys)
}

// floorDiv rounds y / w toward minus infinity.
func floorDiv(y, w int) int {
	q := y / w
	if y%w != 0 && y < 0 {
		q--
	}
	return q
}
