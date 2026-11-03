package conclusionsummaries

import (
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/datevalue"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// titleCollator is built once. Constructing it sorts the whole collation
// table, which is most of a list's time when every call makes a new one.
var titleCollator = collate.New(language.Und, collate.IgnoreCase, collate.IgnoreDiacritics)

// SortByTitle orders any titled row by the title it shows (a Person or Place
// row by its name or toponym, else its working label; a Source group by its
// title), ignoring case and diacritics (Montréal beside Montreal), by the
// Unicode root collation so the order is the same in every locale. A row
// titled only by its ref sorts last, by ref. Ties break by ref. This is the
// one owner of title order; the app shows rows as given.
func SortByTitle[T any](rows []T, title func(T) string, ref func(T) string) {
	sort.SliceStable(rows, func(i, j int) bool {
		ti, tj := strings.TrimSpace(title(rows[i])), strings.TrimSpace(title(rows[j]))
		if (ti == "") != (tj == "") {
			return tj == ""
		}
		if ti != "" {
			if o := titleCollator.CompareString(ti, tj); o != 0 {
				return o < 0
			}
		}
		return strings.ToLower(ref(rows[i])) < strings.ToLower(ref(rows[j]))
	})
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

// EventSpanAndAge is where h sits on a line of years and the age a person born
// on birth had at it, from one placing of the event's date. The span is that of
// the date the event order places it by (a date, else a start date, whichever
// has a place); dated is false when h is undated, so "dated" and "has a span"
// are one fact. The age is in whole years, as a range: Hi counts from the
// earliest birth day to the latest day of the event, Lo from the latest birth
// day to the earliest day of the event. aged is false when hasBirth is false,
// when the event or the birth has no place, when any of the four ends is open
// (BEF, AFT, a one-sided range), or when the event is before the birth.
func EventSpanAndAge(h EventSummary, birth PlacedBirth, hasBirth bool) (span TimelineSpan, dated bool, age AgeRange, aged bool) {
	ev := eventDating(h.Date, h.StartDate)
	span, dated = ev.span, ev.key.ok
	if !hasBirth {
		return span, dated, AgeRange{}, false
	}
	b := birth.placed
	if !dated || ev.first.Open || ev.last.Open || b.first.Open || b.last.Open {
		return span, dated, AgeRange{}, false
	}
	hi := datevalue.CompletedYears(b.first.Lo, ev.last.Hi)
	if hi < 0 {
		return span, dated, AgeRange{}, false
	}
	return span, dated, AgeRange{Lo: max(0, datevalue.CompletedYears(b.last.Hi, ev.first.Lo)), Hi: hi}, true
}

// sortByDateKey orders rows by their date key: dated first, earliest first,
// then by ref. Each key is computed once, and rows with equal keys keep their
// order.
func sortByDateKey[T any](rows []T, key func(T) dateKey) {
	type keyed struct {
		row T
		key dateKey
	}
	keyedRows := make([]keyed, len(rows))
	for i := range rows {
		keyedRows[i] = keyed{row: rows[i], key: key(rows[i])}
	}
	sort.SliceStable(keyedRows, func(i, j int) bool {
		return dateKeyLess(keyedRows[i].key, keyedRows[j].key)
	})
	for i := range rows {
		rows[i] = keyedRows[i].row
	}
}

// personSortKey is the People list order. A name with parts sorts by the
// profile's sorted order, undecorated; the row may add quotes and parentheses. A catalog with no
// default profile joins those parts in stored order. form is the transcription
// and is not a sort input: a name with no parts sorts by its working label.
func personSortKey(h PersonSummary, order []nameformats.PartOrder) string {
	if h.Name != nil && len(h.Name.Parts) > 0 {
		if key := strings.TrimSpace(nameformats.Format(*h.Name, order)); key != "" {
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
