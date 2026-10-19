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
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Entity is one canonical entity's auto-reconciled product Properties.
type Entity struct {
	Properties  map[vocabkey.PropertyKey]Property
	NamePattern string // the catalog's name format key, set on every name value
	Life        []LifeEvent
}

// LifeEvent is one event a person takes part in: its type, the person's role
// in it, and every date the event records.
type LifeEvent struct {
	EventType vocabkey.EventType
	Role      vocabkey.Role
	Dates     []datevalue.Value
}

// Subject is a plain Subject's own Observations by property and its life
// events.
type Subject struct {
	Values      map[vocabkey.PropertyKey][]autoreconcile.Candidate
	NamePattern string // the catalog's name format key, set on every name value
	Life        []LifeEvent
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
// Options.Excluding. lifespans is the registry's Subject.Lifespans for the
// side's kind; for each, a window and a span are added when the life events
// give them. The error is Reconcile's, and only Excluding runs it.
func Build(e Entity, metas subjectcompare.Metas, reconcile valuecompare.Rules, lifespans map[vocabkey.PropertyKey]subjectcompare.Lifespan, opts Options) (subjectcompare.Values, error) {
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
				profile = append(profile, Value(v.Value, false, e.NamePattern))
			case v.Reason == autoreconcile.ReasonDenied:
				profile = append(profile, Value(v.Value, true, e.NamePattern))
			}
		}
		if len(profile) > 0 {
			out[key] = profile
		}
	}
	addLifespans(out, e.Life, lifespans)
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
// negative as recorded, in order. Only Value and Negative are read. lifespans
// is the registry's Subject.Lifespans for the side's kind; for each, a window
// and a span are added when the life events give them.
func ForSubject(s Subject, lifespans map[vocabkey.PropertyKey]subjectcompare.Lifespan) subjectcompare.Values {
	out := subjectcompare.Values{}
	for key, cs := range s.Values {
		for _, c := range cs {
			out[key] = append(out[key], Value(c.Value, c.Negative, s.NamePattern))
		}
	}
	addLifespans(out, s.Life, lifespans)
	return out
}

// addLifespans sets, for each lifespan in key order, the window under its lead
// and the span under its Span, one value each, when the events give them.
func addLifespans(out subjectcompare.Values, events []LifeEvent, lifespans map[vocabkey.PropertyKey]subjectcompare.Lifespan) {
	for _, lead := range slices.Sorted(maps.Keys(lifespans)) {
		l := lifespans[lead]
		if window, span, ok := lifespan(events, l); ok {
			out[lead] = []propertycompare.Value{window}
			out[l.Span] = []propertycompare.Value{span}
		}
	}
}

// lifespan lays the dated events on one timeline. The span is the first to the
// last date. The window starts at the earliest Starts event, or MaxYears
// before the first date, and ends at the latest Ends event, or MaxYears after
// the last date. ok is false when no event is dated, when a shifted end
// cannot be placed, when the window or span would end before it begins, or
// when a date lies wholly outside the window: a side that contradicts itself
// has no lifespan.
func lifespan(events []LifeEvent, l subjectcompare.Lifespan) (window, span propertycompare.Value, ok bool) {
	// early and late are each date's first and last closed end; startEarly
	// and endLate are those of the events that fix the start and the end.
	var early, late, startEarly, endLate []valuecompare.DateEnd
	for _, ev := range events {
		er := subjectcompare.EventRole{EventType: ev.EventType, Role: ev.Role}
		for _, d := range ev.Dates {
			first, last, ok := valuecompare.DateEnds(d)
			if !ok {
				continue
			}
			if first.Open {
				first = last
			}
			if last.Open {
				last = first
			}
			early, late = append(early, first), append(late, last)
			if slices.Contains(l.Starts, er) {
				startEarly = append(startEarly, first)
			}
			if slices.Contains(l.Ends, er) {
				endLate = append(endLate, last)
			}
		}
	}
	spanFirst, ok := valuecompare.Earliest(early)
	if !ok {
		return window, span, false
	}
	spanLast, _ := valuecompare.Latest(late)
	start, ok := valuecompare.Earliest(startEarly)
	if !ok {
		if start, ok = valuecompare.ShiftYears(spanFirst, -l.MaxYears); !ok {
			return window, span, false
		}
	}
	end, ok := valuecompare.Latest(endLate)
	if !ok {
		if end, ok = valuecompare.ShiftYears(spanLast, l.MaxYears); !ok {
			return window, span, false
		}
	}
	for i := range early {
		if late[i].Hi < start.Lo || early[i].Lo > end.Hi {
			return window, span, false
		}
	}
	w, ok := valuecompare.EndsDate(start, end)
	if !ok {
		return window, span, false
	}
	sp, ok := valuecompare.EndsDate(spanFirst, spanLast)
	if !ok {
		return window, span, false
	}
	return propertycompare.Value{Value: valuecompare.Value{Date: &w}}, propertycompare.Value{Value: valuecompare.Value{Date: &sp}}, true
}

// Value is the one conversion from the reconciler's value to layer 2's. Type
// and Ref stay empty: layer 2 sets the type from the meta. Name and Date are
// shared with v.
func Value(v autoreconcile.Value, negative bool, namePattern string) propertycompare.Value {
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
