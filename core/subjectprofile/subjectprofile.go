// Package subjectprofile is the profile builder beside layer 4 of matching:
// it turns an entity's reconciled conclusion, or a Subject's own
// Observations, into the Subject-shaped values layer 3 compares. It compares
// nothing. Pure: no I/O, no catalog, no numbers; the reconciler's output and
// rules come in as arguments.
package subjectprofile

import (
	"bytes"
	"maps"
	"slices"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Entity is one canonical entity's auto-reconciled product Properties.
type Entity struct {
	Properties  map[vocabkey.PropertyKey]Property
	NamePattern string // the catalog's name format key, set on every name value
}

// Property is the reconciler's output for one Property and, for leave-one-out,
// its input.
type Property struct {
	Values     []autoreconcile.ReconciledValue // a Reconcile result's Values, or the cache's rows
	Candidates []Candidate                     // read only by Options.Excluding
}

// Candidate is a reconciler candidate and the member Subject its Observation
// is on.
type Candidate struct {
	Member []byte
	autoreconcile.Candidate
}

// Options change how Build reads the entity.
type Options struct {
	// Excluding builds the profile as if these members' Observations were not
	// in the group: each Property they recorded is reconciled again without them.
	Excluding [][]byte
}

// Build returns the entity's profile: on each Property, the displayed values
// as positives and the values a stronger negative denied as negatives;
// outvoted, weak and provisional values are left out, and a Property with
// nothing left is absent. reconcile is the rules the reconciler runs under:
// pass matchregistry.Default().Value, as the cache does, not a catalog's
// valuerules. metas give each Property's value type and cardinality for
// Options.Excluding. The error is Reconcile's, and only Excluding runs it.
func Build(e Entity, metas subjectcompare.Metas, reconcile valuecompare.Rules, opts Options) (subjectcompare.Values, error) {
	out := subjectcompare.Values{}
	for _, key := range slices.Sorted(maps.Keys(e.Properties)) {
		p := e.Properties[key]
		vals := p.Values
		if kept, touched := without(p.Candidates, opts.Excluding); touched {
			meta := metas[key]
			res, err := autoreconcile.Reconcile(reconcile, meta.ValueType, kept, nil, meta.CardinalityOrSingle())
			if err != nil {
				return nil, err
			}
			vals = res.Values
		}
		var profile []propertycompare.Value
		for _, v := range vals {
			switch {
			case v.Displayed():
				profile = append(profile, value(v.Value, false, e.NamePattern))
			case v.Reason == autoreconcile.ReasonDenied:
				profile = append(profile, value(v.Value, true, e.NamePattern))
			}
		}
		if len(profile) > 0 {
			out[key] = profile
		}
	}
	return out, nil
}

// without returns the candidates not on an excluded member, and whether any
// candidate was.
func without(cs []Candidate, excluding [][]byte) ([]autoreconcile.Candidate, bool) {
	if len(excluding) == 0 {
		return nil, false
	}
	var kept []autoreconcile.Candidate
	touched := false
	for _, c := range cs {
		if slices.ContainsFunc(excluding, func(m []byte) bool { return bytes.Equal(m, c.Member) }) {
			touched = true
			continue
		}
		kept = append(kept, c.Candidate)
	}
	return kept, touched
}

// ForSubject returns a Subject's own values: each candidate positive or
// negative as recorded, in order. Only Value and Negative are read.
func ForSubject(values map[vocabkey.PropertyKey][]autoreconcile.Candidate, namePattern string) subjectcompare.Values {
	out := subjectcompare.Values{}
	for key, cs := range values {
		for _, c := range cs {
			out[key] = append(out[key], value(c.Value, c.Negative, namePattern))
		}
	}
	return out
}

// value is the one conversion from the reconciler's value to layer 2's. Type
// and Ref stay empty: layer 2 sets the type from the meta. Name and Date are
// shared with v.
func value(v autoreconcile.Value, negative bool, namePattern string) propertycompare.Value {
	out := propertycompare.Value{
		Value: valuecompare.Value{
			Text:    v.Text,
			Term:    v.TermKey,
			Integer: v.Integer,
			Name:    v.Name,
			Date:    v.Date,
		},
		Negative: negative,
		Neutral:  v.Neutral,
	}
	if v.Name != nil {
		out.NamePattern = namePattern
	}
	return out
}
