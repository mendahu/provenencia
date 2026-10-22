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
// give them, and the side's life (its first and last dates, or why it has no
// lifespan) is returned beside the values. The error is Reconcile's, and only
// Excluding runs it.
func Build(e Entity, metas subjectcompare.Metas, reconcile valuecompare.Rules, lifespans map[vocabkey.PropertyKey]subjectcompare.Lifespan, opts Options) (subjectcompare.Values, subjectcompare.LifeSide, error) {
	out := subjectcompare.Values{}
	for _, key := range slices.Sorted(maps.Keys(e.Properties)) {
		p := e.Properties[key]
		vals := p.Values
		if kept, touched := without(p.Candidates, opts.Excluding); touched {
			meta := metas[key]
			res, err := autoreconcile.Reconcile(reconcile, meta.ValueType, kept, nil, meta.CardinalityOrSingle())
			if err != nil {
				return nil, subjectcompare.LifeSide{}, err
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
	return out, addLifespans(out, e.Life, lifespans), nil
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
// The side's life is returned beside the values.
func ForSubject(s Subject, lifespans map[vocabkey.PropertyKey]subjectcompare.Lifespan) (subjectcompare.Values, subjectcompare.LifeSide) {
	out := subjectcompare.Values{}
	for key, cs := range s.Values {
		for _, c := range cs {
			v := Value(c.Value, c.Negative, s.NamePattern)
			v.Ref = c.ObservationID
			out[key] = append(out[key], v)
		}
	}
	return out, addLifespans(out, s.Life, lifespans)
}

// addLifespans sets, for each lifespan, the window under its lead
// and the span under its Span, one value each, when the events give them. It
// returns the side's life from the first lead in key order, the zero LifeSide
// when there is none.
func addLifespans(out subjectcompare.Values, events []LifeEvent, lifespans map[vocabkey.PropertyKey]subjectcompare.Lifespan) subjectcompare.LifeSide {
	var side subjectcompare.LifeSide
	first, _ := subjectcompare.FirstLifespanLead(lifespans)
	for lead, l := range lifespans {
		window, span, s := lifespan(events, l)
		if lead == first {
			side = s
		}
		if s.Missing == "" {
			out[lead] = []propertycompare.Value{window}
			out[l.Span] = []propertycompare.Value{span}
		}
	}
	return side
}

// lifespan lays the dated events on one timeline. The span is the first to the
// last date; an end no date closes (a lone BEF or AFT) stays open, so the span
// reads BEF or AFT there. The window takes its bounds from closed ends only.
// It starts at the earliest start of a Starts event when every such start is
// closed, else MaxYears before the earliest closed start of any date, else it
// is open; it ends at the latest end of an Ends event when every such end is
// closed, else MaxYears after the latest closed end of any date, else it is open. The side has a lifespan exactly when side.Missing is empty;
// its First and Last are then the dates as written that set the span's ends.
// When the open ends leave the span's two closed ends the wrong way round (a
// BEF date earlier than an AFT date), the span is read from the collapsed ends.
// A side with no dated event, or whose shifted end cannot be placed, is
// undated; one whose window or span would end before it begins, or with a date
// wholly outside the window (an open end, of a date or of the window, is never outside), contradicts itself and has no lifespan.
func lifespan(events []LifeEvent, l subjectcompare.Lifespan) (window, span propertycompare.Value, side subjectcompare.LifeSide) {
	// early and late are each date's first and last end with an open end
	// collapsed onto the other, dates the date each came from. They name the
	// dates that set First and Last and are the span's fallback.
	var early, late []valuecompare.DateEnd
	// rawEarly and rawLate are the same ends as written, an open one still
	// open; the span and the window are built from them. startEarly and
	// endLate are those of the events that fix the start and the end.
	var rawEarly, rawLate, startEarly, endLate []valuecompare.DateEnd
	var dates []datevalue.Value
	// startOpen and endOpen are set when some Starts record has an open first
	// end, or some Ends record an open last end: that side's records then
	// bound nothing on their own.
	var startOpen, endOpen bool
	for _, ev := range events {
		er := subjectcompare.EventRole{EventType: ev.EventType, Role: ev.Role}
		for _, d := range ev.Dates {
			first, last, ok := valuecompare.DateEnds(d)
			if !ok {
				continue
			}
			rawFirst, rawLast := first, last
			rawEarly, rawLate = append(rawEarly, rawFirst), append(rawLate, rawLast)
			if first.Open {
				first = last
			}
			if last.Open {
				last = first
			}
			early, late, dates = append(early, first), append(late, last), append(dates, d)
			if slices.Contains(l.Starts, er) {
				startEarly = append(startEarly, rawFirst)
				startOpen = startOpen || rawFirst.Open
			}
			if slices.Contains(l.Ends, er) {
				endLate = append(endLate, rawLast)
				endOpen = endOpen || rawLast.Open
			}
		}
	}
	undated := subjectcompare.LifeSide{Missing: subjectcompare.LifespanUndated}
	contradicts := subjectcompare.LifeSide{Missing: subjectcompare.LifespanContradicts}
	spanFirst, ok := valuecompare.Earliest(early)
	if !ok {
		return window, span, undated
	}
	spanLast, _ := valuecompare.Latest(late)
	firstClosed, lastClosed := spanFirst, spanLast
	// The window takes each bound from closed ends only: an open end never
	// pins it. A side no closed end bounds stays open.
	start, ok := valuecompare.Earliest(startEarly)
	if !ok || startOpen {
		if first, closed := valuecompare.Earliest(rawEarly); !closed {
			start = valuecompare.DateEnd{Open: true}
		} else if start, ok = valuecompare.ShiftYears(first, -l.MaxYears); !ok {
			return window, span, undated
		}
	}
	end, ok := valuecompare.Latest(endLate)
	if !ok || endOpen {
		if last, closed := valuecompare.Latest(rawLate); !closed {
			end = valuecompare.DateEnd{Open: true}
		} else if end, ok = valuecompare.ShiftYears(last, l.MaxYears); !ok {
			return window, span, undated
		}
	}
	for i := range rawEarly {
		if (!start.Open && !rawLate[i].Open && rawLate[i].Hi < start.Lo) || (!end.Open && !rawEarly[i].Open && rawEarly[i].Lo > end.Hi) {
			return window, span, contradicts
		}
	}
	w, ok := valuecompare.EndsDate(start, end)
	if !ok {
		return window, span, contradicts
	}
	// An end no date closes stays open: the span is BEF or AFT there, so a
	// closed date on the other side still bounds it.
	firstEnd, ok := valuecompare.Earliest(rawEarly)
	if !ok {
		firstEnd = valuecompare.DateEnd{Open: true}
	}
	lastEnd, ok := valuecompare.Latest(rawLate)
	if !ok {
		lastEnd = valuecompare.DateEnd{Open: true}
	}
	// Two closed ends the wrong way round (a BEF date earlier than an AFT
	// date) are no contradiction: the span falls back to the collapsed ends.
	if !firstEnd.Open && !lastEnd.Open && firstEnd.Lo > lastEnd.Hi {
		firstEnd, lastEnd = firstClosed, lastClosed
	}
	sp, ok := valuecompare.EndsDate(firstEnd, lastEnd)
	if !ok {
		return window, span, contradicts
	}
	firstAt := slices.Index(early, firstClosed)
	lastAt := slices.Index(late, lastClosed)
	side.First, side.Last = &dates[firstAt], &dates[lastAt]
	return propertycompare.Value{Value: valuecompare.Value{Date: &w}}, propertycompare.Value{Value: valuecompare.Value{Date: &sp}}, side
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
