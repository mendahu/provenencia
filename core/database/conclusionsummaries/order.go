package conclusionsummaries

import (
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/valuecompare"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// titleCollator is built once. Constructing it sorts the whole collation
// table, which is most of a list's time when every call makes a new one.
var titleCollator = collate.New(language.Und, collate.IgnoreCase, collate.IgnoreDiacritics)

// SortByTitle orders any titled row by the title the caller supplies (a Place
// row by its toponym, else its working label; a Source group by its title; a
// Person row by its sort title, which leaves honorifics out), ignoring case and diacritics (Montréal beside Montreal), by the
// Unicode root collation so the order is the same in every locale. A row
// titled only by its ref sorts last, by ref. Ties break by ref. This is the
// one owner of title order; the app shows rows as given.
func SortByTitle[T any](rows []T, title func(T) string, ref func(T) string) {
	sortByKey(rows, func(row T) titleKey {
		return titleKey{title: strings.TrimSpace(title(row)), ref: strings.ToLower(ref(row))}
	}, titleKeyLess)
}

// titleKey is a row's title and ref, each prepared once: the trimmed title and
// the lowercased ref.
type titleKey struct{ title, ref string }

// titleKeyLess is title order: a row with no title last, then the root
// collation of the titles, then the refs.
func titleKeyLess(a, b titleKey) bool {
	if (a.title == "") != (b.title == "") {
		return b.title == ""
	}
	if a.title != "" {
		if o := titleCollator.CompareString(a.title, b.title); o != 0 {
			return o < 0
		}
	}
	return a.ref < b.ref
}

// SortByBirth orders people by birth date: dated first, earliest first, then by
// ref; the one owner of birth order. Each key is computed once, and rows with
// equal keys keep their order.
func SortByBirth[T any](rows []T, summary func(T) PersonSummary) {
	sortByDateKey(rows, func(row T) dateKey { return birthKeyOf(summary(row)) })
}

// PlacedBirth is a birth date placed once, to age many events by.
type PlacedBirth struct{ placed placedDate }

// PlaceBirth places a person's birth date; ok is false when it has no place.
func PlaceBirth(birth *datevalue.Value) (PlacedBirth, bool) {
	p, ok := datePlace(birth)
	return PlacedBirth{placed: p}, ok
}

// EventPlacing is where an event sits on a line of years and the age a person
// had at it: EventSpanAndAge's result. Span and Days are set when Dated; Age is
// set when Aged.
type EventPlacing struct {
	Span  TimelineSpan
	Days  valuecompare.Window
	Dated bool
	Age   AgeRange
	Aged  bool
}

// EventSpanAndAge is where h sits on a line of years and the age a person born
// on birth had at it, from one placing of the event's date. The span is that of
// the date the event order places it by (a date, else a start date, whichever
// has a place); Dated is false when h is undated, so "dated" and "has a span"
// are one fact. Days is the days of the same date as written (an ABT date is
// not widened; a BEF or AFT end is open on its side), the zero Window when h is
// not dated. Age is in whole years, as a range: Hi counts from the
// earliest birth day to the latest day of the event, Lo from the latest birth
// day to the earliest day of the event. Aged is false when hasBirth is false,
// when the event or the birth has no place, when any of the four ends is open
// (BEF, AFT, a one-sided range), or when the event is before the birth.
func EventSpanAndAge(h EventSummary, birth PlacedBirth, hasBirth bool) EventPlacing {
	ev := eventDating(h.Date, h.StartDate)
	placing := EventPlacing{Span: ev.span, Dated: ev.key.ok}
	if placing.Dated {
		placing.Days = valuecompare.EndsWindow(ev.first, ev.last)
	}
	if !hasBirth || !placing.Dated {
		return placing
	}
	b := birth.placed
	// The youngest is from the latest birth to the earliest event day; the
	// oldest from the earliest birth to the latest event day. A side is open
	// when either date leaves its end open.
	age := AgeRange{OpenLo: b.last.Open || ev.first.Open, OpenHi: b.first.Open || ev.last.Open}
	if age.OpenLo && age.OpenHi {
		return placing
	}
	if !age.OpenHi {
		if age.Hi = datevalue.CompletedYears(b.first.Lo, ev.last.Hi); age.Hi < 0 {
			return placing
		}
	}
	if !age.OpenLo {
		age.Lo = max(0, datevalue.CompletedYears(b.last.Hi, ev.first.Lo))
	}
	placing.Age, placing.Aged = age, true
	return placing
}

// sortByDateKey orders rows by their date key: dated first, earliest first,
// then by ref. Each key is computed once, and rows with equal keys keep their
// order.
func sortByDateKey[T any](rows []T, key func(T) dateKey) {
	sortByKey(rows, key, dateKeyLess)
}

// sortByKey computes each row's key once, sorts stably by less, and writes the
// rows back. Rows with equal keys keep their order.
func sortByKey[T, K any](rows []T, key func(T) K, less func(a, b K) bool) {
	type keyed struct {
		row T
		key K
	}
	keyedRows := make([]keyed, len(rows))
	for i := range rows {
		keyedRows[i] = keyed{row: rows[i], key: key(rows[i])}
	}
	sort.SliceStable(keyedRows, func(i, j int) bool {
		return less(keyedRows[i].key, keyedRows[j].key)
	})
	for i := range rows {
		rows[i] = keyedRows[i].row
	}
}

// sortPersons is the Persons list order.
func sortPersons(rows []PersonSummary, profile nameformats.Profile) {
	SortByTitle(rows, func(h PersonSummary) string { return personSortKey(h, profile) }, func(h PersonSummary) string { return h.Entity.Ref })
}

// sortPlaces is the Places list order.
func sortPlaces(rows []PlaceSummary) {
	SortByTitle(rows, placeTitle, func(h PlaceSummary) string { return h.Entity.Ref })
}

// personSortKey is the Persons list order. A name with parts sorts by the
// profile's sorted order, undecorated (the row may add quotes and parentheses),
// without honorific prefixes: see nameformats.Profile.SortKey. A catalog with no
// default profile joins those parts in stored order. form is the transcription
// and is not a sort input: a name with no parts sorts by its working label.
func personSortKey(h PersonSummary, profile nameformats.Profile) string {
	if h.Name != nil {
		if key := profile.SortKey(*h.Name); key != "" {
			return key
		}
	}
	return strings.TrimSpace(h.Entity.Label)
}

func placeTitle(h PlaceSummary) string {
	if len(h.Names) > 0 {
		return h.Names[0]
	}
	return h.Entity.Label
}
