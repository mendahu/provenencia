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
// negative as recorded, in order. Each value's Ref is its Observation id.
// lifespans is the registry's Subject.Lifespans for the side's kind; for each,
// a window and a span are added when the life events give them, with no Ref.
func ForSubject(s Subject, lifespans map[vocabkey.PropertyKey]subjectcompare.Lifespan) subjectcompare.Values {
	out := subjectcompare.Values{}
	for key, cs := range s.Values {
		for _, c := range cs {
			v := Value(c.Value, c.Negative, s.NamePattern)
			v.Ref = c.ObservationID
			out[key] = append(out[key], v)
		}
	}
	addLifespans(out, s.Life, lifespans)
	return out
}

// addLifespans sets, for each lifespan, the window under its lead
// and the span under its Span, one value each, when the events give them.
func addLifespans(out subjectcompare.Values, events []LifeEvent, lifespans map[vocabkey.PropertyKey]subjectcompare.Lifespan) {
	for lead, l := range lifespans {
		if window, span, ok := lifespan(events, l); ok {
			out[lead] = []propertycompare.Value{window}
			out[l.Span] = []propertycompare.Value{span}
		}
	}
}

// lifespan lays the exact dated events on one timeline. A date with an open
// end (BEF, AFT, a half-open range) is skipped: it is too vague to rule
// anything out. The span is the earliest first end to the latest last end of
// the counted dates. The window starts at the earliest first end of a Starts
// event, else MaxYears before the span's first end, and ends at the latest
// last end of an Ends event, else MaxYears after the span's last end; it is
// then widened by SlackYears each way, so only an egregious case conflicts.
// A side with no counted date, or whose shifted end cannot be placed, is undated; one
// whose window would end before it begins, or with a counted date wholly
// outside the window, contradicts itself and has no lifespan.
func lifespan(events []LifeEvent, l subjectcompare.Lifespan) (window, span propertycompare.Value, ok bool) {
	var firsts, lasts, startFirsts, endLasts []valuecompare.DateEnd
	for _, ev := range events {
		er := subjectcompare.EventRole{EventType: ev.EventType, Role: ev.Role}
		for _, d := range ev.Dates {
			first, last, ok := valuecompare.DateEnds(d)
			if !ok || first.Open || last.Open {
				continue
			}
			firsts, lasts = append(firsts, first), append(lasts, last)
			if slices.Contains(l.Starts, er) {
				startFirsts = append(startFirsts, first)
			}
			if slices.Contains(l.Ends, er) {
				endLasts = append(endLasts, last)
			}
		}
	}
	spanFirst, ok := valuecompare.Earliest(firsts)
	if !ok {
		return window, span, false
	}
	spanLast, _ := valuecompare.Latest(lasts)
	start, ok := valuecompare.Earliest(startFirsts)
	if !ok {
		start = spanFirst
		if start, ok = valuecompare.ShiftYears(start, -l.MaxYears); !ok {
			return window, span, false
		}
	}
	end, ok := valuecompare.Latest(endLasts)
	if !ok {
		if end, ok = valuecompare.ShiftYears(spanLast, l.MaxYears); !ok {
			return window, span, false
		}
	}
	start, okStart := valuecompare.ShiftYears(start, -l.SlackYears)
	end, okEnd := valuecompare.ShiftYears(end, l.SlackYears)
	if !okStart || !okEnd {
		return window, span, false
	}
	for i := range firsts {
		if lasts[i].Hi < start.Lo || firsts[i].Lo > end.Hi {
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
// and Ref stay empty: layer 2 sets the type from the meta. Name, Date and Depiction
// are shared with v.
func Value(v autoreconcile.Value, negative bool, namePattern string) propertycompare.Value {
	out := propertycompare.Value{
		Value: valuecompare.Value{
			Text:      v.Text,
			Term:      v.TermKey,
			Integer:   v.Integer,
			Name:      v.Name,
			Date:      v.Date,
			Depiction: v.Depiction,
		},
		Negative: negative,
		Neutral:  v.Neutral,
	}
	if v.Name != nil {
		out.NamePattern = namePattern
	}
	return out
}
