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

// Pool lists properties whose values join the lead's on each side. Terms
// maps a pooled property's term keys into the lead's vocabulary, so the
// gender "man" meets the sex at birth "male"; a pooled property listed in
// Terms drops a term it does not map, and one not listed joins unchanged.
type Pool struct {
	With  []vocabkey.PropertyKey
	Terms map[vocabkey.PropertyKey]map[string]string
}

// Period reads two date properties as one range that stands for the lead on
// a side with no positive lead value.
type Period struct{ Start, End vocabkey.PropertyKey }

// EventRole names an event a person takes part in by its type and the
// person's role in it.
type EventRole struct {
	EventType vocabkey.EventType
	Role      vocabkey.Role
}

// Lifespan is a derived lifespan: the lead holds its window, where a life can
// lie, and Span its first to last dated event. Compare checks each side's span
// against the other side's window; the profile builder builds both.
type Lifespan struct {
	Span     vocabkey.PropertyKey // not a lead
	MaxYears int                  // pads an end no Starts or Ends event fixes
	Starts   []EventRole          // events that fix the start
	Ends     []EventRole          // events that fix the end
}

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
	// Lifespans lists, per Subject type and lead property, the derived
	// lifespan whose window is the lead and whose span is Span; the profile
	// builder builds both from life events.
	Lifespans map[vocabkey.SubjectType]map[vocabkey.PropertyKey]Lifespan
}

// Values are one Subject's values by property: product properties, and the
// derived properties the profile builder adds (a person's lifespan).
type Values map[vocabkey.PropertyKey][]propertycompare.Value

// Metas are the catalog's value type and cardinality for each property.
type Metas map[vocabkey.PropertyKey]propertymeta.Meta

// Result is what Compare decides about two Subjects.
type Result struct {
	Similarity float64          // 0 to 1: weighted mean over counted properties; 0 when none
	Evidence   float64          // points: agree weights minus conflict penalties
	Properties []PropertyResult // one per lead property, ascending key
}

// Conflicts reports whether any property comparison conflicts.
func (r Result) Conflicts() bool {
	return slices.ContainsFunc(r.Properties, func(p PropertyResult) bool {
		return p.Result.Outcome == propertycompare.OutcomeConflict
	})
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
		var res propertycompare.Result
		if l, ok := rules.Lifespans[kind][key]; ok {
			res = lifespan(key, a, b, l, value)
		} else {
			res = propertycompare.Compare(Side(kind, key, a, rules), Side(kind, key, b, rules), metas[key], value)
		}
		pr := PropertyResult{Key: key, Result: res}
		pr.Points = Points(res, w)
		score += max(pr.Points, 0)
		switch res.Outcome {
		case propertycompare.OutcomeAgree, propertycompare.OutcomePartial:
			counted += w.Agree
		case propertycompare.OutcomeConflict:
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

// Points is what one lead property adds to Evidence: an agree earns the agree
// weight, a partial the agree weight × similarity, a conflict loses the
// conflict penalty, and an unknown counts for nothing.
func Points(res propertycompare.Result, w Weights) float64 {
	switch res.Outcome {
	case propertycompare.OutcomeAgree:
		return w.Agree
	case propertycompare.OutcomePartial:
		return w.Agree * res.Similarity()
	case propertycompare.OutcomeConflict:
		return -w.Conflict
	}
	return 0
}

// Derived reports whether key is a property the profile builder derives for
// the kind: a lifespan's lead or its Span. A derived property has no catalog
// row.
func (r Rules) Derived(kind vocabkey.SubjectType, key vocabkey.PropertyKey) bool {
	for lead, l := range r.Lifespans[kind] {
		if key == lead || key == l.Span {
			return true
		}
	}
	return false
}

// Role is how a property counts under its lead.
type Role int

const (
	// RoleLead is a weighted property: it is its own lead.
	RoleLead Role = iota
	// RolePooled is a property a lead pools its values with.
	RolePooled
	// RolePeriodEnd is the start or end a lead reads as a period.
	RolePeriodEnd
)

// Role returns the lead property a property counts under for the kind and how
// it counts: a weighted property is its own lead, else the first lead, in
// ascending key order, that pools it or reads it as a period end. A
// lifespan's Span is in neither, so it has no lead.
func (r Rules) Role(kind vocabkey.SubjectType, key vocabkey.PropertyKey) (vocabkey.PropertyKey, Role, bool) {
	if _, ok := r.Weights[kind][key]; ok {
		return key, RoleLead, true
	}
	for _, lead := range slices.Sorted(maps.Keys(r.Weights[kind])) {
		if slices.Contains(r.Pools[kind][lead].With, key) {
			return lead, RolePooled, true
		}
		if p, ok := r.Periods[kind][lead]; ok && (p.Start == key || p.End == key) {
			return lead, RolePeriodEnd, true
		}
	}
	return "", RoleLead, false
}

// Lead returns the lead property a property counts under for the kind: the
// property itself when it is weighted, else the first lead, in ascending key
// order, that pools it or reads it as a period end. A lifespan's Span is in
// neither, so it has no lead.
func (r Rules) Lead(kind vocabkey.SubjectType, key vocabkey.PropertyKey) (vocabkey.PropertyKey, bool) {
	lead, _, ok := r.Role(kind, key)
	return lead, ok
}
