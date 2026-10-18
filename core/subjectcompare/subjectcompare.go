// Package subjectcompare is layer 3 of matching: how alike two Subjects of one
// type are over the properties both record. Pure: no I/O, no catalog, no
// numbers; the subject section and layer 1's rules come in as arguments.
package subjectcompare

import (
	"maps"
	"slices"

	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Weights are one lead property's part in a Subject comparison, in points.
type Weights struct {
	Agree    float64 // earned on agree; Agree × similarity on partial
	Conflict float64 // lost on conflict; also its weight in Similarity
}

// Pool lists properties whose values join the lead's on each side.
type Pool struct{ With []vocabkey.PropertyKey }

// Period reads two date properties as one range that stands for the lead on
// a side with no positive lead value.
type Period struct{ Start, End vocabkey.PropertyKey }

// Rules is the match registry's subject section.
type Rules struct {
	// Weights lists, per Subject type, the lead properties compared and what
	// each counts.
	Weights map[vocabkey.SubjectType]map[vocabkey.PropertyKey]Weights
	// Pools lists, per Subject type and lead property, the properties whose
	// values are compared with the lead's.
	Pools map[vocabkey.SubjectType]map[vocabkey.PropertyKey]Pool
	// Periods lists, per Subject type and lead property, the start and end
	// properties read as one range for a side with no positive lead value.
	Periods map[vocabkey.SubjectType]map[vocabkey.PropertyKey]Period
}

// Values are one Subject's values by property. Product properties only.
type Values map[vocabkey.PropertyKey][]propertycompare.Value

// Metas are the catalog's value type and cardinality for each property.
type Metas map[vocabkey.PropertyKey]propertymeta.Meta

// Result is what Compare decides about two Subjects.
type Result struct {
	Similarity float64          // 0 to 1: weighted mean over counted properties; 0 when none
	Evidence   float64          // points: agree weights minus conflict penalties
	Properties []PropertyResult // one per lead property, ascending key
}

// PropertyResult is layer 2's verdict on one lead property and what it added.
type PropertyResult struct {
	Key    vocabkey.PropertyKey // the lead property
	Result propertycompare.Result
	Points float64 // what it added to Evidence
}

// Compare says how alike two Subjects of one type are. It runs layer 2 once
// for each lead property in rules.Weights[kind], in ascending key order, on
// the values the property, its pool or its period gives each side; an agree
// earns the agree weight, a partial the agree weight × the speaking pair's
// similarity, a conflict loses the conflict penalty, and an unknown counts for
// nothing. A kind with no Weights entry gives the zero Result.
func Compare(kind vocabkey.SubjectType, a, b Values, metas Metas, rules Rules, value valuecompare.Rules) Result {
	weights, ok := rules.Weights[kind]
	if !ok {
		return Result{}
	}
	var out Result
	var counted, score float64
	for _, key := range slices.Sorted(maps.Keys(weights)) {
		w := weights[key]
		res := propertycompare.Compare(side(kind, key, a, rules), side(kind, key, b, rules), metas[key], value)
		pr := PropertyResult{Key: key, Result: res}
		switch res.Outcome {
		case propertycompare.OutcomeAgree:
			pr.Points = w.Agree
			counted += w.Agree
			score += w.Agree
		case propertycompare.OutcomePartial:
			s := res.Speaking.Result.Similarity
			pr.Points = w.Agree * s
			counted += w.Agree
			score += w.Agree * s
		case propertycompare.OutcomeConflict:
			pr.Points = -w.Conflict
			counted += w.Conflict
		}
		out.Evidence += pr.Points
		out.Properties = append(out.Properties, pr)
	}
	if counted > 0 {
		out.Similarity = score / counted
	}
	return out
}
