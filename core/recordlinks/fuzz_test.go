package recordlinks_test

import (
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/matchregistry"
	pt "github.com/mendahu/provenencia/core/propertycompare/propertycomparetest"
	"github.com/mendahu/provenencia/core/recordlinks"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// FuzzLines reads a byte string as two Evidence graphs (three bytes per item:
// kind, x, y; at most sixty items) and checks that counting never panics and
// that the two graphs are interchangeable.
func FuzzLines(f *testing.F) {
	for _, seed := range [][]byte{
		{0, 2, 1, 1, 2, 1}, // one shared person with a name
		{0, 2, 1, 1, 2, 1, 2, 1, 3, 3, 1, 3, 4, 0, 0, 5, 0, 0},          // a shared birth
		{0, 2, 0, 0, 3, 0, 0, 4, 0, 1, 2, 0, 1, 3, 0, 6, 0, 1, 7, 0, 1}, // links with no entity
		{0, 2, 1, 0, 2, 1, 0, 2, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1},          // many members of one entity
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		var ms [2][]recordlinks.Member
		var es [2][]kinship.Edge
		var persons, events [2]int
		prefix := [2]byte{0xA0, 0xB0}
		ids := func(g int, n int, event bool) []byte {
			if event {
				return []byte{prefix[g] + 8 + byte(n)}
			}
			return []byte{prefix[g] + byte(n)}
		}
		entity := func(b byte) []byte {
			if b%6 == 0 {
				return nil
			}
			return ent(b % 6)
		}
		for i := 0; i+2 < len(data) && i < 3*60; i += 3 {
			k, x, y := data[i], data[i+1], data[i+2]
			g := int(k>>3) % 2
			switch k % 8 {
			case 0, 1: // a person
				if persons[g] >= 8 {
					continue
				}
				v := subjectcompare.Values{name: {pt.Name(invNames[int(y)%len(invNames)])}}
				ms[g] = append(ms[g], person(ids(g, persons[g], false), entity(x), v))
				persons[g]++
			case 2, 3: // an event
				if events[g] >= 8 {
					continue
				}
				ms[g] = append(ms[g], event(ids(g, events[g], true), entity(x)))
				events[g]++
			case 4, 5: // a participation
				if persons[g] == 0 || events[g] == 0 {
					continue
				}
				es[g] = append(es[g], part(ids(g, int(x)%persons[g], false), ids(g, int(y)%events[g], true),
					invRoles[int(x)%len(invRoles)], invEvents[int(y)%len(invEvents)], byte(i)))
			default: // a relationship
				if persons[g] < 2 {
					continue
				}
				a, c := int(x)%persons[g], int(y)%persons[g]
				if a == c {
					continue
				}
				es[g] = append(es[g], rel(ids(g, a, false), ids(g, c, false),
					[]vocabkey.BridgeTerm{"parent", "spouse"}[int(y)%2], byte(i)))
			}
		}
		ga, gb := build(ms[0], es[0]...), build(ms[1], es[1]...)
		cfg := matchregistry.Default().RecordLinksConfig()
		metas := subjectcomparetest.Metas()
		got := recordlinks.Lines(ent(1), ga, gb, metas, cfg)
		if back := recordlinks.Lines(ent(1), gb, ga, metas, cfg); !reflect.DeepEqual(got, back) {
			t.Fatalf("Lines(a, b) = %s, Lines(b, a) = %s", show(got), show(back))
		}
	})
}
