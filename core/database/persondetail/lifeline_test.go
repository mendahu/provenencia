package persondetail

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func TestLifeMarkers(t *testing.T) {
	span := &conclusionsummaries.TimelineSpan{From: 1800, To: 1801}
	id := func(n byte) []byte { return []byte{n} }
	typeOf := map[byte]vocabkey.EventType{
		1: vocabkey.EventBirth, 2: "census", 3: vocabkey.EventFuneral, 4: vocabkey.EventBurial, 5: vocabkey.EventMarriage,
	}
	types := func(e []byte) vocabkey.EventType { return typeOf[e[0]] }
	row := func(n byte, dated bool) PersonEvent {
		r := PersonEvent{Event: id(n)}
		if dated {
			r.Span = span
		}
		return r
	}
	cases := []struct {
		name                     string
		rows                     []PersonEvent
		birth, death             []byte
		birthBeside, deathBeside []byte
	}{
		{"only a burial dated", []PersonEvent{row(4, true)}, nil, nil, id(4), nil},
		{"a funeral after a census", []PersonEvent{row(2, true), row(3, true)}, nil, nil, id(2), id(2)},
		{"no dated row", []PersonEvent{row(2, false)}, nil, nil, nil, nil},
		{"a recorded birth with no span", []PersonEvent{row(1, false), row(2, true), row(5, true)}, id(1), nil, id(2), id(5)},
		{"a recorded birth with a span", []PersonEvent{row(1, true), row(2, true)}, id(1), nil, nil, id(2)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, d := lifeMarkers(tc.rows, types, tc.birth, tc.death)
			if !bytes.Equal(b.Event, tc.birth) || !bytes.Equal(d.Event, tc.death) {
				t.Errorf("events %v %v", b.Event, d.Event)
			}
			if !bytes.Equal(b.Beside, tc.birthBeside) {
				t.Errorf("birth beside %v, want %v", b.Beside, tc.birthBeside)
			}
			if !bytes.Equal(d.Beside, tc.deathBeside) {
				t.Errorf("death beside %v, want %v", d.Beside, tc.deathBeside)
			}
		})
	}
}
