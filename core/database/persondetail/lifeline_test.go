package persondetail

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

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
	before := func(n byte) *MarkerPlace { return &MarkerPlace{Row: id(n), Side: SideBefore} }
	after := func(n byte) *MarkerPlace { return &MarkerPlace{Row: id(n), Side: SideAfter} }
	same := func(a, b *MarkerPlace) bool {
		if a == nil || b == nil {
			return a == b
		}
		return a.Side == b.Side && bytes.Equal(a.Row, b.Row)
	}
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
