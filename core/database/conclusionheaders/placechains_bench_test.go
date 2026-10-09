package conclusionheaders

import (
	"fmt"
	"testing"

	"github.com/mendahu/provenencia/core/database/canonicalentities"
)

// A Places list composes every Place's chain from one graph: 1 country, 10
// provinces, 50 counties each, 20 towns per county (10,511 Places).
func BenchmarkParentChainsForEveryPlace(b *testing.B) {
	g := newPlaceGraph()
	var ids [][]byte
	n := 0
	add := func(id, parent string) {
		g.setPlace(PlaceHeader{Entity: canonicalentities.Entity{ID: []byte(id), Ref: id}, Names: []string{id}})
		ids = append(ids, []byte(id))
		if parent != "" {
			n++
			g.partOf = append(g.partOf, placeLink{from: []byte(id), to: []byte(parent), association: []byte(fmt.Sprint("a", n))})
			i := len(g.partOf) - 1
			g.partUp[id] = append(g.partUp[id], i)
			g.partDown[parent] = append(g.partDown[parent], i)
		}
	}
	add("C", "")
	for p := 0; p < 10; p++ {
		pid := fmt.Sprintf("P%d", p)
		add(pid, "C")
		for c := 0; c < 50; c++ {
			cid := fmt.Sprintf("%s-C%d", pid, c)
			add(cid, pid)
			for t := 0; t < 20; t++ {
				add(fmt.Sprintf("%s-T%d", cid, t), cid)
			}
		}
	}
	today := TodayDate()
	b.ResetTimer()
	for b.Loop() {
		g.chains = map[string]parentChainResult{}
		for _, id := range ids {
			g.ParentChain(id, &today)
		}
	}
}
