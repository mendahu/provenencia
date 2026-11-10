package persondetail

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/databasetest"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func markerBefore(row []byte) *MarkerPlace { return &MarkerPlace{Row: row, Side: SideBefore} }
func markerAfter(row []byte) *MarkerPlace  { return &MarkerPlace{Row: row, Side: SideAfter} }
func sameMarkerPlace(a, b *MarkerPlace) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Side == b.Side && bytes.Equal(a.Row, b.Row)
}

// id is the event id of lifelineTestRow(n).
func id(n byte) []byte { return []byte{n} }

// lifelineTestTypes gives the type of the event id(n) that lifelineTestRow(n)
// makes: 1 birth, 2 census, 3 funeral, 4 and 7 burial, 5 marriage, 6 baptism,
// 8 death.
var lifelineTestTypes = map[byte]vocabkey.EventType{
	1: vocabkey.EventBirth, 2: "census", 3: vocabkey.EventFuneral, 4: vocabkey.EventBurial, 5: vocabkey.EventMarriage,
	6: "baptism", 7: vocabkey.EventBurial, 8: vocabkey.EventDeath,
}

func types(e []byte) vocabkey.EventType { return lifelineTestTypes[e[0]] }

// lifelineTestRow is the person's row for event id(n): the person's roles on
// it, the subject when none are given.
func lifelineTestRow(n byte, roles ...vocabkey.Role) PersonEvent {
	r := PersonEvent{Event: id(n)}
	if len(roles) == 0 {
		roles = []vocabkey.Role{vocabkey.RoleSubject}
	}
	for _, role := range roles {
		r.Roles = append(r.Roles, conclusionsummaries.Term{Key: string(role), Origin: vocabkey.OriginProvenencia})
	}
	return r
}

func TestLifeMarkers(t *testing.T) {
	span := &conclusionsummaries.TimelineSpan{From: 1800, To: 1801}
	row := func(n byte, dated bool, roles ...vocabkey.Role) PersonEvent {
		r := lifelineTestRow(n, roles...)
		if dated {
			r.Span = span
		}
		return r
	}
	before := func(n byte) *MarkerPlace { return markerBefore(id(n)) }
	after := func(n byte) *MarkerPlace { return markerAfter(id(n)) }
	same := sameMarkerPlace
	cases := []struct {
		name           string
		types          lifelineTypes
		rows           []PersonEvent
		birth, death   []byte
		birthP, deathP *MarkerPlace
	}{
		{"a witness at a burial", lifeline, []PersonEvent{row(2, true), row(4, true, vocabkey.RoleWitness)}, nil, nil, before(2), after(4)},
		{"subject and witness on one burial", lifeline, []PersonEvent{row(2, true), row(4, true, vocabkey.RoleWitness, vocabkey.RoleSubject)}, nil, nil, before(2), before(4)},
		{"only a burial dated", lifeline, []PersonEvent{row(4, true)}, nil, nil, before(4), before(4)},
		{"a baptism and a burial", lifeline, []PersonEvent{row(6, true), row(4, true)}, nil, nil, before(6), before(4)},
		{"two burials", lifeline, []PersonEvent{row(4, true), row(7, true)}, nil, nil, before(4), before(4)},
		{"a funeral after a census", lifeline, []PersonEvent{row(2, true), row(3, true)}, nil, nil, before(2), before(3)},
		{"no burial or funeral dated", lifeline, []PersonEvent{row(2, true), row(5, true)}, nil, nil, before(2), after(5)},
		{"no dated row", lifeline, []PersonEvent{row(2, false)}, nil, nil, nil, nil},
		{"a recorded birth with no span", lifeline, []PersonEvent{row(1, false), row(2, true), row(5, true)}, id(1), nil, before(2), after(5)},
		{"a recorded birth with a span", lifeline, []PersonEvent{row(1, true), row(2, true)}, id(1), nil, nil, after(2)},
		{"a recorded death with no span and a dated funeral", lifeline, []PersonEvent{row(2, true), row(3, true), row(5, false)}, nil, id(5), before(2), before(3)},
		{"a recorded death with a span", lifeline, []PersonEvent{row(2, true), row(5, true)}, nil, id(5), before(2), nil},
		{"a type that may come before birth", lifelineTypes{beforeBirth: []subjectcompare.EventRole{{EventType: "census", Role: vocabkey.RoleSubject}}},
			[]PersonEvent{row(2, true), row(5, true)}, nil, nil, after(2), after(5)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, d := tc.types.markers(tc.rows, types, tc.birth, tc.death)
			if !bytes.Equal(b.Event, tc.birth) || !bytes.Equal(d.Event, tc.death) {
				t.Errorf("events %v %v", b.Event, d.Event)
			}
			if !same(b.Place, tc.birthP) {
				t.Errorf("birth place %+v, want %+v", b.Place, tc.birthP)
			}
			if !same(d.Place, tc.deathP) {
				t.Errorf("death place %+v, want %+v", d.Place, tc.deathP)
			}
		})
	}
}

func TestLifelineOrder(t *testing.T) {
	type in struct {
		n     byte
		d     *datevalue.Value
		roles []vocabkey.Role
	}
	witness := []vocabkey.Role{vocabkey.RoleWitness}
	death := datevaluetest.YMD(2011, 3, 29)
	birthDay := datevaluetest.YMD(1850, 3, 12)
	beforeBirth := lifelineTypes{beforeBirth: []subjectcompare.EventRole{{EventType: "census", Role: vocabkey.RoleSubject}}}
	cases := []struct {
		name         string
		types        lifelineTypes
		rows         []in
		birth, death []byte
		want         []byte
	}{
		{"the dogfood case", lifeline, []in{{4, datevaluetest.About(2011, 0, 0), nil}, {8, death, nil}, {3, datevaluetest.After(2011, 3, 29), nil}}, nil, id(8), []byte{8, 3, 4}},
		{"a burial the dates put before the death", lifeline, []in{{4, datevaluetest.YMD(1990, 0, 0), nil}, {8, death, nil}}, nil, id(8), []byte{4, 8}},
		{"and a later funeral", lifeline, []in{{4, datevaluetest.YMD(1990, 0, 0), nil}, {8, death, nil}, {3, datevaluetest.After(2011, 3, 29), nil}}, nil, id(8), []byte{4, 8, 3}},
		{"a witness's burial", lifeline, []in{{4, datevaluetest.About(2011, 0, 0), witness}, {8, death, nil}}, nil, id(8), []byte{4, 8}},
		{"funeral and burial overlap after the death", lifeline, []in{{8, death, nil}, {4, datevaluetest.YMD(2011, 4, 0), nil}, {3, datevaluetest.YMD(2011, 4, 5), nil}}, nil, id(8), []byte{8, 3, 4}},
		{"funeral and burial that do not overlap", lifeline, []in{{8, death, nil}, {4, datevaluetest.YMD(2011, 4, 2), nil}, {3, datevaluetest.YMD(2011, 4, 10), nil}}, nil, id(8), []byte{8, 4, 3}},
		{"a funeral does not pass a row the dates put before it", lifeline, []in{{4, datevaluetest.YMD(2011, 0, 0), nil}, {2, datevaluetest.YMD(2011, 2, 0), nil}, {3, datevaluetest.After(2011, 3, 1), nil}}, nil, nil, []byte{2, 3, 4}},
		{"a burial the dates put before the birth does not pass it", lifeline, []in{{4, datevaluetest.Before(2010, 6, 1), nil}, {3, datevaluetest.Before(2010, 12, 1), nil}, {1, datevaluetest.YMD(2010, 6, 15), nil}}, id(1), nil, []byte{4, 1, 3}},
		{"no death", lifeline, []in{{4, datevaluetest.About(2011, 0, 0), nil}, {3, datevaluetest.After(2011, 3, 29), nil}}, nil, nil, []byte{3, 4}},
		{"a burial and a funeral both before the death", lifeline, []in{{4, datevaluetest.YMD(1990, 0, 0), nil}, {3, datevaluetest.YMD(1990, 5, 0), nil}, {8, death, nil}}, nil, id(8), []byte{4, 3, 8}},
		{"a recorded death with no days", lifeline, []in{{4, datevaluetest.About(2011, 0, 0), nil}, {3, datevaluetest.After(2011, 3, 29), nil}, {8, nil, nil}}, nil, id(8), []byte{3, 4, 8}},
		{"an undated burial", lifeline, []in{{8, death, nil}, {4, nil, nil}}, nil, id(8), []byte{8, 4}},
		{"a baptism in the birth's year", lifeline, []in{{6, datevaluetest.YMD(1850, 0, 0), nil}, {1, birthDay, nil}}, id(1), nil, []byte{1, 6}},
		{"a baptism the year before", lifeline, []in{{6, datevaluetest.YMD(1849, 0, 0), nil}, {1, birthDay, nil}}, id(1), nil, []byte{6, 1}},
		{"an infant", lifeline, []in{{6, datevaluetest.YMD(1850, 0, 0), nil}, {4, datevaluetest.YMD(1850, 0, 0), nil}, {1, birthDay, nil}, {8, datevaluetest.YMD(1850, 9, 1), nil}}, id(1), id(8), []byte{1, 6, 8, 4}},
		{"a burial the death would carry past a birth it does not overlap", lifeline, []in{{4, datevaluetest.Before(2011, 1, 1), nil}, {1, datevaluetest.YMD(2011, 6, 15), nil}, {8, datevaluetest.Before(2011, 12, 31), nil}}, id(1), id(8), []byte{4, 1, 8}},
		{"a type that may come before birth", beforeBirth, []in{{2, datevaluetest.YMD(1850, 0, 0), nil}, {6, datevaluetest.YMD(1850, 0, 0), nil}, {1, birthDay, nil}}, id(1), nil, []byte{2, 1, 6}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var rows []PersonEvent
			days := map[byte]valuecompare.Window{}
			for _, r := range tc.rows {
				pe := lifelineTestRow(r.n, r.roles...)
				if r.d != nil {
					pe.Span = &conclusionsummaries.TimelineSpan{From: 1800, To: 1801}
					pe.Age = &conclusionsummaries.AgeRange{Lo: int(r.n)}
					w, ok := valuecompare.DateWindow(*r.d)
					if !ok {
						t.Fatalf("row %d has no days", r.n)
					}
					days[r.n] = w
				}
				rows = append(rows, pe)
			}
			daysOf := func(e []byte) (valuecompare.Window, bool) { w, ok := days[e[0]]; return w, ok }
			input := append([]PersonEvent(nil), rows...)
			got := tc.types.order(rows, types, daysOf, tc.birth, tc.death)
			var order []byte
			for _, r := range got {
				order = append(order, r.Event[0])
			}
			if !bytes.Equal(order, tc.want) {
				t.Fatalf("order %v, want %v", order, tc.want)
			}
			if len(got) != len(input) {
				t.Fatalf("%d rows, want %d", len(got), len(input))
			}
			for _, r := range got {
				for _, in := range input {
					if bytes.Equal(in.Event, r.Event) && !reflect.DeepEqual(in, r) {
						t.Errorf("row %d changed: %+v, was %+v", r.Event[0], r, in)
					}
				}
			}
		})
	}
}

// TestLifelineOrderMovesOnlyOverlappingRows runs order over random years 2010
// and 2011, with a dated birth, a dated death, both or neither, and holds it to what
// it promises: the output is the input rows, and a pair of rows whose
// order changed has overlapping days.
func TestLifelineOrderMovesOnlyOverlappingRows(t *testing.T) {
	rng := rand.New(rand.NewSource(16))
	kinds := []byte{2, 3, 4, 5, 6, 7}
	date := func() *datevalue.Value {
		y, m, d := 2010+rng.Intn(2), rng.Intn(13), 0
		if m > 0 {
			d = rng.Intn(29)
		}
		switch rng.Intn(5) {
		case 0:
			return datevaluetest.After(y, m, d)
		case 1:
			return datevaluetest.Before(y, m, d)
		default:
			return datevaluetest.YMD(y, m, d)
		}
	}
	for c := 0; c < 5000; c++ {
		var ns []byte
		days := map[byte]valuecompare.Window{}
		add := func(n byte) {
			w, ok := valuecompare.DateWindow(*date())
			if !ok {
				t.Fatalf("case %d: row %d has no days", c, n)
			}
			days[n] = w
			ns = append(ns, n)
		}
		var birth, death []byte
		for i, n := 0, 2+rng.Intn(7); i < n; i++ {
			add(kinds[rng.Intn(len(kinds))] + byte(i)*10)
		}
		// types reads a row's type from its id: the kind, whatever the tens.
		typeOf := func(e []byte) vocabkey.EventType { return lifelineTestTypes[e[0]%10] }
		switch rng.Intn(4) {
		case 0:
			add(1)
			birth = id(1)
		case 1:
			add(8)
			death = id(8)
		case 2:
			add(1)
			birth = id(1)
			add(8)
			death = id(8)
		}
		// Date order: earliest start first, an open start first.
		sort.SliceStable(ns, func(i, j int) bool {
			a, b := days[ns[i]].Lo, days[ns[j]].Lo
			return a == nil && b != nil || a != nil && b != nil && *a < *b
		})
		var rows []PersonEvent
		for _, n := range ns {
			rows = append(rows, lifelineTestRow(n))
		}
		daysOf := func(e []byte) (valuecompare.Window, bool) { w, ok := days[e[0]]; return w, ok }
		got := lifeline.order(rows, typeOf, daysOf, birth, death)
		if len(got) != len(rows) {
			t.Fatalf("case %d: %d rows, want %d", c, len(got), len(rows))
		}
		at := map[byte]int{}
		for i, r := range got {
			at[r.Event[0]] = i
		}
		if len(at) != len(rows) {
			t.Fatalf("case %d: output %v is not a permutation of %v", c, got, ns)
		}
		for i := range rows {
			for j := i + 1; j < len(rows); j++ {
				a, b := rows[i].Event[0], rows[j].Event[0]
				if at[a] > at[b] && !days[a].Overlaps(days[b]) {
					var ws []string
					for _, n := range ns {
						w := days[n]
						lo, hi := "-", "-"
						if w.Lo != nil {
							lo = fmt.Sprint(*w.Lo)
						}
						if w.Hi != nil {
							hi = fmt.Sprint(*w.Hi)
						}
						ws = append(ws, fmt.Sprintf("%d:[%s,%s]", n, lo, hi))
					}
					var o []byte
					for _, r := range got {
						o = append(o, r.Event[0])
					}
					t.Fatalf("case %d: rows %d and %d changed order and do not overlap; input %v got %v", c, a, b, ws, o)
				}
			}
		}
	}
}

// TestLifelineOrderLargeBudget holds order to a time on a crafted catalog: far
// more burials and funerals than a person has, all in one window with a dated
// death. Pass (3) leaves them in date order above maxReorderedAfterDeath, so
// the cost is that of passes (1) and (2).
func TestLifelineOrderLargeBudget(t *testing.T) {
	n := 20 * 64
	days := map[byte]valuecompare.Window{}
	w, _ := valuecompare.DateWindow(*datevaluetest.YMD(2011, 0, 0))
	rows := []PersonEvent{lifelineTestRow(8)}
	days[8] = w
	for i := 0; i < n; i++ {
		kind := byte(4)
		if i%2 == 1 {
			kind = 3
		}
		// Ids are two bytes so that no row repeats; the type is the first.
		r := lifelineTestRow(kind)
		r.Event = []byte{kind, byte(i / 256), byte(i)}
		days[kind] = w
		rows = append(rows, r)
	}
	typeOf := func(e []byte) vocabkey.EventType {
		if len(e) == 1 {
			return vocabkey.EventDeath
		}
		return lifelineTestTypes[e[0]]
	}
	daysOf := func(e []byte) (valuecompare.Window, bool) {
		if len(e) == 1 {
			return days[8], true
		}
		return days[e[0]], true
	}
	d := databasetest.Fastest(3, func() time.Duration {
		return databasetest.Since(func() { lifeline.timeline(rows, typeOf, daysOf, id(8), id(8)) })
	})
	t.Logf("%d rows: %v", len(rows), d)
	databasetest.WithinBudget(t, "order on a crafted catalog", d, 2*time.Millisecond)
}
