package graphalign_test

import (
	"bytes"
	"reflect"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/propertycompare/propertycomparetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// sparseAnnAmes is TestSparseNameAndBirthDate's fixture: one look-alike per
// birth year, so the right handle is the one whose links line up.
func sparseAnnAmes() (graphalign.Layer, graphalign.Canon) {
	date := vocabkey.PropertyKey(vocabkey.PropertyDate)
	year := func(y int) propertycompare.Value { return propertycomparetest.Date(datevaluetest.YMD(y, 0, 0)) }
	et, birth := termProp("event_type", "birth")
	sig := partLink("subject", "birth")
	per, ev := id("s-person"), id("s-event")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: per, Ref: "PER-S", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: ev, Ref: "EVT-S", Kind: "event", Values: vals(et, birth, date, year(1817))},
		},
		Bridges: []graphalign.Bridge{bridge(per, ev, sig)},
		Metas:   mergeMetas(nameMetas(), eventMetas()),
	}
	p1, e1, p2, e2 := id("h-person-1"), id("h-event-1"), id("h-person-2"), id("h-event-2")
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: p1, Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: e1, Ref: "EVT-1", Kind: "event", Values: vals(et, birth, date, year(1817))},
			{ID: p2, Ref: "PER-2", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: e2, Ref: "EVT-2", Kind: "event", Values: vals(et, birth, date, year(1830))},
		},
		Edges: []graphalign.CanonEdge{edge(p1, e1, sig), edge(p2, e2, sig)},
	}
	return layer, canon
}

func sortedRows(p graphalign.Proposal) []graphalign.Row {
	rows := slices.Clone(p.Rows)
	slices.SortFunc(rows, func(a, b graphalign.Row) int { return bytes.Compare(a.SubjectID, b.SubjectID) })
	return rows
}

func TestAlignInvariants(t *testing.T) {
	cfg := matchregistry.Default().AlignConfig()
	fixtures := []struct {
		name  string
		build func() (graphalign.Layer, graphalign.Canon)
		fixed []graphalign.Fixed
	}{
		{"obituary", gracieObituary, []graphalign.Fixed{{SubjectID: id("s-gracie"), HandleID: id("h-gracie")}}},
		{"sparse name and birth", sparseAnnAmes, nil},
		{"part-of chain", partOfChain, nil},
		{"marriage on both sides", func() (graphalign.Layer, graphalign.Canon) {
			return marriageFixture("marriage", "marriage", true)
		}, []graphalign.Fixed{{SubjectID: id("s-ann"), HandleID: id("h-ann")}}},
	}
	for _, fx := range fixtures {
		t.Run(fx.name, func(t *testing.T) {
			layer, canon := fx.build()
			base := alignWith(layer, canon, fx.fixed, cfg)

			// (a) input order does not matter.
			rl, rc := layer, canon
			rl.Subjects, rl.Bridges = slices.Clone(layer.Subjects), slices.Clone(layer.Bridges)
			rc.Handles, rc.Edges = slices.Clone(canon.Handles), slices.Clone(canon.Edges)
			slices.Reverse(rl.Subjects)
			slices.Reverse(rl.Bridges)
			slices.Reverse(rc.Handles)
			slices.Reverse(rc.Edges)
			if rev := alignWith(rl, rc, fx.fixed, cfg); !reflect.DeepEqual(sortedRows(rev), sortedRows(base)) {
				t.Fatalf("reversed input differs:\n%+v\n%+v", sortedRows(rev), sortedRows(base))
			}

			// (b) a second run is identical.
			if again := alignWith(layer, canon, fx.fixed, cfg); !reflect.DeepEqual(again, base) {
				t.Fatalf("second run differs:\n%+v\n%+v", again, base)
			}

			// (c) fixing a proposed handle row on its handle changes no other row.
			for _, row := range base.Rows {
				if row.Target != graphalign.TargetHandle || slices.ContainsFunc(fx.fixed, func(f graphalign.Fixed) bool { return bytes.Equal(f.SubjectID, row.SubjectID) }) {
					continue
				}
				fixed := append(slices.Clone(fx.fixed), graphalign.Fixed{SubjectID: row.SubjectID, HandleID: row.HandleID})
				after := alignWith(layer, canon, fixed, cfg)
				for _, other := range base.Rows {
					if bytes.Equal(other.SubjectID, row.SubjectID) {
						continue
					}
					got := rowBySubject(after, other.SubjectID)
					if got.Target != other.Target || !bytes.Equal(got.HandleID, other.HandleID) || got.Assessment != other.Assessment {
						t.Fatalf("fixing %s on %s moved %s: before %+v, after %+v", row.SubjectID, row.HandleID, other.SubjectID, other, got)
					}
				}
				// (d) the fixed row is what was fixed.
				if got := rowBySubject(after, row.SubjectID); got.Target != graphalign.TargetHandle || !bytes.Equal(got.HandleID, row.HandleID) {
					t.Fatalf("fixed row %s moved: %+v", row.SubjectID, got)
				}
			}
			for _, f := range fx.fixed {
				if got := rowBySubject(base, f.SubjectID); got.Target != graphalign.TargetHandle || !bytes.Equal(got.HandleID, f.HandleID) {
					t.Fatalf("fixed row %s moved: %+v", f.SubjectID, got)
				}
			}
		})
	}
}
