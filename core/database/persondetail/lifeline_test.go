package persondetail

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
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

func TestLifeMarkers(t *testing.T) {
	span := &conclusionsummaries.TimelineSpan{From: 1800, To: 1801}
	id := func(n byte) []byte { return []byte{n} }
	typeOf := map[byte]vocabkey.EventType{
		1: vocabkey.EventBirth, 2: "census", 3: vocabkey.EventFuneral, 4: vocabkey.EventBurial, 5: vocabkey.EventMarriage,
		6: "baptism", 7: vocabkey.EventBurial,
	}
	types := func(e []byte) vocabkey.EventType { return typeOf[e[0]] }
	row := func(n byte, dated bool, roles ...vocabkey.Role) PersonEvent {
		r := PersonEvent{Event: id(n)}
		if dated {
			r.Span = span
		}
		if len(roles) == 0 {
			roles = []vocabkey.Role{vocabkey.RoleSubject}
		}
		for _, role := range roles {
			r.Roles = append(r.Roles, conclusionsummaries.Term{Key: string(role), Origin: vocabkey.OriginProvenencia})
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
	id := func(n byte) []byte { return []byte{n} }
	typeOf := map[byte]vocabkey.EventType{
		1: vocabkey.EventBirth, 2: "census", 3: vocabkey.EventFuneral, 4: vocabkey.EventBurial, 5: vocabkey.EventMarriage,
		6: "baptism", 7: vocabkey.EventBurial, 8: vocabkey.EventDeath,
	}
	types := func(e []byte) vocabkey.EventType { return typeOf[e[0]] }
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
		{"no death", lifeline, []in{{4, datevaluetest.About(2011, 0, 0), nil}, {3, datevaluetest.After(2011, 3, 29), nil}}, nil, nil, []byte{3, 4}},
		{"a burial and a funeral both before the death", lifeline, []in{{4, datevaluetest.YMD(1990, 0, 0), nil}, {3, datevaluetest.YMD(1990, 5, 0), nil}, {8, death, nil}}, nil, id(8), []byte{4, 3, 8}},
		{"a recorded death with no days", lifeline, []in{{4, datevaluetest.About(2011, 0, 0), nil}, {3, datevaluetest.After(2011, 3, 29), nil}, {8, nil, nil}}, nil, id(8), []byte{3, 4, 8}},
		{"an undated burial", lifeline, []in{{8, death, nil}, {4, nil, nil}}, nil, id(8), []byte{8, 4}},
		{"a baptism in the birth's year", lifeline, []in{{6, datevaluetest.YMD(1850, 0, 0), nil}, {1, birthDay, nil}}, id(1), nil, []byte{1, 6}},
		{"a baptism the year before", lifeline, []in{{6, datevaluetest.YMD(1849, 0, 0), nil}, {1, birthDay, nil}}, id(1), nil, []byte{6, 1}},
		{"an infant", lifeline, []in{{6, datevaluetest.YMD(1850, 0, 0), nil}, {4, datevaluetest.YMD(1850, 0, 0), nil}, {1, birthDay, nil}, {8, datevaluetest.YMD(1850, 9, 1), nil}}, id(1), id(8), []byte{1, 6, 8, 4}},
		{"a type that may come before birth", beforeBirth, []in{{2, datevaluetest.YMD(1850, 0, 0), nil}, {6, datevaluetest.YMD(1850, 0, 0), nil}, {1, birthDay, nil}}, id(1), nil, []byte{2, 1, 6}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var rows []PersonEvent
			days := map[byte]valuecompare.Window{}
			for _, r := range tc.rows {
				roles := r.roles
				if len(roles) == 0 {
					roles = []vocabkey.Role{vocabkey.RoleSubject}
				}
				pe := PersonEvent{Event: id(r.n)}
				for _, role := range roles {
					pe.Roles = append(pe.Roles, conclusionsummaries.Term{Key: string(role), Origin: vocabkey.OriginProvenencia})
				}
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
