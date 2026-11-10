package persondetail

import (
	"bytes"

	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// lifelineTypes names the event-and-role pairs that may sit outside the years
// of a life: a person's own burial or funeral comes after the death. It orders
// the rows of a timeline the dates cannot order (order) and places the birth
// and death markers when the records name no event for them (markers); it
// judges no event, and flags none. afterDeath lists its entries in the order
// the events happen when their dates overlap (a funeral before a burial): the
// position is a rank. beforeBirth's order means nothing.
type lifelineTypes struct{ afterDeath, beforeBirth []subjectcompare.EventRole }

var lifeline = lifelineTypes{
	afterDeath: []subjectcompare.EventRole{
		{EventType: vocabkey.EventFuneral, Role: vocabkey.RoleSubject},
		{EventType: vocabkey.EventBurial, Role: vocabkey.RoleSubject},
	},
}

// MarkerSide is which side of its row a marker sits on.
type MarkerSide int

const (
	// SideAfter puts the marker just after its row.
	SideAfter MarkerSide = iota + 1
	// SideBefore puts the marker just before its row.
	SideBefore
)

// MarkerPlace is the one dated row a marker sits next to, and on which side.
type MarkerPlace struct {
	Row  []byte
	Side MarkerSide
}

// LifeMarker is where a timeline puts a birth or a death. Event is the
// recorded event, nil when the records name none (a placeholder). Place is set
// whenever the marker has no span of its own and a row is dated, to the dated
// event it sits next to; no date is invented. It is nil when the marker's
// event has a span, or when no row is dated.
type LifeMarker struct {
	Event []byte
	Place *MarkerPlace
}

// rank is the index in list of the first entry row, of type t, makes: the type
// and one of the person's roles on the row make an entry. A role matches when
// it is the product's own term with that key. It is -1 when the row makes none.
func (l lifelineTypes) rank(list []subjectcompare.EventRole, row PersonEvent, t vocabkey.EventType) int {
	for i, x := range list {
		if x.EventType != t {
			continue
		}
		for j := range row.Roles {
			if row.Roles[j].ProductKey() == string(x.Role) {
				return i
			}
		}
	}
	return -1
}

// lifelineRow is a row with its days, read once.
type lifelineRow struct {
	row    PersonEvent
	window valuecompare.Window
	dated  bool
	// after is the row's afterDeath rank, -1 for none.
	after int
}

// timeline is the Timeline's rows and markers: events, which are in date order
// (undated last), put in the Timeline's order by order, then the markers on
// that order. It is the one place the two run, in that order.
func (l lifelineTypes) timeline(events []PersonEvent, types func(event []byte) vocabkey.EventType, days func(event []byte) (valuecompare.Window, bool), birth, death []byte) Timeline {
	t := Timeline{Events: l.order(events, types, days, birth, death)}
	t.Birth, t.Death = l.markers(t.Events, types, birth, death)
	return t
}

// order is the Timeline's order. It starts from events in date order (undated
// last) and moves a row only where the dates cannot say which came first, and
// changes no row's value. days gives a dated row's days, as written. A row with
// no days never moves and never makes another move. birth and death are the
// recorded events (nil for none). In this order, on one slice:
//
//  1. When the recorded birth has days, a row before it whose days overlap the
//     birth's, and that may not come before birth, goes just after the birth.
//  2. When the recorded death has days, a burial or funeral of the person's own
//     (afterDeath) before it whose days overlap the death's goes just after it.
//  3. Of the afterDeath rows with days placed after the death (all of them when
//     the death is nil or has no days), a row that has a later one of a lesser
//     rank whose days overlap its own goes just after the last such one: a
//     burial after a funeral. Rows still before the death (a clear conflict)
//     keep date order, and so do all of them when there are more than
//     maxReorderedAfterDeath.
//
// A row placed before a date whose days do not overlap it ends before that date
// begins, so "overlap" is exactly "the dates cannot say which came first".
func (l lifelineTypes) order(events []PersonEvent, types func(event []byte) vocabkey.EventType, days func(event []byte) (valuecompare.Window, bool), birth, death []byte) []PersonEvent {
	rows := make([]lifelineRow, len(events))
	for i, e := range events {
		w, ok := days(e.Event)
		rows[i] = lifelineRow{row: e, window: w, dated: ok, after: l.rank(l.afterDeath, e, types(e.Event))}
	}
	find := func(id []byte) int {
		if id == nil {
			return -1
		}
		for i := range rows {
			if bytes.Equal(rows[i].row.Event, id) {
				return i
			}
		}
		return -1
	}
	if i := find(birth); i >= 0 && rows[i].dated {
		rows = moveAfterOverlapping(rows, i, func(r lifelineRow) bool {
			return l.rank(l.beforeBirth, r.row, types(r.row.Event)) < 0
		})
	}
	deathAt := find(death)
	if deathAt >= 0 && rows[deathAt].dated {
		rows = moveAfterOverlapping(rows, deathAt, func(r lifelineRow) bool { return r.after >= 0 })
		deathAt = find(death)
	} else {
		deathAt = -1
	}
	rows = reorderAfterDeath(rows, deathAt+1)
	out := make([]PersonEvent, len(rows))
	for i := range rows {
		out[i] = rows[i].row
	}
	return out
}

// maxReorderedAfterDeath is the most afterDeath rows pass (3) of order reorders.
//
// decided: a real person has one to three burials and funerals; only a crafted
// catalog passes 64. Above it pass (3) leaves those rows in date order, so a
// crafted person costs the two linear passes and a count.
const maxReorderedAfterDeath = 64

// reorderAfterDeath is pass (3) of order on rows[from:]: each dated afterDeath
// row with a later one of a lesser rank whose days overlap its own goes just
// after the last such one. It walks back from the end, so rows that move to one
// place keep their order. It does nothing when it would consider more than
// maxReorderedAfterDeath rows.
func reorderAfterDeath(rows []lifelineRow, from int) []lifelineRow {
	n := 0
	for i := from; i < len(rows); i++ {
		if rows[i].dated && rows[i].after >= 0 {
			n++
		}
	}
	if n < 2 || n > maxReorderedAfterDeath {
		return rows
	}
	for i := len(rows) - 1; i >= from; i-- {
		r := rows[i]
		if !r.dated || r.after < 0 {
			continue
		}
		last := -1
		for j := i + 1; j < len(rows); j++ {
			if rows[j].dated && rows[j].after >= 0 && rows[j].after < r.after && rows[j].window.Overlaps(r.window) {
				last = j
			}
		}
		if last >= 0 {
			moveRowAfter(rows, i, last)
		}
	}
	return rows
}

// moveRowAfter moves rows[from] to just after rows[to], to > from, in place.
func moveRowAfter(rows []lifelineRow, from, to int) {
	moved := rows[from]
	copy(rows[from:to], rows[from+1:to+1])
	rows[to] = moved
}

// moveAfterOverlapping moves each dated row before rows[anchor] whose days
// overlap the anchor's and for which moves is true to just after the anchor,
// the moved rows in their order and ahead of the rows already after it. The
// other rows keep their order. It rebuilds the slice once.
func moveAfterOverlapping(rows []lifelineRow, anchor int, moves func(lifelineRow) bool) []lifelineRow {
	out := make([]lifelineRow, 0, len(rows))
	var moved []lifelineRow
	for _, r := range rows[:anchor] {
		if r.dated && r.window.Overlaps(rows[anchor].window) && moves(r) {
			moved = append(moved, r)
			continue
		}
		out = append(out, r)
	}
	out = append(out, rows[anchor])
	out = append(out, moved...)
	return append(out, rows[anchor+1:]...)
}

// markers places the birth and death markers on events, which are in the
// Timeline's order (date order, then order's moves). birth and death are the
// recorded events (nil for none). A marker whose event has a span needs no
// Place. Otherwise a birth sits after the last dated row that may come before
// birth when there is one, else before the first dated row; a death sits before
// the first of the dated rows that may come after death, in the Timeline's
// order, when there is one, else after the last dated row. Place is nil when no
// row is dated. types gives an event's type.
func (l lifelineTypes) markers(events []PersonEvent, types func(event []byte) vocabkey.EventType, birth, death []byte) (Birth, Death LifeMarker) {
	hasSpan := func(id []byte) bool {
		for _, e := range events {
			if bytes.Equal(e.Event, id) {
				return e.Span != nil
			}
		}
		return false
	}
	Birth, Death = LifeMarker{Event: birth}, LifeMarker{Event: death}
	if birth == nil || !hasSpan(birth) {
		var first, lastBefore []byte
		for _, e := range events {
			if e.Span == nil {
				continue
			}
			if first == nil {
				first = e.Event
			}
			if l.rank(l.beforeBirth, e, types(e.Event)) >= 0 {
				lastBefore = e.Event
			}
		}
		switch {
		case lastBefore != nil:
			Birth.Place = &MarkerPlace{Row: lastBefore, Side: SideAfter}
		case first != nil:
			Birth.Place = &MarkerPlace{Row: first, Side: SideBefore}
		}
	}
	if death == nil || !hasSpan(death) {
		var last, firstAfter []byte
		for _, e := range events {
			if e.Span == nil {
				continue
			}
			last = e.Event
			if firstAfter == nil && l.rank(l.afterDeath, e, types(e.Event)) >= 0 {
				firstAfter = e.Event
			}
		}
		switch {
		case firstAfter != nil:
			Death.Place = &MarkerPlace{Row: firstAfter, Side: SideBefore}
		case last != nil:
			Death.Place = &MarkerPlace{Row: last, Side: SideAfter}
		}
	}
	return Birth, Death
}
